// Copyright (c) Huawei Technologies Co., Ltd. 2022-2022. All rights reserved.

// 资源配置
package cloudresource

import (
	"encoding/base64"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/huaweicloud/huaweicloud-sdk-go-v3/core/auth/basic"
	"github.com/huaweicloud/huaweicloud-sdk-go-v3/core/sdkerr"
	as "github.com/huaweicloud/huaweicloud-sdk-go-v3/services/as/v1"
	asmodel "github.com/huaweicloud/huaweicloud-sdk-go-v3/services/as/v1/model"
	ecs "github.com/huaweicloud/huaweicloud-sdk-go-v3/services/ecs/v2"
	ecsmodel "github.com/huaweicloud/huaweicloud-sdk-go-v3/services/ecs/v2/model"
	"github.com/pkg/errors"

	"scase.io/application-auto-scaling-service/pkg/api/model"
	"scase.io/application-auto-scaling-service/pkg/common"
	"scase.io/application-auto-scaling-service/pkg/db"
	"scase.io/application-auto-scaling-service/pkg/setting"
	"scase.io/application-auto-scaling-service/pkg/utils"
	"scase.io/application-auto-scaling-service/pkg/utils/logger"
)

const (
	// waitGroupStableTimes 等待伸缩组稳定的最大次数
	waitGroupStableTimes = 60
	// eachWaitDurationForGroupStable 每次等待时长
	eachWaitDurationForGroupStable = 10 * time.Second

	// waitVmShutoffTimes 等待vm关机的最大次数
	waitVmShutoffTimes = 120
	// eachWaitDurationForVmShutoff 每次等待时长
	eachWaitDurationForVmShutoff = 30 * time.Second

	// resource controller 有效时长
	agencyValidDuration = 23 * time.Hour

	// ecs 虚机状态
	ecsServerStatusShutoff = "SHUTOFF"
	ecsServerStatusDeleted = "DELETED"

	// 批量移除as实例操作，单次最多操作50个实例
	batchRemoveAsInstancesLimit = 50
	// 批量关闭vm操作，单次最多操作1000个vm
	batchStopServersLimit = 1000
	// 获取as伸缩组实例信息，单次最多获取100个实例信息
	getAsScalingInstanceLimit = 100

	chargingModeBandwidth = "bandwidth"

	ipType5GVM    = "5_g-vm"
	ipType5SBGP   = "5_sbgp"
	ipType5TELCOM = "5_telcom"
	ipType5UNION  = "5_union"

	volumeTypeSAS   = "SAS"
	volumeTypeSSD   = "SSD"
	volumeTypeGPSSD = "GPSSD"
	volumeTypeCOP1  = "co-p1"
	volumeTypeUHL1  = "uh-l1"

	asBatchRemoveFiledErrorCode = "AS.4030"
	asInstanceNotExistErrorCode = "AS.4006"
)

var resCtrlMgmt = ResourceControllerMgmt{
	resCtrlMap: make(map[string]*ResourceController),
}

type ResourceControllerMgmt struct {
	resCtrlMap map[string]*ResourceController
	lock       sync.Mutex
}

// GetResourceController get resource controller of the resource tenant
func GetResourceController(projectId string) (*ResourceController, error) {
	resCtrlMgmt.lock.Lock()
	defer resCtrlMgmt.lock.Unlock()

	cs, ok := resCtrlMgmt.resCtrlMap[projectId]
	if ok && cs.isValid() {
		return cs, nil
	}

	agencyInfo, err := db.GetAgencyInfo(projectId)
	if err != nil {
		return nil, err
	}
	controller, err := newResourceController(agencyInfo.ProjectId, agencyInfo.AgencyName, agencyInfo.DomainId)
	if err != nil {
		return nil, err
	}
	resCtrlMgmt.resCtrlMap[projectId] = controller
	return controller, nil
}

// ResourceController controller for cloud resource
type ResourceController struct {
	asClient  *as.AsClient
	ecsClient *ecs.EcsClient

	expireAt  time.Time
	projectId string
}

func (c *ResourceController) isValid() bool {
	return time.Now().Before(c.expireAt)
}

// AsClient return as client
func (c *ResourceController) AsClient() *as.AsClient {
	return c.asClient
}

// EcsClient return ecs client
func (c *ResourceController) EcsClient() *ecs.EcsClient {
	return c.ecsClient
}

func newResourceController(projectId string, agencyName string, resDomainId string) (*ResourceController, error) {
	credInfo, err := opIamCli.getAgencyCredentialInfo(resDomainId, agencyName)
	if err != nil {
		return nil, err
	}
	cred := basic.NewCredentialsBuilder().
		WithAk(credInfo.Access).
		WithSk(credInfo.Secret).
		WithSecurityToken(credInfo.SecurityToken).
		WithProjectId(projectId).
		Build()

	// 使用资源账号永久AK/SK创VM：①规避功能开启；②当前环境为规避环境；③当前请求的projectID为规避的资源账号projectID
	if setting.AvoidingAgencyEnable && setting.AvoidingAgencyRegion == setting.CloudClientRegion &&
		setting.ResUserProjectId == projectId {
		cred = basic.NewCredentialsBuilder().
			WithAk(string(setting.ResUserAk)).
			WithSk(string(setting.ResUserSk)).
			WithProjectId(setting.ResUserProjectId).
			Build()
	}

	return &ResourceController{
		asClient:  newAsClient(cred, projectId),
		ecsClient: newEcsClient(cred, projectId),
		expireAt:  time.Now().Add(agencyValidDuration),
		projectId: projectId,
	}, nil
}

// CreateAsScalingConfig 创建伸缩配置
func (c *ResourceController) CreateAsScalingConfig(log *logger.FMLogger, fleetId string, instanceScalingGroupId string,
	vmTemplate *model.VmTemplate) (string, error) {
	if vmTemplate == nil {
		return "", errors.New("The vm template for creating as scaling config is null")
	}
	flavorIds := strings.Join(vmTemplate.AvailableFlavorIds, ",")
	request := &asmodel.CreateScalingConfigRequest{
		Body: &asmodel.CreateScalingConfigOption{
			ScalingConfigurationName: fmt.Sprintf("Fleet_%s_aass_config", fleetId),
			InstanceConfig: &asmodel.InstanceConfig{
				FlavorRef:      &flavorIds,
				ImageRef:       vmTemplate.ImageID,
				Disk:           setDisks(vmTemplate.Disks),
				PublicIp:       setEip(vmTemplate.Eip),
				KeyName:        vmTemplate.KeyName,
				UserData:       genUserData(fleetId, instanceScalingGroupId),
				SecurityGroups: setSecurityGroups(vmTemplate.SecurityGroups),
			},
		}}
	resp, err := c.asClient.CreateScalingConfig(request)
	if err != nil {
		log.Info("CreateScalingConfig request body: %+v", *request)
		return "", errors.Wrap(err, "as client create scaling config err")
	}
	asScalingConfigId := *resp.ScalingConfigurationId
	log.Info("Create AsScalingConfig[%s] of ScalingGroup[%s] success", asScalingConfigId, instanceScalingGroupId)
	return asScalingConfigId, nil
}

// DeleteAsScalingConfig 删除伸缩配置
func (c *ResourceController) DeleteAsScalingConfig(log *logger.FMLogger, configId string) error {
	req := &asmodel.DeleteScalingConfigRequest{
		ScalingConfigurationId: configId,
	}
	_, err := c.asClient.DeleteScalingConfig(req)
	if err != nil {
		// as伸缩配置已删除，直接返回
		newErr, ok := err.(*sdkerr.ServiceResponseError)
		if ok && (newErr.StatusCode == http.StatusNotFound) {
			log.Info("The as scaling config[%s] has been deleted, do nothing", configId)
			return nil
		}
		return errors.Wrapf(err, "as client delete scaling config[%s] err", configId)
	}
	log.Info("Delete scaling config[%s] success", configId)
	return nil
}

// CreateAsScalingGroup 创建ScalingGroup
func (c *ResourceController) CreateAsScalingGroup(log *logger.FMLogger, params CreatAsGroupParams,
	groupId, resourceId string) (string, error) {
	var (
		desireInstanceNumber       int32 = 0
		minInstanceNumber          int32 = 0
		maxInstanceNumber                = int32(setting.GetInstanceMaximumLimitPreGroup())
		EnterpriseProjectId       		 = setting.GetEnterpriseProjectId()
		deleteEIP                        = true
		deleteVol                        = true
	)
	if params.EnterpriseProjectId != "" {
		EnterpriseProjectId = params.EnterpriseProjectId
	}
	// 创建AS伸缩组
	createScalingGroupReq := &asmodel.CreateScalingGroupRequest{
		Body: &asmodel.CreateScalingGroupOption{
			ScalingGroupName:       fmt.Sprintf("Fleet_%s_aass_group", params.FleetId),
			ScalingConfigurationId: &params.AsConfigId,
			MinInstanceNumber:      &minInstanceNumber,
			MaxInstanceNumber:      &maxInstanceNumber,
			DesireInstanceNumber:   &desireInstanceNumber,
			VpcId:                  params.VpcId,
			Networks: []asmodel.Networks{{
				Id: params.SubnetId,
			}},
			DeletePublicip:      &deleteEIP,
			DeleteVolume:        &deleteVol,
			EnterpriseProjectId: &EnterpriseProjectId,
			IamAgencyName:       &params.IamAgencyName,
		},
	}
	createResp, err := c.asClient.CreateScalingGroup(createScalingGroupReq)
	if err != nil {
		log.Info("CreateScalingGroup request body: %+v", *createScalingGroupReq.Body)
		return "", errors.Wrap(err, "as client create ScalingGroup err")
	}
	asGroupId := *createResp.ScalingGroupId
	log.Info("Create AsScalingGroup[%s] of ScalingGroup[%s] success", asGroupId, groupId)
	return asGroupId, nil
}

// ResumeAsScalingGroup 启用弹性伸缩组
func (c *ResourceController) ResumeAsScalingGroup(log *logger.FMLogger, groupId, asGroupId string) error {
	// 启用弹性伸缩组
	resumeScalingGroupReq := &asmodel.ResumeScalingGroupRequest{
		ScalingGroupId: asGroupId,
		Body: &asmodel.ResumeScalingGroupOption{
			Action: asmodel.GetResumeScalingGroupOptionActionEnum().RESUME,
		},
	}
	_, err := c.asClient.ResumeScalingGroup(resumeScalingGroupReq)
	if err != nil {
		return errors.Wrapf(err, "as client resume scaling group[%s] err", asGroupId)
	}
	log.Info("Resume AsScalingGroup[%s] of ScalingGroup[%s] success", asGroupId, groupId)
	return nil
}

// DelAsGroup 删除伸缩组
func (c *ResourceController) DelAsGroup(log *logger.FMLogger, groupId string) error {
	// 停止AS伸缩组
	pauseReq := &asmodel.PauseScalingGroupRequest{
		ScalingGroupId: groupId,
		Body: &asmodel.PauseScalingGroupOption{
			Action: asmodel.GetPauseScalingGroupOptionActionEnum().PAUSE,
		},
	}
	_, err := c.asClient.PauseScalingGroup(pauseReq)
	if err != nil {
		// as伸缩组已删除，直接返回
		respErr, ok := err.(*sdkerr.ServiceResponseError)
		if ok && (respErr.StatusCode == http.StatusNotFound) {
			log.Info("The as group[%s] has been deleted, do nothing", groupId)
			return nil
		}
		return errors.Wrapf(err, "as client pause scaling group[%s] err", groupId)
	}

	// 关闭所有instance
	instanceIds, err := c.GetAsScalingInstanceIds(log, groupId)
	if err != nil {
		return err
	}
	if len(instanceIds) != 0 {
		err = c.BatchStopServers(log, instanceIds)
		if err != nil {
			return err
		}
	}
	// 等待所有vm关机完毕
	for _, id := range instanceIds {
		if err = c.WaitVmShutoff(log, id); err != nil {
			return err
		}
	}

	// 删除as伸缩组
	forceDel := asmodel.GetDeleteScalingGroupRequestForceDeleteEnum().YES
	var delReq = &asmodel.DeleteScalingGroupRequest{
		ScalingGroupId: groupId,
		ForceDelete:    &forceDel,
	}
	_, err = c.asClient.DeleteScalingGroup(delReq)
	if err != nil {
		return errors.Wrapf(err, "as client delete scaling group[%s] err", groupId)
	}
	log.Info("Delete scaling group[%s] success", groupId)
	return nil
}


// 为弹性伸缩组新建或删除标签，覆盖类型，生命周期随弹性伸缩组
func (c *ResourceController) UpdateAsScalingGroupTags(groupId string, tags []model.InstanceTag) error {
	// 先获取资源标签列表，对比查询出的标签与传入的标签，判断哪些标签需要删除，那些标签需要修改或新建
	listTagsReq := &asmodel.ListScalingTagInfosByResourceIdRequest{
		ResourceType: asmodel.GetListScalingTagInfosByResourceIdRequestResourceTypeEnum().SCALING_GROUP_TAG,
		ResourceId: groupId,
	}
	listTagsResp, err := c.asClient.ListScalingTagInfosByResourceId(listTagsReq)
	if err != nil {
		return err
	}
	createTags, deleteTags, err := generateCreateAndDeleteTagModel(*listTagsResp.Tags, tags)
	if err != nil {
		return err
	}
	if len(createTags) > 0 {
		if err = c.CreateAsScalingGroupTags(groupId, createTags); err != nil {
			return err
		}
	}
	if len(deleteTags) > 0 {
		if err = c.DeleteAsScalingGroupTags(groupId, deleteTags); err != nil {
			return err
		}
	}
	return nil
}
 
// 创建弹性伸缩组的tag
func (c *ResourceController) CreateAsScalingGroupTags(groupId string, tags []asmodel.TagsSingleValue) error {
	createReq := &asmodel.CreateScalingTagInfoRequest{
		ResourceType: asmodel.GetCreateScalingTagInfoRequestResourceTypeEnum().SCALING_GROUP_TAG,
		ResourceId: groupId,
		Body: &asmodel.CreateTagsOption{
			Tags: tags,
			Action: asmodel.GetCreateTagsOptionActionEnum().CREATE,
		},
	}
	_, err := c.asClient.CreateScalingTagInfo(createReq)
	if err != nil {
		return err
	}
	return nil
}
 
// 删除弹性伸缩组的tag
func (c *ResourceController) DeleteAsScalingGroupTags(groupId string, tags []asmodel.TagsSingleValue) error {
	deleteReq := &asmodel.DeleteScalingTagInfoRequest{
		ResourceType: asmodel.GetDeleteScalingTagInfoRequestResourceTypeEnum().SCALING_GROUP_TAG,
		ResourceId: groupId,
		Body: &asmodel.DeleteTagsOption{
			Tags: tags,
			Action: asmodel.GetDeleteTagsOptionActionEnum().DELETE,
		},
	}
	_, err := c.asClient.DeleteScalingTagInfo(deleteReq)
	if err != nil {
		return err
	}
	return nil
 
}

// GetAsGroupCurrentInstanceNum 获取当前as伸缩组的当前实例个数
func (c *ResourceController) GetAsGroupCurrentInstanceNum(groupId string) (int32, error) {
	resp, err := c.asClient.ShowScalingGroup(&asmodel.ShowScalingGroupRequest{
		ScalingGroupId: groupId})
	if err != nil {
		return 0, errors.Wrapf(err, "as client show ScalingGroup[%s] err", groupId)
	}
	return *resp.ScalingGroup.CurrentInstanceNumber, nil
}

// WaitAsGroupStable 等待ScalingGroup中server数量达到预期，若长期(10min)达不到预期，则修改期望值
func (c *ResourceController) WaitAsGroupStable(tLogger *logger.FMLogger, groupId string) error {
	waitTimes := 0
	for {
		resp, err := c.asClient.ShowScalingGroup(&asmodel.ShowScalingGroupRequest{ScalingGroupId: groupId})
		if err != nil {
			return errors.Wrapf(err, "as client show ScalingGroup[%s] err", groupId)
		}
		curNum := *resp.ScalingGroup.CurrentInstanceNumber
		desireNum := *resp.ScalingGroup.DesireInstanceNumber
		tLogger.Info("Waiting as group[%s] stable(curNum[%d]/desireNum[%d])……", groupId, curNum, desireNum)
		if curNum == desireNum {
			break
		}
		if waitTimes == waitGroupStableTimes {
			// 更新DesireInstanceNumber
			_, err = c.asClient.UpdateScalingGroup(&asmodel.UpdateScalingGroupRequest{
				ScalingGroupId: groupId,
				Body: &asmodel.UpdateScalingGroupOption{
					DesireInstanceNumber: &curNum,
				},
			})
			if err != nil {
				return errors.Wrapf(err, "as client update ScalingGroup[%s] desire_num[%d] err", groupId, curNum)
			}
			tLogger.Info("The instances num of scaling group[%s] failed to reach the desire number for a long time,"+
				" group detail: '%s', set desireNum[%d] = curNum[%d]", groupId,
				utils.ToJson(resp.ScalingGroup.Detail), desireNum, curNum)
			break
		}
		time.Sleep(eachWaitDurationForGroupStable)
		waitTimes++
	}
	return nil
}

// GetAsScalingInstanceIds 获取ScalingGroup中所有实例的id
// 需要等待伸缩组稳定，不要在同步方法中调用
func (c *ResourceController) GetAsScalingInstanceIds(log *logger.FMLogger, groupId string) ([]string, error) {
	// 等待实例数变为DesireInstanceNumber
	if err := c.WaitAsGroupStable(log, groupId); err != nil {
		return nil, err
	}

	// 查询弹性伸缩组中的实例列表
	// 获取前100个实例信息
	instanceIds, err := c.listAsScalingInstanceIds(log, groupId, 0, getAsScalingInstanceLimit)
	if err != nil {
		return nil, err
	}
	// 若结果id数量小于100，直接返回
	if int32(len(instanceIds)) < getAsScalingInstanceLimit {
		return instanceIds, nil
	}

	// 获取后100个实例信息
	otherIds, err := c.listAsScalingInstanceIds(log, groupId, getAsScalingInstanceLimit, getAsScalingInstanceLimit)
	if err != nil {
		return nil, err
	}
	if len(otherIds) > 0 {
		instanceIds = append(instanceIds, otherIds...)
	}
	return instanceIds, nil
}

// listAsScalingInstanceIds ...
func (c *ResourceController) listAsScalingInstanceIds(log *logger.FMLogger, groupId string,
	startIndex, limit int32) ([]string, error) {
	serverIds := []string{}
	// 请求as获取伸缩组的实例信息
	resp, err := c.asClient.ListScalingInstances(&asmodel.ListScalingInstancesRequest{
		ScalingGroupId: groupId,
		StartNumber:    &startIndex,
		Limit:          &limit,
	})
	if err != nil {
		return nil, errors.Wrapf(err, "as client list instances of group[%s] err", groupId)
	}
	log.Info("AsClient list instance resp: %s", resp.String())

	// 便利获取实例Id，若vm处于初始化状态，可能没有InstanceId
	for _, instance := range *resp.ScalingGroupInstances {
		if instance.InstanceId == nil {
			return nil, errors.Wrapf(common.ErrAsInstanceIdInvalid,
				"id of instance[name: %s] invalid, try again later", utils.ToJson(instance.InstanceName))
		}
		serverIds = append(serverIds, *(instance.InstanceId))
	}
	return serverIds, nil
}

// BatchRemoveAsScalingInstances 移除as伸缩组的实例(不删除vm)
func (c *ResourceController) BatchRemoveAsScalingInstances(log *logger.FMLogger, groupId string,
	instanceIds []string) error {
	// 每50个实例分批执行
	var left, right int
	for i := 0; i < len(instanceIds)/batchRemoveAsInstancesLimit+1; i++ {
		left = i * batchRemoveAsInstancesLimit
		right = left + batchRemoveAsInstancesLimit
		if right > len(instanceIds) {
			right = len(instanceIds)
		}
		if left == right {
			break
		}
		if err := c.batchRemoveAsScalingInstances(log, groupId, instanceIds[left:right]); err != nil {
			return err
		}
	}
	return nil
}

// batchRemoveAsScalingInstances 移除as伸缩组的实例(不删除vm)，单次最多批量操作实例个数为50
func (c *ResourceController) batchRemoveAsScalingInstances(log *logger.FMLogger, groupId string,
	instanceIds []string) error {
	if len(instanceIds) > batchRemoveAsInstancesLimit {
		return errors.Errorf("the number[%d] of instance to be removed is greater than 50", len(instanceIds))
	}

	// 当伸缩组没有伸缩活动时，才能移除实例。不然会报错“AS group lock conflict.”
	if err := c.WaitAsGroupStable(log, groupId); err != nil {
		return err
	}

	// 向as发送移除实例请求，不删除vm
	notDelVm := asmodel.GetBatchRemoveInstancesOptionInstanceDeleteEnum().NO
	_, err := c.asClient.BatchRemoveScalingInstances(&asmodel.BatchRemoveScalingInstancesRequest{
		ScalingGroupId: groupId,
		Body: &asmodel.BatchRemoveInstancesOption{
			InstancesId:    instanceIds,
			InstanceDelete: &notDelVm,
			Action:         asmodel.GetBatchRemoveInstancesOptionActionEnum().REMOVE,
		},
	})
	if err != nil {
		// as实例已经移除，直接返回：
		// 1.将err解析为sdk服务详情响应的标准结构体：ServiceResponseError；
		// 2.若能正确解析，则判断请求响应的错误是否为 实例不存在：
		//   ①请求响应的状态码为404；或
		//   ②请求响应的错误码为AS.4030且错误信息中明确AS.4006(The instance does not exist)
		respErr, ok := err.(*sdkerr.ServiceResponseError)
		if ok && (respErr.StatusCode == http.StatusNotFound ||
			(respErr.ErrorCode == asBatchRemoveFiledErrorCode &&
				strings.Contains(respErr.ErrorMessage, asInstanceNotExistErrorCode))) {
			log.Info("Instances[%v] for as group[%s] has been removed and no further operation is required",
				instanceIds, groupId)
			return nil
		}
		return errors.Wrapf(err, "as client remove servers[%v] for as group[%s] err", instanceIds, groupId)
	}
	log.Info("Remove as scaling instances[%v] for group[%s] success", instanceIds, groupId)
	return nil
}

// DeleteVm ...
func (c *ResourceController) DeleteVm(log *logger.FMLogger, vmId string) error {
	// 删除虚机
	var deleteEIP = true
	var deleteVol = true
	delReq := &ecsmodel.DeleteServersRequest{
		Body: &ecsmodel.DeleteServersRequestBody{
			DeletePublicip: &deleteEIP,
			DeleteVolume:   &deleteVol,
			Servers:        []ecsmodel.ServerId{{Id: vmId}},
		},
	}
	_, err := c.ecsClient.DeleteServers(delReq)
	if err != nil {
		// vm已删除，直接返回
		respErr, ok := err.(*sdkerr.ServiceResponseError)
		if ok && (respErr.StatusCode == http.StatusNotFound) {
			log.Info("Vm[%s] has been removed and no further operation is required", vmId)
			return nil
		}
		log.Error("EcsClient delete servers err: %+v", err)
		return err
	}
	log.Info("Server[%s] deleted successfully", vmId)
	return nil
}

// WaitVmShutoff 等待vm关机
func (c *ResourceController) WaitVmShutoff(log *logger.FMLogger, serverId string) error {
	for {
		showReq := &ecsmodel.ShowServerRequest{ServerId: serverId}
		resp, err := c.ecsClient.ShowServer(showReq)
		if err != nil {
			// vm已删除，忽略
			respErr, ok := err.(*sdkerr.ServiceResponseError)
			if ok && (respErr.StatusCode == http.StatusNotFound) {
				log.Info("Vm[%s] has been removed and no further operation is required", serverId)
				break
			}
			return errors.Wrapf(err, "ecs client show server[%s] err", serverId)
		}
		if resp.Server.Status == ecsServerStatusShutoff || resp.Server.Status == ecsServerStatusDeleted {
			break
		}
		log.Info("Server[%s] status[%s], wait to be 'SHUTOFF'", serverId, resp.Server.Status)
		time.Sleep(eachWaitDurationForVmShutoff)
	}
	return nil
}

// BatchStopServers 根据给定的云服务器ID列表，批量关闭云服务器，一次最多可以关闭1000台
func (c *ResourceController) BatchStopServers(tLogger *logger.FMLogger, serverIds []string) error {
	if len(serverIds) == 0 {
		tLogger.Info("ServerIds are not specified, do nothing")
		return nil
	}
	if len(serverIds) > batchStopServersLimit {
		return errors.Errorf("the number[%d] of vm to be shut down is greater than 1000", len(serverIds))
	}

	// 关闭所有vm
	reqServerIds := make([]ecsmodel.ServerId, 0, len(serverIds))
	for _, id := range serverIds {
		reqServerIds = append(reqServerIds, ecsmodel.ServerId{Id: id})
	}
	stopReq := &ecsmodel.BatchStopServersRequest{
		Body: &ecsmodel.BatchStopServersRequestBody{
			OsStop: &ecsmodel.BatchStopServersOption{
				Servers: reqServerIds,
			}}}
	_, err := c.ecsClient.BatchStopServers(stopReq)
	if err != nil {
		return errors.Wrap(err, "ecs client batch stop server err")
	}
	tLogger.Info("Send stop req for servers[%v] success", serverIds)
	return nil
}

func setEip(eip model.Eip) *asmodel.PublicIp {
	var chargingModeTraffic asmodel.BandwidthInfoChargingMode
	if setting.GetBandwidthChargingMode() == chargingModeBandwidth {
		chargingModeTraffic = asmodel.GetBandwidthInfoChargingModeEnum().BANDWIDTH
	} else {
		chargingModeTraffic = asmodel.GetBandwidthInfoChargingModeEnum().TRAFFIC
	}

	bandwidth := asmodel.BandwidthInfo{
		Size:         eip.Bandwidth.Size,
		ShareType:    asmodel.GetBandwidthInfoShareTypeEnum().PER,
		ChargingMode: &chargingModeTraffic,
	}
	return &asmodel.PublicIp{Eip: &asmodel.EipInfo{
		IpType:    newAsEipInfoIpType(eip.IpType),
		Bandwidth: &bandwidth,
	}}
}

func setDisks(disks []model.Disk) *[]asmodel.DiskInfo {
	diskInfos := make([]asmodel.DiskInfo, len(disks))
	for i, d := range disks {
		diskInfos[i] = asmodel.DiskInfo{
			Size:       *d.Size,
			VolumeType: newAsDiskInfoVolumeType(*d.VolumeType),
			DiskType:   newAsDiskInfoDiskType(*d.DiskType),
		}
	}
	return &diskInfos
}

func setSecurityGroups(sg []model.SecurityGroup) []asmodel.SecurityGroups {
	groups := make([]asmodel.SecurityGroups, len(sg))
	for i, s := range sg {
		groups[i] = asmodel.SecurityGroups{
			Id: *s.Id,
		}
	}
	return groups
}

func newAsEipInfoIpType(ipType *string) asmodel.EipInfoIpType {
	var reqIpType string
	if ipType != nil {
		reqIpType = *ipType
	}
	ipTypeEnum := asmodel.GetEipInfoIpTypeEnum()
	switch reqIpType {
	default:
		return ipTypeEnum.E_5_BGP
	case ipType5SBGP:
		return ipTypeEnum.E_5_SBGP
	case ipType5TELCOM:
		return ipTypeEnum.E_5_TELCOM
	case ipType5UNION:
		return ipTypeEnum.E_5_UNION
	case ipType5GVM:
		type5GVM := asmodel.EipInfoIpType{}
		err := type5GVM.UnmarshalJSON([]byte(ipType5GVM))
		if err != nil {
			logger.R.Error("Unmarshal ipType5GVM err: %v, use 5_bgp instead", err)
			return ipTypeEnum.E_5_BGP
		}
		return type5GVM
	}
}

func newAsDiskInfoVolumeType(volumeType string) asmodel.DiskInfoVolumeType {
	volumeTypeEnum := asmodel.GetDiskInfoVolumeTypeEnum()
	switch volumeType {
	default:
		return volumeTypeEnum.SATA
	case volumeTypeSAS:
		return volumeTypeEnum.SAS
	case volumeTypeSSD:
		return volumeTypeEnum.SSD
	case volumeTypeGPSSD:
		return volumeTypeEnum.GPSSD
	case volumeTypeCOP1:
		return volumeTypeEnum.CO_PL
	case volumeTypeUHL1:
		return volumeTypeEnum.UH_11
	}
}

func newAsDiskInfoDiskType(diskType string) asmodel.DiskInfoDiskType {
	diskTypeEnum := asmodel.GetDiskInfoDiskTypeEnum()
	if diskType == common.DiskTypeDATA {
		return diskTypeEnum.DATA
	}
	return diskTypeEnum.SYS
}

func genUserData(fleetId, instanceScalingGroupId string) *string {
	data := struct {
		FleetId        string `json:"fleet_id"`
		ScalingGroupId string `json:"scaling_group_id"`
		GatewayAddress string `json:"gateway_address"`
	}{
		FleetId:        fleetId,
		ScalingGroupId: instanceScalingGroupId,
		GatewayAddress: setting.GetAppGwEndpoint(),
	}
	bytes := []byte(utils.ToJson(data))
	userData := base64.StdEncoding.EncodeToString(bytes)
	return &userData
}

// 将传入的tag列表转成map
func generateTagsMapFromInstanceTag(Tags []model.InstanceTag) (map[string]*string, error) {
	res := make(map[string]*string)
	for id, tag := range Tags {
		// 检查伸缩组标签重复则需要返回错误
		if _, ok := res[tag.Key]; ok {
			return nil, fmt.Errorf("scaling group tags duplicated")
		} else {
			// 不使用id遍历数组会导致值被覆盖
			res[tag.Key] = &Tags[id].Value
		}
	}
	return res, nil
}
 
func generateCreateAndDeleteTagModel(tagResp []asmodel.TagsSingleValue, 
	instanceTags []model.InstanceTag) ([]asmodel.TagsSingleValue, []asmodel.TagsSingleValue, error) {
	tagsMap, err := generateTagsMapFromInstanceTag(instanceTags)
	if err != nil {
		return nil, nil, err
	}
	createTags := []asmodel.TagsSingleValue{}
	deleteTags := []asmodel.TagsSingleValue{}
	for _, tag := range tagResp {
		if _, ok := tagsMap[tag.Key]; !ok {
			deleteTags = append(deleteTags, asmodel.TagsSingleValue{
				Key: tag.Key,
				Value: tag.Value,
			})
		}
	}
	for key, value := range tagsMap {
		createTags = append(createTags, asmodel.TagsSingleValue{
			Key: key,
			Value: value,
		})
	}
	return createTags, deleteTags, nil
}