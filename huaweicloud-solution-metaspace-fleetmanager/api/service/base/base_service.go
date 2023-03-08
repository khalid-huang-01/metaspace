// Copyright (c) Huawei Technologies Co., Ltd. 2022-2022. All rights reserved.

// 基础服务定义
package base

import (
	"encoding/json"
	"fleetmanager/api/errors"
	"fleetmanager/api/model/fleet"
	"fleetmanager/api/params"
	"fleetmanager/api/common/query"
	"fleetmanager/api/service/constants"
	"fleetmanager/client"
	"fleetmanager/db/dao"
	"fleetmanager/logger"
	"fleetmanager/api/user"
	"fleetmanager/resdomain/service"
	"fmt"
	"net/http"
	"strconv"

	"github.com/beego/beego/v2/client/orm"
	"github.com/beego/beego/v2/server/web/context"
)

type FleetService struct {
	Ctx    *context.Context
	Logger *logger.FMLogger
	Fleet  *dao.Fleet
}

// SetFleet设置Fleet对象
func (s *FleetService) SetFleet() *errors.CodedError {
	fleetId := s.Ctx.Input.Param(params.FleetId)
	if fleetId == "" {
		return errors.NewError(errors.MissingFleetId)
	}

	filter := dao.Filters{
		"Id":         fleetId,
		"ProjectId":  s.Ctx.Input.Param(params.ProjectId),
		"Terminated": false,
	}
	f, err := dao.GetFleetStorage().Get(filter)
	if err != nil {
		if err == orm.ErrNoRows {
			return errors.NewError(errors.FleetNotFound)
		}
		s.Logger.Error("get fleet info db error: %v", err)
		return errors.NewError(errors.DBError)
	}

	s.Fleet = f
	return nil
}

// SetFleet设置Fleet对象
func (s *FleetService) SetFleetById(fleetId string) *errors.CodedError {
	if fleetId == "" {
		return errors.NewError(errors.MissingFleetId)
	}

	filter := dao.Filters{
		"Id":         fleetId,
		"ProjectId":  s.Ctx.Input.Param(params.ProjectId),
		"Terminated": false,
	}
	f, err := dao.GetFleetStorage().Get(filter)
	if err != nil {
		if err == orm.ErrNoRows {
			return errors.NewError(errors.FleetNotFound)
		}
		s.Logger.Error("get fleet info db error: %v", err)
		return errors.NewError(errors.DBError)
	}

	s.Fleet = f
	return nil
}

// SetFleetByServerSessionId设置Fleet对象
func (s *FleetService) SetFleetByServerSessionId(fssId string) *errors.CodedError {
	if fssId == "" {
		return errors.NewError(errors.MissingServerSessionId)
	}

	filter := dao.Filters{"ServerSessionId": fssId}
	fss, err := dao.GetFleetServerSessionStorage().GetOne(filter)
	if err != nil {
		if err == orm.ErrNoRows {
			return errors.NewError(errors.ServerSessionNotFound)
		}
		s.Logger.Error("get fleet info db error: %v", err)
		return errors.NewError(errors.DBError)
	}

	if err := s.SetFleetById(fss.FleetId); err != nil {
		return err
	}

	return nil
}

// SetFleet设置Fleet对象 通过获取Query参数
func (s *FleetService) SetFleetByQuery() *errors.CodedError {
	fleetId := s.Ctx.Input.Query(params.QueryFleetId)
	filter := dao.Filters{"Id": fleetId}
	f, err := dao.GetFleetStorage().Get(filter)
	if err != nil {
		if err == orm.ErrNoRows {
			return errors.NewError(errors.FleetNotFound)
		}
		s.Logger.Error("get fleet info db error: %v", err)
		return errors.NewError(errors.DBError)
	}

	s.Fleet = f
	return nil
}

// RspCheck
func (s *FleetService) ForwardRspCheck(code int, rsp []byte, err error) (int, []byte, *errors.CodedError){
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

func GetFleetQueryField() []string {
	var query_fileds []string
	query_fileds = append(query_fileds, params.QueryFleetId)
	return query_fileds
}

func GetInstancesQueryField() []string {
	var query_fileds []string
	query_fileds = append(query_fileds, params.QueryFleetId, params.Limit, params.ParamOffset, 
		params.InstanceId, params.HealthState, params.LifeCycleState, 
		params.DurationStart, params.DurationEnd)
	return query_fileds
}

func GetAppProcessesQueryField() []string {
	var query_fileds []string
	query_fileds = append(query_fileds, params.QueryFleetId, params.InstanceId, params.ProcessId,
		params.State, params.IpAddress, params.DurationStart, params.DurationEnd,
		params.ParamOffset, params.ParamLimit)
	return query_fileds
}

func GetServerSessionsQueryFiled() []string {
	var query_fileds []string
	query_fileds = append(query_fileds, params.QueryFleetId, params.InstanceId, params.ProcessId,
		params.QueryServerSessionId, params.State, params.IpAddress, params.StartTime, params.EndTime, 
		params.ParamOffset, params.ParamLimit)
	return query_fileds
}

func GetQueryParams(c *context.Context, query_fileds []string) map[string]string {
	query_params := make(map[string]string)
	for _, filed := range query_fileds {
		param := c.Input.Query(filed)
		if param == "" {
			param = c.Input.Param(":" + filed)
		}
		query_params[filed] = param
	}
	return query_params
}

func ForwardToAppgw(c *context.Context, region string, url string, params map[string]string) (
	code int, rsp []byte, err error) {
	url = client.GetServiceEndpoint(client.ServiceNameAPPGW, region) + url
	req := client.NewRequest(client.ServiceNameAPPGW, url, http.MethodGet, nil)
	for key, value := range params {
		req.SetQuery(key, value)
	}
	req.SetHeader(map[string]string{
		logger.RequestId: fmt.Sprintf("%s", c.Input.GetData(logger.RequestId)),
	})
	return req.DoRequest()
}

func ForwardToAASS(c *context.Context, region string, url string, params map[string]string) (
	code int, rsp []byte, err error) {
	url = client.GetServiceEndpoint(client.ServiceNameAASS, region) + url
	req := client.NewRequest(client.ServiceNameAASS, url, http.MethodGet, nil)
	for key, value := range params {
		req.SetQuery(key, value)
	}
	req.SetHeader(map[string]string{
		logger.RequestId: fmt.Sprintf("%s", c.Input.GetData(logger.RequestId)),
	})
	return req.DoRequest()
}

func AcceptResp(logger *logger.FMLogger, m interface{}, 
	code int, rsp []byte, err error) *errors.CodedError {
	if code < http.StatusOK || code >= http.StatusBadRequest {
		// 创建失败的场景
		errCode := &errors.CodedError{}
		if newErr := json.Unmarshal(rsp, &errCode); newErr != nil {
			logger.Error("unmarshal rep: %s to %+v error: %v", rsp, errCode, newErr)
			return errors.NewError(errors.ServerInternalError)
		}
		return errCode
		// 其他错误场景直接透传
	}

	if newErr := json.Unmarshal(rsp, &m); newErr != nil {
		logger.Error("unmarshal rep: %s to %+v error: %v", rsp, m, newErr)
		return errors.NewError(errors.ServerInternalError)
	}
	return nil
}

func GetDuration(CtxParams map[string]string, tLogger *logger.FMLogger) (int, int, error) {
	var start, end = 0, query.MaxServerSessionNum
	var err error
	if CtxParams[params.DurationStart] == "" {
		tLogger.Info("[query checker] duration start is not valid, not set")
	} else {
		start, err = strconv.Atoi(CtxParams[params.DurationStart])
		if err != nil {
			return start, end, fmt.Errorf("invalid duration start, please check")
		}
	}
	if CtxParams[params.DurationEnd] == "" {
		tLogger.Info("[query checker] duration end is not valid, not set")
	} else {
		end, err = strconv.Atoi(CtxParams[params.DurationEnd])
		if err != nil {
			return start, end, fmt.Errorf("invalid duration end, please check")
		}
	}
	if start < 0 || end < 0 || start > query.MaxServerSessionNum || end > query.MaxServerSessionNum {
		return start, end, fmt.Errorf("duration strat and end should be in [0, %d], please check", query.MaxServerSessionNum)
	}

	if start > end {
		return start, end, fmt.Errorf("duration start larger duration end, please check")
	}
	
	return start, end, nil
}

func GenerateInstances(CtxParams map[string]string, resp_aass *fleet.ListInstanceResonseFromAASS, 
	appg_instance map[string]map[string]int, ips map[string]string, tLogger *logger.FMLogger) (
	*fleet.ListMonitorInstancesResponce, *errors.CodedError) {
	resp := fleet.ListMonitorInstancesResponce{
		TotalCount: 	resp_aass.TotalNumber,
		Instances: 		[]fleet.MonitorInstanceResponce{},
	}
	duration_start, duration_end, err := GetDuration(CtxParams, tLogger)
	if err != nil {
		return nil, errors.NewErrorF(errors.InvalidParameterValue, err.Error())
	}
	instance_id					 := CtxParams[params.InstanceId]

	for _, ins := range resp_aass.Instances {
		if duration_start >= 0 && appg_instance[ins.InstanceId]["server_session_count"] < duration_start {
			continue
		}
		if duration_end >= 0 && appg_instance[ins.InstanceId]["server_session_count"] > duration_end {
			continue
		}
		if instance_id == "" || instance_id == ins.InstanceId {
			resp.Instances = append(resp.Instances, 
			*GenerateInstancesResponce(appg_instance[ins.InstanceId], ips[ins.InstanceId], &ins))
			resp.Count += 1
			if instance_id != "" {
				break
			}
		}
	}
	return &resp, nil
}

func GenerateInstancesResponce(appw map[string]int, ip string, 
	aass *fleet.InstanceResponseFromAASS) *fleet.MonitorInstanceResponce {
	var process_count int
	var server_session_count int
	var max_server_session_num int
	if appw == nil {
		process_count = 0
		server_session_count = 0
		max_server_session_num = 0
	} else {
		process_count = appw["process_count"]
		server_session_count = appw["server_session_count"]
		max_server_session_num = appw["max_server_session_num"]
	}
	return &fleet.MonitorInstanceResponce{
		InstanceId:         aass.InstanceId,
		InstanceName:       aass.InstanceName,
		LifeCycleState:     aass.LifeCycleState,
		HealthStatus:       aass.HealthStatus,
		CreatedAt:          aass.CreatedAt.Format(constants.TimeFormatLayout),
		ProcessCount:       process_count,
		IpAddress:          ip,
		ServerSessionCount: server_session_count,
		MaxServerSessionNum: max_server_session_num,
	}
}

// @Title GetResDomainInfo
// @Description  get resDomainInfo by originProject
// @Author wangnannan 2022-05-07 09:21:40 ${time}
// @Param originProjectId
// @Param region
// @Return string
// @Return string
// @Return *dao.ResAgency
// @Return string
// @Return *errors.CodedError
func GetResDomainInfo(originProjectId string, region string) (*user.ResDomainAgency, *errors.CodedError) {

	var resDomainInfo user.ResDomainAgency

	// Get ResDomain info first
	domainAPI := service.ResDomain{}
	resUser, err := domainAPI.GetUser(originProjectId, region)
	if err != nil {
		return &resDomainInfo,
			errors.NewErrorF(errors.ServerInternalError, "Get ResDomianInfo by projectId failed!", err)
	}
	resDomainInfo.ResDomainId = resUser.ResDomainId
	resDomainInfo.OriginUserDomain = resUser.OriginDomainId

	// 获取资源租户ProjectID
	resProject, err := domainAPI.GetProject(originProjectId, region)
	if err != nil {
		return &resDomainInfo,
			errors.NewErrorF(errors.ServerInternalError, "Get ResUser projectID  failed!", err)
	}
	resDomainInfo.ResProjectId = resProject.ResProjectId

	// 获取资源租户委托名称
	agency, err := domainAPI.GetAgency(resUser.OriginDomainId, region)
	if err != nil {
		return &resDomainInfo,
			errors.NewErrorF(errors.ServerInternalError, "Get ResUser agency info failed!", err)
	}
	resDomainInfo.AgencyName = agency.AgencyName

	return &resDomainInfo, nil

}

// @Title GetOriginDomainInfo
// @Description  get  originDomainID and agencyName return originDomainID agencyName
// @Author wangnannan 2022-05-07 09:21:59 ${time}
// @Param originProjectId
// @Param region
// @Return string
// @Return string
// @Return *errors.CodedError
func GetOriginDomainInfo(originProjectId string, region string) (string, string, *errors.CodedError) {

	domainAPI := service.ResDomain{}
	user, err := domainAPI.GetUser(originProjectId, region)
	if err != nil {
		return "", "",
		errors.NewErrorF(errors.ServerInternalError, "Get OriginDomainInfo by projectIdf failed!", err)
	}

	// 获取原始租户对服务的委托名称
	agency, err := domainAPI.GetOriginUserAgency(user.OriginDomainId, region)
	if err != nil {
		return "", "", errors.NewErrorF(errors.ServerInternalError, "Get OriginUserAgency info failed!", err)
	}

	return user.OriginDomainId, agency.AgencyName, nil

}