// Copyright (c) Huawei Technologies Co., Ltd. 2022-2022. All rights reserved.

// fleet服务方法
package fleet

import (
	"encoding/json"
	"fleetmanager/api/common/query"
	"fleetmanager/api/errors"
	"fleetmanager/api/model/fleet"
	"fleetmanager/api/params"
	"fleetmanager/api/service/base"
	"fleetmanager/api/service/constants"
	"fleetmanager/api/validator"
	"fleetmanager/client"
	"fleetmanager/db/dao"
	"fleetmanager/db/dbm"
	"fleetmanager/logger"
	"fleetmanager/setting"
	"fleetmanager/utils"
	"fleetmanager/workflow"
	"fleetmanager/workflow/directer"
	"fleetmanager/worknode"
	"fmt"
	"net/http"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/beego/beego/v2/client/orm"
	"github.com/beego/beego/v2/server/web/context"
	"github.com/google/uuid"
)

const (
	BandwidthSize2000      = 2000
	BandwidthSize1000      = 1000
	BandwidthSize300       = 300
	BandwidthTimes50       = 50
	BandwidthTimes500      = 500
	CreateSubnetRetryTimes = 10
)

type Service struct {
	ctx                  *context.Context
	logger               *logger.FMLogger
	createRequest        *fleet.CreateRequest
	build                *dao.Build
	fleet                *dao.Fleet
	inboundPermissions   []*dao.InboundPermission
	runtimeConfiguration *dao.RuntimeConfiguration
	attributesUpdate     *fleet.UpdateAttributesRequest
	capacityUpdate       *fleet.UpdateFleetCapacityRequest
	workflow             *dao.Workflow
}

// CheckQuota TODO(nkaptx): 当前先做全局配置校验, 不做强校验, 不保证一致性(小概率), 后续统一对接配额管理服务
func (s *Service) CheckQuota(projectId string) *errors.CodedError {
	filter := dao.Filters{
		"ProjectId":  projectId,
		"Terminated": false,
	}

	count, err := dao.GetFleetStorage().Count(filter)
	if err != nil {
		s.logger.Error("CheckQuota db error:%+v", err)
		return errors.NewError(errors.DBError)
	}

	if count >= int64(setting.FleetQuota) {
		return errors.NewError(errors.FleetExccedQuota)
	}

	return nil
}

// RspCheck
func (s *Service) ForwardRspCheck(code int, rsp []byte, err error) (int, []byte, *errors.CodedError) {
	if code < http.StatusOK || code >= http.StatusBadRequest {
		// 创建失败的场景
		if err != nil {
			return 0, nil, errors.NewError(errors.ServerInternalError)
		}
		// 其他错误场景直接透传
		return code, rsp, nil
	}
	return code, rsp, nil
}

// NewFleetService 新建Fleet管理服务
func NewFleetService(ctx *context.Context, logger *logger.FMLogger) *Service {
	s := &Service{
		ctx:    ctx,
		logger: logger,
	}
	return s
}

func (s *Service) checkRegion() *errors.CodedError {
	supportRegions := setting.SupportRegions
	regionSlice := strings.Split(supportRegions, ",")
	for _, r := range regionSlice {
		if s.createRequest.Region == r {
			return nil
		}
	}

	s.logger.Error("region %v not in %v", s.createRequest.Region, regionSlice)
	return errors.NewError(errors.InvalidRegion)
}

func (s *Service) checkRuntimeConfiguration() *errors.CodedError {
	pc := s.createRequest.RuntimeConfiguration.ProcessConfigurations
	totalProcessNum := 0
	for _, c := range pc {
		totalProcessNum += c.ConcurrentExecutions
	}

	maxProcessNum := setting.DefaultMaxProcessNumPerFleet
	if totalProcessNum > maxProcessNum {
		s.logger.Error("total concurrent executions %v exceed limit %v", totalProcessNum, maxProcessNum)
		return errors.NewError(errors.ProcessNumExceedMaxSize)
	}

	return nil
}

/*
*
bandwidth有效性检查：https://support.huaweicloud.com/api-eip/eip_apiBandwidth_0003.html
小于等于300Mbit/s：默认最小单位为1Mbit/s。
300Mbit/s~1000Mbit/s：默认最小单位为50Mbit/s。
大于1000Mbit/s：默认最小单位为500Mbit/s。
*/
func (s *Service) checkBandwidth() *errors.CodedError {
	bandwidth := s.createRequest.Bandwidth
	if bandwidth < 0 {
		s.logger.Error("bandwidth %v in create fleet not in 1-200000", s.createRequest.Bandwidth)
		return errors.NewError(errors.InvalidBandwidth)
	}

	if bandwidth > BandwidthSize300 && bandwidth%BandwidthTimes50 != 0 {
		s.logger.Error("bandwidth %v in create fleet can not be divided by 50 in 300-1000",
			s.createRequest.Bandwidth)
		return errors.NewError(errors.InvalidBandwidth)
	}

	if bandwidth > BandwidthSize1000 && bandwidth%BandwidthTimes500 != 0 {
		s.logger.Error("bandwidth %v in create fleet can not be divided by 500 in 1000-2000",
			s.createRequest.Bandwidth)
		return errors.NewError(errors.InvalidBandwidth)
	}

	return nil
}

func (s *Service) checkBuildState() *errors.CodedError {
	b, err := dao.GetBuildById(s.createRequest.BuildId, s.ctx.Input.Param(params.ProjectId))
	if err != nil {
		s.logger.Error("build %v in create fleet error: %v", s.createRequest.BuildId, err)
		if err == orm.ErrNoRows {
			return errors.NewError(errors.BuildNotExists)
		}
		return errors.NewError(errors.DBError)
	}

	if b.State != constants.BuildStateReady {
		s.logger.Error("build %v not available in create fleet", s.createRequest.BuildId)
		return errors.NewError(errors.BuildIsNotAvailable)
	}

	s.build = b
	return nil
}

func (s *Service) buildFleet() error {
	u, _ := uuid.NewUUID()
	fleetId := u.String()
	tagsStr := "[]"
	if len(s.createRequest.InstanceTags) > 0 {
		tagsByte, err := json.Marshal(s.createRequest.InstanceTags)
		if err != nil {
			return err
		}
		tagsStr = string(tagsByte)
	}

	fd := &dao.Fleet{
		Id:                                      fleetId,
		ProjectId:                               s.ctx.Input.Param(params.ProjectId),
		Name:                                    s.createRequest.Name,
		Description:                             s.createRequest.Description,
		BuildId:                                 s.createRequest.BuildId,
		Region:                                  s.createRequest.Region,
		Bandwidth:                               s.createRequest.Bandwidth,
		InstanceSpecification:                   s.createRequest.InstanceSpecification,
		ServerSessionProtectionPolicy:           s.createRequest.ServerSessionProtectionPolicy,
		ServerSessionProtectionTimeLimitMinutes: s.createRequest.ServerSessionProtectionTimeLimitMinutes,
		EnableAutoScaling:                       false,
		ScalingIntervalMinutes:                  setting.DefaultScalingInterval,
		InstanceType:                            constants.InstanceTypeVm,
		InstanceTags:                            tagsStr,
		Minimum:                                 setting.DefaultGroupMinSize,
		Maximum:                                 setting.DefaultGroupMaxSize,
		Desired:                                 setting.DefaultGroupDesiredSize,
		OperatingSystem:                         s.build.OperatingSystem,
		State:                                   dao.FleetStateCreating,
		CreationTime:                            time.Now().UTC(),
		UpdateTime:                              time.Now().UTC(),
		PolicyPeriodInMinutes:                   s.createRequest.ResourceCreationLimitPolicy.PolicyPeriodInMinutes,
		NewSessionsPerCreator:                   s.createRequest.ResourceCreationLimitPolicy.NewSessionsPerCreator,
		EnterpriseProjectId:                     s.createRequest.EnterpriseProjectId,
	}
	s.fleet = fd
	return nil
}

func (s *Service) buildInboundPermissions() {
	s.inboundPermissions = make([]*dao.InboundPermission, 0)
	for _, p := range s.createRequest.InboundPermissions {
		pd := &dao.InboundPermission{
			FleetId:  s.fleet.Id,
			Protocol: p.Protocol,
			IpRange:  p.IpRange,
			FromPort: p.FromPort,
			ToPort:   p.ToPort,
		}
		s.inboundPermissions = append(s.inboundPermissions, pd)
	}
}

func (s *Service) buildRuntimeConfiguration() {
	ri, _ := uuid.NewUUID()
	s.runtimeConfiguration = &dao.RuntimeConfiguration{
		Id:                    ri.String(),
		FleetId:               s.fleet.Id,
		ProcessConfigurations: utils.ToJson(&s.createRequest.RuntimeConfiguration.ProcessConfigurations),
		ServerSessionActivationTimeoutSeconds: s.createRequest.RuntimeConfiguration.
			ServerSessionActivationTimeoutSeconds,
		MaxConcurrentServerSessionsPerProcess: s.createRequest.RuntimeConfiguration.
			MaxConcurrentServerSessionsPerProcess,
	}
}

func (s *Service) updateStateError() *errors.CodedError {
	f := &dao.Fleet{
		Id:    s.fleet.Id,
		State: dao.FleetStateError,
	}
	if err := dao.GetFleetStorage().Update(f, "State"); err != nil {
		s.logger.Error("update fleet state to error failed: %v", err)
		return errors.NewError(errors.DBError)
	}
	return nil
}

func (s *Service) insertDb() *errors.CodedError {
	err := s.buildFleet()
	if err != nil {
		s.logger.Error("build fleet err: %v", err)
		return errors.NewError(errors.ServerInternalError)
	}
	s.buildInboundPermissions()
	s.buildRuntimeConfiguration()
	// TODO(wj): 两次db插入做成原子操作
	if err := dao.GetFleetStorage().Insert(s.fleet); err != nil {
		s.logger.Error("insert fleet error: %v", err)
		return errors.NewError(errors.DBError)
	}
	if err := dao.GetRuntimeConfigurationStorage().Insert(s.runtimeConfiguration); err != nil {
		s.logger.Error("insert runtime config to db error: %v", err)
		return errors.NewError(errors.DBError)
	}

	return nil
}

func (s *Service) startDeleteWorkflow(f *dao.Fleet) *errors.CodedError {
	parameter := map[string]interface{}{
		"fleet":                         f,
		directer.WfKeySubnetName:        f.Id,
		directer.WfKeySecurityGroupName: f.Id,
		directer.WfKeyOriginProjectId:   f.ProjectId,
		directer.WfKeyRegion:            f.Region,
		directer.WfKeyScalingGroupName:  s.fleet.Id,
		directer.WfKeyRequestId:         fmt.Sprintf("%s", s.ctx.Input.GetData(logger.RequestId)),
	}
	wf, err := workflow.CreateWorkflow(
		"./conf/workflow/delete_fleet_workflow.json",
		parameter,
		f.Id,
		f.ProjectId,
		s.logger,
		worknode.WorkNodeId)
	if err != nil {
		s.logger.Error("create workflow in delete fleet error: %v", err)
		return errors.NewError(errors.ServerInternalError)
	}
	wf.Run()

	return nil
}

func (s *Service) startCreateWorkflow() *errors.CodedError {
	fvs, errCidr := s.GenerateVPCAndSubnetCidrForCreateFleet(s.createRequest.VPCId, s.createRequest.SubnetId, s.fleet.Id)
	if errCidr != nil {
		return errCidr
	}
	EnterpriseProject := setting.EnterpriseProject
	if s.createRequest.EnterpriseProjectId != "" {
		EnterpriseProject = s.createRequest.EnterpriseProjectId
	}

	parameter := map[string]interface{}{
		"fleet":                         s.fleet,
		"build":                         s.build,
		"runtime_configuration":         s.runtimeConfiguration,
		"inbound_permissions":           s.inboundPermissions,
		directer.WfKeyRegion:            s.fleet.Region,
		directer.WfKeyVpcName:           fvs.VpcName,
		directer.WfKeyVpcCidr:           fvs.VpcCidr,
		directer.WfKeyVpcId:             fvs.VpcId,
		directer.WfKeySecurityGroupName: s.fleet.Id,
		directer.WfKeySubnetName:        fvs.SubnetName,
		directer.WfKeySubnetCidr:        fvs.SubnetCidr,
		directer.WfKeySubnetId:          fvs.SubnetId,
		directer.WfKeySubnetGatewayIp:   fvs.GatewayIp,
		directer.WfKeyOriginProjectId:   s.fleet.ProjectId,
		directer.WfKeyScalingGroupName:  s.fleet.Id,
		directer.WfKeyInstanceTags:      s.fleet.InstanceTags,
		directer.WfKeyEipType: setting.Config.Get(
			fmt.Sprintf("%s.%s", setting.EipType, s.fleet.Region)).ToString(setting.DefaultEipType),
		directer.WfDnsConfig: setting.Config.Get(
			fmt.Sprintf("%s.%s", setting.DnsConfig, s.fleet.Region)).ToString(setting.DefaultDnsConfig),
		directer.WfKeyEnterpriseProjectId: EnterpriseProject,
		directer.WfKeyRetryTimes:          CreateSubnetRetryTimes,
		directer.WfKeyRequestId:           fmt.Sprintf("%s", s.ctx.Input.GetData(logger.RequestId)),
	}
	wf, err := workflow.CreateWorkflow(
		"./conf/workflow/create_fleet_workflow.json",
		parameter,
		s.fleet.Id,
		s.fleet.ProjectId,
		s.logger,
		worknode.WorkNodeId)
	if err != nil {
		s.logger.Error("create workflow in create fleet error: %v", err)
		return errors.NewError(errors.ServerInternalError)
	}

	wf.Run()
	return nil
}

func (s *Service) updateAttributeToAASS() (code int, rsp []byte, e *errors.CodedError) {
	group, err := dao.GetScalingGroupStorage().GetOne(dao.Filters{"FleetId": s.fleet.Id})
	if err != nil {
		s.logger.Error("get scaling group db error: %v", err)
		return 0, nil, errors.NewError(errors.DBError)
	}

	var ic *fleet.UpdateInstanceConfiguration
	if s.attributesUpdate.ServerSessionProtectionPolicy != nil ||
		s.attributesUpdate.ServerSessionProtectionTimeLimitMinutes != nil {
		ic = &fleet.UpdateInstanceConfiguration{
			ServerSessionProtectionPolicy:           s.attributesUpdate.ServerSessionProtectionPolicy,
			ServerSessionProtectionTimeLimitMinutes: s.attributesUpdate.ServerSessionProtectionTimeLimitMinutes,
		}
	}

	updateReq := fleet.UpdateScalingGroupRequest{
		Id:                    &group.Id,
		InstanceConfiguration: ic,
		EnableAutoScaling:     s.attributesUpdate.EnableAutoScaling,
		CoolDownTime:          s.attributesUpdate.ScalingIntervalMinutes,
		InstanceTags:          s.attributesUpdate.InstanceTags,
	}
	code, rsp, err = forwardUpdateScalingGroupRequest(s.ctx, s.fleet.Region, group.ResourceProjectId, updateReq)
	s.logger.Info("update scaling group to aass, code: %d, rsp: %s, error: %+v", code, rsp, err)
	return s.ForwardRspCheck(code, rsp, err)
}

func (s *Service) updateAttributeToDb() *errors.CodedError {
	UpdateAttributeDao, err := BuildUpdateAttributeDao(s.logger, *s.attributesUpdate)
	if err != nil {
		return errors.NewError(errors.InvalidParameterValue)
	}
	attrUpdateJson, err := json.Marshal(UpdateAttributeDao)
	if err != nil {
		s.logger.Error("update fleet attribute marshal error: %+v", err)
		return errors.NewError(errors.InvalidParameterValue)
	}

	err = json.Unmarshal(attrUpdateJson, s.fleet)
	if err != nil {
		s.logger.Error("update fleet attribute make req error: %+v", err)
		return errors.NewError(errors.InvalidParameterValue)
	}

	// 设置NewSessionsPerCreator和PolicyPeriodInMinutes参数
	if s.attributesUpdate.ResourceCreationLimitPolicy != nil {
		if s.attributesUpdate.ResourceCreationLimitPolicy.NewSessionsPerCreator != nil {
			s.fleet.NewSessionsPerCreator = *s.attributesUpdate.ResourceCreationLimitPolicy.NewSessionsPerCreator
		}
		if s.attributesUpdate.ResourceCreationLimitPolicy.PolicyPeriodInMinutes != nil {
			s.fleet.PolicyPeriodInMinutes = *s.attributesUpdate.ResourceCreationLimitPolicy.PolicyPeriodInMinutes
		}
	}

	err = dao.GetFleetStorage().Update(s.fleet, "Name", "Description", "ServerSessionProtectionPolicy",
		"ServerSessionProtectionTimeLimitMinutes", "EnableAutoScaling", "ScalingIntervalMinutes",
		"PolicyPeriodInMinutes", "NewSessionsPerCreator", "InstanceTags")
	if err != nil {
		s.logger.Error("update fleet attribute db error: %v", err)
		return errors.NewError(errors.DBError)
	}

	return nil
}

// Create 创建Fleet
func (s *Service) Create(r *fleet.CreateRequest) (*fleet.CreateResponse, *errors.CodedError) {
	s.createRequest = r
	if err := s.checkRegion(); err != nil {
		return nil, err
	}

	if err := s.checkBandwidth(); err != nil {
		return nil, err
	}

	if err := s.checkRuntimeConfiguration(); err != nil {
		return nil, err
	}

	if err := s.checkBuildState(); err != nil {
		return nil, err
	}

	if err := s.CheckQuota(s.ctx.Input.Param(params.ProjectId)); err != nil {
		return nil, err
	}

	if err := s.insertDb(); err != nil {
		return nil, err
	}

	if err := s.startCreateWorkflow(); err != nil {
		// 更新db, fleet状态刷为error
		if e := s.updateStateError(); e != nil {
			s.logger.Error("update fleet state to error failed: %v", e)
		}

		return nil, err
	}

	rsp := &fleet.CreateResponse{
		Fleet: BuildFleetModel(s.fleet),
	}

	return rsp, nil
}

func (s *Service) checkOffsetLimit(count int64, offset int, limit int) *errors.CodedError {
	if int64(offset*limit) >= count {
		return errors.NewErrorF(errors.InvalidParameterValue, " offset and limit over total count")
	}
	return nil
}

// List 查询fleet列表
func (s *Service) List(offset int, limit int) (fleet.List, *errors.CodedError) {
	var list fleet.List
	projectId := s.ctx.Input.Param(params.ProjectId)
	queryState := s.ctx.Input.Query(params.QueryState)
	queryId := s.ctx.Input.Query(params.QueryFleetId)
	queryName := s.ctx.Input.Query(params.QueryFleetName)

	offset, err := query.CheckOffset(s.ctx)
	if err != nil {
		s.logger.Error("invalid offset: %v", err)
		return list, errors.NewError(errors.InvalidParameterValue)
	}
	limit, err = query.CheckLimit(s.ctx)
	if err != nil {
		s.logger.Error("invalid limit: %v", err)
		return list, errors.NewError(errors.InvalidParameterValue)
	}
	if queryState != "" && queryState != dao.FleetStateCreating && queryState != dao.FleetStateActive &&
		queryState != dao.FleetStateError && queryState != dao.FleetStateDeleting {
		s.logger.Error("invalid status: %s", queryState)
		return list, errors.NewError(errors.InvalidParameterValue)
	}

	fleets, count, err := dao.QueryFleetByCondition(projectId, offset, limit, queryId, queryName, queryState)
	if err != nil {
		s.logger.Error("list fleet from db error: %v", err)
		return list, errors.NewError(errors.DBError)
	}
	list.Count = count
	for _, f := range fleets {
		f := BuildFleetModel(&f)
		list.Fleets = append(list.Fleets, f)
	}

	return list, nil
}

// Show 查询fleet详情
func (s *Service) Show() (*fleet.ShowFleetResponse, *errors.CodedError) {
	fleetId := s.ctx.Input.Param(params.FleetId)
	filter := dao.Filters{"Id": fleetId}
	fd, err := dao.GetFleetStorage().Get(filter)
	if err != nil {
		s.logger.Error("show fleet db error: %v", err)
		if err == orm.ErrNoRows {
			return nil, errors.NewError(errors.FleetNotFound)
		}
		return nil, errors.NewError(errors.DBError)
	}

	rsp := &fleet.ShowFleetResponse{
		Fleet: BuildFleetModel(fd),
	}
	return rsp, nil
}

// 获取fleet的实时运行信息
func (s *Service) ShowMonitorFleet() (*fleet.ShowMonitorFleetResponse, *errors.CodedError) {
	fleet_id := s.ctx.Input.Param(params.FleetId)
	s.logger.Info("receive a minitor fleet query requests: %s", fleet_id)
	ft, err := s.Show()
	if err != nil {
		return nil, err
	}

	return s.GetMonitorFleet(&ft.Fleet)
}

// list monitor fleets
func (s *Service) ListMonitorFleets(offset int, limit int) (*fleet.ListMonitorFleetsResponse,
	*errors.CodedError) {
	s.logger.Info("receive a minitor fleet list requests")
	fts, err1 := s.List(offset, limit)
	if err1 != nil {
		s.logger.WithField(logger.Error, err1.Error()).Error("query fleet list from db error")
		return nil, err1
	}
	allFleets, err2 := s.List(params.DefaultNumber, int(fts.Count))
	if err2 != nil {
		s.logger.WithField(logger.Error, err2.Error()).Error("query fleet list from db error")
		return nil, err2
	}
	mfts := &fleet.ListMonitorFleetsResponse{
		TotalCount:          int(fts.Count),
		Count:               len(fts.Fleets),
		Fleets:              []fleet.ShowMonitorFleetResponse{},
		AllFleetIdsAndNames: *s.GenarateFleetNamesAndIds(&allFleets),
	}
	wg := sync.WaitGroup{}

	// 异步获取每个fleet的信息
	wg.Add(int(mfts.Count))
	for _, ft := range fts.Fleets {
		go func(oneFt fleet.Fleet) {
			defer wg.Done()
			mftp, err := s.GetMonitorFleet(&oneFt)
			if err != nil {
				errftp := &fleet.ShowMonitorFleetResponse{
					Fleet: oneFt,
				}
				mfts.Fleets = append(mfts.Fleets, *errftp)
				s.logger.Info("get fleet: %s info error: %+v", oneFt.FleetId, err)
			} else {
				mfts.Fleets = append(mfts.Fleets, *mftp)
			}
		}(ft)
	}
	wg.Wait()
	sortFleet := ShowMonitorFleet{}
	sortFleet = mfts.Fleets
	sort.Sort(sortFleet)
	mfts.TotalCount = mfts.TotalCount - (mfts.Count - len(mfts.Fleets))
	mfts.Count = len(mfts.Fleets)
	return mfts, nil
}

func (s *Service) GenarateFleetNamesAndIds(fts *fleet.List) *[]fleet.FleetIdAndName {
	IdsAndNames := &[]fleet.FleetIdAndName{}
	for _, ft := range fts.Fleets {
		IdName := &fleet.FleetIdAndName{
			FleetId:   ft.FleetId,
			FleetName: ft.Name,
			State:     ft.State,
		}
		*IdsAndNames = append(*IdsAndNames, *IdName)
	}
	return IdsAndNames
}

func (s *Service) ListMonitorInstances() (*fleet.ListMonitorInstancesResponce,
	*errors.CodedError) {
	s.logger.Info("receive a minitor instances list requests")
	if err := s.setFleet(); err != nil {
		return nil, errors.NewError(errors.FleetNotFound)
	}
	query_params := base.GetQueryParams(s.ctx, base.GetInstancesQueryField())
	duration_start, duration_end, err := base.GetDuration(query_params, s.logger)
	if err != nil {
		return nil, errors.NewErrorF(errors.InvalidParameterValue, err.Error())
	}
	if duration_start > duration_end {
		return nil, errors.NewErrorF(errors.InvalidParameterValue, "duration start largger than duration end")
	}
	code, rsp, NewErr := base.ForwardToAppgw(s.ctx, s.fleet.Region,
		constants.APPGWMonitorInstancesUrl, query_params)
	s.logger.Info("list instances from appgateway, code: %d, rsp: %s, error: %+v", code, rsp, NewErr)
	resp_appg := &fleet.ListInstancesFromAppGW{}
	if NewErrC := base.AcceptResp(s.logger, resp_appg, code, rsp, NewErr); NewErrC != nil {
		return nil, NewErrC
	}
	appg_instance, ips := GenerateApAndSessionCount(&resp_appg.Instances)
	group, err := dao.GetScalingGroupStorage().GetOne(dao.Filters{"FleetId": s.fleet.Id})
	if err != nil {
		s.logger.Info("Get scaling group by fleet id %s error: %+v", s.fleet.Id, err)
	}
	code, rsp, err = base.ForwardToAASS(s.ctx, s.fleet.Region,
		fmt.Sprintf(constants.AASSMonitorInstancesUrlPattern,
			group.ResourceProjectId), query_params)
	s.logger.Info("list instances from aass, code: %d, "+
		"rsp: %s, error: %+v", code, rsp, err)
	resp_aass := &fleet.ListInstanceResonseFromAASS{}
	if NewErrC := base.AcceptResp(s.logger, resp_aass, code, rsp, err); NewErrC != nil {
		return nil, NewErrC
	}
	return base.GenerateInstances(query_params, resp_aass, appg_instance, ips, s.logger)
}

func (s *Service) GetMonitorFleet(ft *fleet.Fleet) (*fleet.ShowMonitorFleetResponse,
	*errors.CodedError) {

	query_params := make(map[string]string)
	query_params[params.QueryFleetId] = ft.FleetId

	// 获取fleet已绑定的别名列表
	var alisaList []fleet.AssociatedAliasRsp
	var alisaModel []dao.Alias
	_, err := dbm.Ormer.QueryTable(dao.AliasTable).Filter("AssociatedFleets__contains", ft.FleetId).
		Filter("Type__in", dao.AliasTypeActive, dao.AliasTypeDeactive).All(&alisaModel)
	if err != nil {
		s.logger.Info("Get associated alisa by fleet id %s error: %+v", ft.FleetId, err)
		return nil, errors.NewError("Get associated alisa by fleet id from DB error")
	}
	alisaList = buildAssociateAlisa(alisaModel)

	// 从appgateway中获取每个fleet的进程数量与会话数量
	code, rsp, NewErr := base.ForwardToAppgw(s.ctx, ft.Region,
		constants.APPGWMonitorFleetsUrl, query_params)
	s.logger.Info("get fleet from appgateway, code: %d, rsp: %s, error: %+v", code, rsp, NewErr)
	resp_appg := &fleet.FleetResponseFromAPPGW{}
	if NewErrC := base.AcceptResp(s.logger, resp_appg, code, rsp, NewErr); NewErrC != nil {
		s.logger.Error("accept rsp: %s from app gateway error, error: %s,", rsp, NewErrC)
		return nil, NewErrC
	}
	// 从aass中获取每个fleet对应的弹性伸缩组的进程数量信息
	group, err := dao.GetScalingGroupStorage().GetOne(dao.Filters{"FleetId": ft.FleetId})
	if err != nil {
		s.logger.Info("Get scaling group by fleet id %s error: %+v", ft.FleetId, err)
		return s.GenerateMonitorFleetResponse(ft, resp_appg, 0, alisaList), nil
	}
	code, rsp, NewErr = base.ForwardToAASS(s.ctx, ft.Region,
		fmt.Sprintf(constants.AASSMonitorInstancesUrlPattern,
			group.ResourceProjectId), query_params)
	s.logger.Info("get instances from aass, code: %d, rsp: %s, error: %+v", code, rsp, NewErr)
	rsp_aass := &fleet.ListInstanceResonseFromAASS{
		Instances: []fleet.InstanceResponseFromAASS{},
	}
	if NewErrC := base.AcceptResp(s.logger, rsp_aass, code, rsp, NewErr); NewErrC != nil {
		s.logger.Error("accept rsp: %s from aass error, error: %s,", rsp, NewErrC)
		return nil, NewErrC
	}
	resp := s.GenerateMonitorFleetResponse(ft, resp_appg, rsp_aass.TotalNumber, alisaList)
	return resp, nil
}

func (s *Service) ListMonitorAppProcesses() (*fleet.ListMonitorAppProcessResponseFromAppGW, *errors.CodedError) {
	s.logger.Info("receive a minitor app process list requests")
	if err := s.setFleet(); err != nil {
		return nil, errors.NewError(errors.ServerInternalError)
	}
	query_params := base.GetQueryParams(s.ctx, base.GetAppProcessesQueryField())
	code, rsp, NewErr := base.ForwardToAppgw(s.ctx, s.fleet.Region, constants.APPGWMonitorAppProcessesUrl,
		query_params)
	s.logger.Info("list app processes from appgateway, code: %d, rsp: %s, error: %+v", code, rsp, NewErr)
	resp_appg := &fleet.ListMonitorAppProcessResponseFromAppGW{
		Processes: []fleet.MonitorProcessFromApGW{},
	}
	if NewErrC := base.AcceptResp(s.logger, resp_appg, code, rsp, NewErr); NewErrC != nil {
		return nil, NewErrC
	}
	return resp_appg, nil
}

func (s *Service) ListMonitorServerSessions() (*fleet.ListMonitorServerSessionResponseFromAppGW, *errors.CodedError) {
	s.logger.Info("receive a minitor server session list requests")
	if err := s.setFleet(); err != nil {
		return nil, errors.NewError(errors.ServerInternalError)
	}
	query_params := base.GetQueryParams(s.ctx, base.GetServerSessionsQueryFiled())
	code, rsp, NewErr := base.ForwardToAppgw(s.ctx, s.fleet.Region, constants.APPGWMonitorServerSessionsUrl,
		query_params)
	s.logger.Info("list server sessions from appgateway, code: %d, rsp: %s, error: %+v", code, rsp, NewErr)
	resp_appg := &fleet.ListMonitorServerSessionResponseFromAppGW{
		ServerSessions: []fleet.MonitorServerSessionFromAppGW{},
	}
	if NewErrC := base.AcceptResp(s.logger, resp_appg, code, rsp, NewErr); NewErrC != nil {
		return nil, NewErrC
	}
	return resp_appg, nil
}

func (s *Service) GenerateMonitorFleetResponse(ft *fleet.Fleet, fleet_appg *fleet.FleetResponseFromAPPGW,
	instanceCount int, alisa []fleet.AssociatedAliasRsp) *fleet.ShowMonitorFleetResponse {
	return &fleet.ShowMonitorFleetResponse{
		InstanceCount:       instanceCount,
		ProcessCount:        fleet_appg.ProcessCount,
		ServerSessionCount:  fleet_appg.ServerSessionCount,
		MaxServerSessionNum: fleet_appg.MaxServerSessionNum,
		Fleet:               *ft,
		AliasCount:          len(alisa),
		AssociatedAlias:     alisa,
	}
}

func (s *Service) setFleet() error {
	fleetId := s.ctx.Input.Param(params.FleetId)
	projectId := s.ctx.Input.Param(params.ProjectId)
	filter := dao.Filters{
		"Id":         fleetId,
		"ProjectId":  projectId,
		"Terminated": false,
	}
	f, err := dao.GetFleetStorage().Get(filter)
	if err != nil {
		return err
	}

	s.fleet = f
	return nil
}

func (s *Service) buildUpdateAttributeReq() *errors.CodedError {
	req := &fleet.UpdateAttributesRequest{}
	if err := json.Unmarshal(s.ctx.Input.RequestBody, req); err != nil {
		s.logger.Error("unmarshal request body %v error: %v", s.ctx.Input.RequestBody, err)
		return errors.NewErrorF(errors.InvalidParameterValue, " read request params error")
	}

	if err := validator.Validate(req); err != nil {
		s.logger.Error("request params invalid, reqBody:%s, err:%+v", s.ctx.Input.RequestBody, err)
		return errors.NewErrorF(errors.InvalidParameterValue, err.Error())
	}
	if req.InstanceTags == nil {
		instanceTags := &[]fleet.InstanceTag{}
		err := json.Unmarshal([]byte(s.fleet.InstanceTags), instanceTags)
		if err != nil {
			return errors.NewError(errors.ServerInternalError)
		}
		req.InstanceTags = instanceTags
	}
	s.attributesUpdate = req
	return nil
}

// UpdateAttribute 更新fleet属性
func (s *Service) UpdateAttribute() (code int, rsp []byte, e *errors.CodedError) {
	if err := s.setFleet(); err != nil {
		s.logger.Error("update fleet attr db error: %v", err)
		if err == orm.ErrNoRows {
			return 0, nil, errors.NewError(errors.FleetNotFound)
		} else {
			return 0, nil, errors.NewError(errors.DBError)
		}
	}

	if s.fleet.State != dao.FleetStateActive {
		s.logger.Error("fleet %v do not support update", s.fleet)
		return 0, nil, errors.NewError(errors.FleetStateNotSupportUpdate)
	}

	if e := s.buildUpdateAttributeReq(); e != nil {
		return 0, nil, e
	}

	// TODO(wj): 保持原子性
	code, rsp, e = s.updateAttributeToAASS()
	// 如果调用接口失败了, 直接返回
	if (code < http.StatusOK || code >= http.StatusBadRequest) || e != nil {
		return
	}

	if e = s.updateAttributeToDb(); e != nil {
		return
	}
	return
}

// GetInstanceCapacity 获取fleet实例容量
func (s *Service) GetInstanceCapacity() (rsp fleet.ShowInstanceCapacityRsp, error *errors.CodedError) {
	fleetId := s.ctx.Input.Param(params.FleetId)
	filter := dao.Filters{"Id": fleetId}
	f, err := dao.GetFleetStorage().Get(filter)
	if err != nil {
		s.logger.Error("get instance capacity from db error: %v", err)
		if err == orm.ErrNoRows {
			return rsp, errors.NewError(errors.FleetNotFound)
		}
		return rsp, errors.NewError(errors.DBError)
	}

	rsp = fleet.ShowInstanceCapacityRsp{
		FleetId: f.Id,
		InstanceCapacity: fleet.InstanceCapacity{
			Minimum: f.Minimum,
			Maximum: f.Maximum,
			Desired: f.Desired,
		},
	}

	return rsp, nil
}

func (s *Service) updateCapacityToAASS() (code int, rsp []byte, e *errors.CodedError) {
	group, err := dao.GetScalingGroupStorage().GetOne(dao.Filters{"FleetId": s.fleet.Id})
	if err != nil {
		s.logger.Error("get scaling group db error: %v", err)
		return 0, nil, errors.NewError(errors.DBError)
	}

	updateReq := fleet.UpdateScalingGroupRequest{
		Id:                   &group.Id,
		MinInstanceNumber:    s.capacityUpdate.Minimum,
		MaxInstanceNumber:    s.capacityUpdate.Maximum,
		DesireInstanceNumber: s.capacityUpdate.Desired,
	}
	code, rsp, err = forwardUpdateScalingGroupRequest(s.ctx, s.fleet.Region, group.ResourceProjectId, updateReq)
	s.logger.Info("update scaling group to aass, code: %d, rsp: %s, error: %+v", code, rsp, err)
	return s.ForwardRspCheck(code, rsp, err)
}

func (s *Service) updateCapacityToDb() *errors.CodedError {
	s.fleet.Minimum = *s.capacityUpdate.Minimum
	s.fleet.Maximum = *s.capacityUpdate.Maximum
	s.fleet.Desired = *s.capacityUpdate.Desired

	err := dao.GetFleetStorage().Update(s.fleet, "Minimum", "Maximum", "Desired")
	if err != nil {
		s.logger.Error("update fleet capacity db error: %v", err)
		return errors.NewError(errors.DBError)
	}

	return nil
}

func (s *Service) buildUpdateInstanceCapacityReq() *errors.CodedError {
	capacity := &fleet.UpdateFleetCapacityRequest{}
	if err := json.Unmarshal(s.ctx.Input.RequestBody, capacity); err != nil {
		s.logger.Error("unmarshal request body error: %v", err)
		return errors.NewErrorF(errors.InvalidParameterValue, " read request params error")
	}

	if err := validator.Validate(capacity); err != nil {
		s.logger.Error("request params invalid: %v", err)
		return errors.NewErrorF(errors.InvalidParameterValue, " parameter invalid %v", err)
	}

	s.capacityUpdate = capacity
	return nil
}

// UpdateInstanceCapacity 更新实例容量
func (s *Service) UpdateInstanceCapacity() (code int, rsp []byte, e *errors.CodedError) {
	if e := s.setFleet(); e != nil {
		if e == orm.ErrNoRows {
			return 0, nil, errors.NewErrorF(errors.InvalidParameterValue, " fleet not found")
		} else {
			s.logger.Error("find fleet info db error: %v", e)
			return 0, nil, errors.NewError(errors.DBError)
		}
	}

	if s.fleet.State != dao.FleetStateActive {
		s.logger.Error("fleet %v do not support update", s.fleet)
		return 0, nil, errors.NewError(errors.FleetStateNotSupportUpdate)
	}

	if e := s.buildUpdateInstanceCapacityReq(); e != nil {
		return 0, nil, e
	}

	code, rsp, e = s.updateCapacityToAASS()
	// 如果调用接口失败了
	if (code < http.StatusOK || code >= http.StatusBadRequest) || e != nil {
		return
	}

	if e = s.updateCapacityToDb(); e != nil {
		return
	}
	return
}

// Delete 删除fleet
func (s *Service) Delete() *errors.CodedError {
	if e := s.setFleet(); e != nil {
		if e == orm.ErrNoRows {
			return errors.NewErrorF(errors.FleetNotFound, " fleet not found")
		} else {
			s.logger.Error("get fleet info db error: %v", e)
			return errors.NewError(errors.DBError)
		}
	}
	isAssociated, err := DeleteFleetCheckAssociatedFromAlias(s.fleet.Id, s.fleet.ProjectId)
	if err != nil {
		s.logger.Error("check associated fleet on alias error: %v", err)
		return errors.NewError(errors.DBError)
	}

	if isAssociated {
		s.logger.Error("fleet %s used by alias, not allow to delete", s.fleet.Id)
		return errors.NewError(errors.FleetUsedByAlias)
	}

	if s.fleet.State == dao.FleetStateTerminated || s.fleet.State == dao.FleetStateDeleting {
		return nil
	}

	if s.fleet.State == dao.FleetStateCreating {
		s.logger.Error("fleet %v do not support delete", s.fleet)
		return errors.NewError(errors.FleetStateNotSupportDelete)
	}

	s.fleet.State = dao.FleetStateDeleting
	if err := dao.GetFleetStorage().Update(s.fleet, "State"); err != nil {
		s.logger.Error("update fleet deleting in db error: %v", err)
		return errors.NewError(errors.ServerInternalError)
	}

	if err := s.startDeleteWorkflow(s.fleet); err != nil {
		return err
	}

	return nil
}

func (s *Service) forwardShowProcessCountsToAPPGW() (code int, rsp []byte, err error) {
	url := client.GetServiceEndpoint(client.ServiceNameAPPGW, s.fleet.Region) + constants.ProcessCountsUrl
	req := client.NewRequest(client.ServiceNameAPPGW, url, http.MethodGet, nil)
	req.SetQuery(params.QueryFleetId, s.ctx.Input.Param(params.FleetId))
	req.SetHeader(map[string]string{
		logger.RequestId: fmt.Sprintf("%s", s.ctx.Input.GetData(logger.RequestId)),
	})
	return req.DoRequest()
}

// ShowProcessCounts 查询应用进程数量
func (s *Service) ShowProcessCounts() (code int, rsp []byte, e *errors.CodedError) {
	if err := s.setFleet(); err != nil {
		if err == orm.ErrNoRows {
			return 0, nil, errors.NewErrorF(errors.InvalidParameterValue, " fleet not found")

		} else {
			s.logger.Error("find fleet info db error: %v", err)
			return 0, nil, errors.NewError(errors.DBError)
		}
	}

	code, rsp, err := s.forwardShowProcessCountsToAPPGW()
	if err != nil {
		s.logger.Error("forward get request to app gateway error: %v", err)
		return 0, nil, errors.NewError(errors.ServerInternalError)
	}

	return code, rsp, nil
}

func (s *Service) GenerateVPCAndSubnetCidrForCreateFleet(vpcId string, subnetId string, fleetId string) (
	*fleet.FleetVpcAndSubnet, *errors.CodedError) {
	// 指定了VPC获取VPC的网段信息
	if vpcId != "" {
		if subnetId == "" {
			return nil, errors.NewErrorF(errors.InvalidParameterValue,
				"subnet_id required when have vpc_id")
		} else {
			originProjectId := s.ctx.Input.Param(params.ProjectId)
			region := s.createRequest.Region
			return GetVpcAndSubnetFromCloudResource(originProjectId, region, vpcId, subnetId)
		}
	}
	// 不指定则新建
	return GetVpcAndSubnetFromcidrmanager(fleetId)

}
