// Copyright (c) Huawei Technologies Co., Ltd. 2022-2022. All rights reserved.

// 弹性伸缩组服务
package service

import (
	"net/http"
	"encoding/json"
	"github.com/google/uuid"
	pkgerrors "github.com/pkg/errors"
	wraperrors "github.com/pkg/errors"

	"scase.io/application-auto-scaling-service/pkg/api/errors"
	"scase.io/application-auto-scaling-service/pkg/api/model"
	"scase.io/application-auto-scaling-service/pkg/cloudresource"
	"scase.io/application-auto-scaling-service/pkg/common"
	"scase.io/application-auto-scaling-service/pkg/db"
	"scase.io/application-auto-scaling-service/pkg/metricmonitor"
	"scase.io/application-auto-scaling-service/pkg/service/taskservice"
	"scase.io/application-auto-scaling-service/pkg/utils/logger"
)

// 更新 group 触发的扩缩决策，决策失效时会重新决策，最多决策3次
const maxDecisionCount = 3

// CreateInstanceScalingGroup create instance scaling group
func CreateInstanceScalingGroup(log *logger.FMLogger, projectId string,
	req model.CreateScalingGroupReq) (*model.CreateScalingGroupResp, *errors.ErrorResp) {
	if err := db.AddOrUpdateAgencyInfo(getAgencyInfo(req, projectId)); err != nil {
		log.Error("Add or update AgencyInfo to db err: %+v", err)
		return nil, errors.NewErrorRespWithHttpCode(errors.ServerInternalError, http.StatusInternalServerError)
	}
	rc, err := cloudresource.GetResourceController(projectId)
	if err != nil {
		log.Error("Get resource controller of project[%s] err: %+v", projectId, err)
		return nil, errors.NewErrorRespWithHttpCode(errors.AgencyClientError, http.StatusBadRequest)
	}
	if exist := db.IsScalingGroupExistByName(projectId, *req.Name); exist {
		log.Error(errors.ScalingGroupNameExist.Msg())
		return nil, errors.NewErrorRespWithHttpCode(errors.ScalingGroupNameExist, http.StatusBadRequest)
	}
	group, err := addScalingGroupInfo(projectId, req)
	if err != nil {
		log.Error("Add scaling group info err: %+v", err)
		return nil, errors.NewErrorRespWithHttpCode(errors.ServerInternalError, http.StatusInternalServerError)
	}
	if err = creatCloudResourcesForScalingGroup(log, *rc, req, *group); err != nil {
		log.Error("Create cloud resources for scaling group[%s] error: %+v", group.Id, err)
		_ = db.UpdateScalingGroupState(group.Id, db.ScalingGroupStateError)
		_ = taskservice.StartDeleteScalingGroupTask(group.Id, metricmonitor.GetMgmt())
		return nil, errors.NewErrorRespWithHttpCodeAndMessage(errors.ServerInternalError,
			http.StatusInternalServerError, err.Error())
	}

	_ = db.UpdateScalingGroupVisibleState(group.Id, db.ScalingGroupStateStable)
	if group.DesireInstanceNumber > 0 {
		_ = taskservice.StartScaleOutGroupTask(group.Id, group.DesireInstanceNumber)
	}
	return &model.CreateScalingGroupResp{ScalingGroupId: group.Id}, nil
}

func addScalingGroupInfo(projectId string, req model.CreateScalingGroupReq) (*db.ScalingGroup, error) {
	ic, err := getNewInstanceConfiguration(*req.InstanceConfiguration, uuid.NewString())
	if err != nil {
		return nil, err
	}
	if err = db.AddInstanceConfiguration(ic); err != nil {
		return nil, err
	}
	resourceId := uuid.NewString()
	if err := db.AddVmScalingGroup(&db.VmScalingGroup{Id: resourceId}); err != nil {
		return nil, err
	}
	group, err := convertCreateScalingGroupReq(req, projectId, uuid.NewString(), resourceId, ic)
	if err != nil {
		return nil, err
	}
	if err = db.AddScalingGroup(group); err != nil {
		return nil, err
	}
	return group, nil
}

func creatCloudResourcesForScalingGroup(log *logger.FMLogger, rc cloudresource.ResourceController,
	req model.CreateScalingGroupReq, group db.ScalingGroup) error {
	// 创建as弹性伸缩组的弹性伸缩配置
	asConfId, err := rc.CreateAsScalingConfig(log, *req.FleetId, group.Id, req.VmTemplate)
	if err != nil {
		return err
	}
	if err = db.UpdateAsConfigIdOfVmScalingGroup(group.ResourceId, asConfId); err != nil {
		return err
	}
	// 创建as弹性伸缩组
	asGroupId, err := rc.CreateAsScalingGroup(log, getCreatAsGroupOption(req, asConfId), group.Id, group.ResourceId)
	if err != nil {
		return err
	}
	if err = db.UpdateAsGroupIdOfVmScalingGroup(group.ResourceId, asGroupId); err != nil {
		return err
	}
	if err = rc.ResumeAsScalingGroup(log, group.Id, asGroupId); err != nil {
		return err
	}
	// 绑定弹性伸缩组标签
	if err = rc.UpdateAsScalingGroupTags(asGroupId, req.InstanceTags); err != nil {
		return err
	}
	return nil
}

// DeleteInstanceScalingGroup delete instance scaling group
func DeleteInstanceScalingGroup(log *logger.FMLogger, projectId, groupId string) *errors.ErrorResp {
	if !db.IsScalingGroupExist(projectId, groupId) {
		log.Error("The scaling group[%s] of project[%s] is not found", groupId, projectId)
		return errors.NewErrorRespWithHttpCode(errors.ScalingGroupNotFound, http.StatusNotFound)
	}

	if err := taskservice.StartDeleteScalingGroupTask(groupId, metricmonitor.GetMgmt()); err != nil {
		// 任务已存在，直接返回
		if pkgerrors.Is(err, common.ErrDeleteTaskAlreadyExists) {
			log.Info("Delete task for group[%s] is running, do nothing", groupId)
			return nil
		}
		log.Error("Start delete scaling group task for group[%s] err: %+v", groupId, err)
		return errors.NewErrorRespWithHttpCode(errors.ServerInternalError, http.StatusInternalServerError)
	}
	return nil
}

// UpdateInstanceScalingGroup update instance scaling group
func UpdateInstanceScalingGroup(log *logger.FMLogger, projectId, groupId string,
	req model.UpdateScalingGroupReq) *errors.ErrorResp {
	if exist := db.IsScalingGroupExist(projectId, groupId); !exist {
		log.Error("The scaling group[%s] of project[%s] is not found", groupId, projectId)
		return errors.NewErrorRespWithHttpCode(errors.ScalingGroupNotFound, http.StatusNotFound)
	}
	group, err := db.GetScalingGroupById(projectId, groupId)
	if err != nil {
		log.Error("Read scaling group[%s] from db err: %+v", groupId, err)
		return errors.NewErrorRespWithHttpCode(errors.ServerInternalError, http.StatusInternalServerError)
	}
	rc, err := cloudresource.GetResourceController(projectId)
	if err != nil {
		log.Error("Get resource controller of project[%s] err: %+v", projectId, err)
		return errors.NewErrorRespWithHttpCode(errors.AgencyClientError, http.StatusBadRequest)
	}
	if errResp := updateJudgmentForInstanceNumberParam(req, group); errResp != nil {
		log.Error(errResp.ErrCode.Msg())
		return errResp
	}
	if req.EnableAutoScaling != nil {
		if len(group.ScalingPolicies) == 0 && *req.EnableAutoScaling {
			log.Error(errors.AutoScalingUpdateError.Msg())
			return errors.NewErrorRespWithHttpCode(errors.AutoScalingUpdateError, http.StatusBadRequest)
		}
		group.EnableAutoScaling = *req.EnableAutoScaling
	}
	if req.CoolDownTime != nil {
		group.CoolDownTime = int64(*req.CoolDownTime)
	}
	if errC := UpdateScalingGroupTagsAndInstanceConfiguration(rc, req, group, log); err != nil {
		return errC
	}
	if err = scalingJudgmentForUpdateScalingGroup(log, projectId, group); err != nil {
		if pkgerrors.Is(err, common.ErrScalingGroupNotStable) {
			log.Info("The instance num of group[%s] cannot be updated because group is locked", group.Id)
			return errors.NewErrorRespWithHttpCode(errors.GroupLockUpdateNumError, http.StatusBadRequest)
		}
		log.Error("Scaling judgment for group[%s] failed, err: %+v", group.Id, err)
		return errors.NewErrorRespWithHttpCode(errors.ServerInternalError, http.StatusInternalServerError)
	}
	if err = db.UpdateScalingGroup(group); err != nil {
		log.Error("Update scaling group[%s] to db err: %+v", group.Id, err)
		return errors.NewErrorRespWithHttpCode(errors.ServerInternalError, http.StatusInternalServerError)
	}
	return nil
}

func UpdateScalingGroupTagsAndInstanceConfiguration(rc *cloudresource.ResourceController, req model.UpdateScalingGroupReq, 
	group *db.ScalingGroup, log *logger.FMLogger) *errors.ErrorResp {
	// 获取AS弹性伸缩组的ID
	if req.InstanceTags != nil {
		AsGroupId, errC := GetScalingGroupByInsConfId(group.ResourceId, log)
		if errC != nil {
			return errC
		}
		if err := rc.UpdateAsScalingGroupTags(AsGroupId, req.InstanceTags); err != nil {
			log.Error("update scaling group tags %+v failed, err: %+v", req.InstanceTags, err)
			return errors.NewErrorRespWithHttpCode(errors.ServerInternalError, http.StatusInternalServerError)
		}
		tagsByte, err := json.Marshal(req.InstanceTags)
		if err != nil {
			log.Error("Marshal instance tags: %+v failed, err: %+v", req.InstanceTags, err)
			return errors.NewErrorRespWithHttpCode(errors.ServerInternalError, http.StatusInternalServerError)
		}
		group.InstanceTags = string(tagsByte)
	}
	if req.InstanceConfiguration != nil {
		instanceConfig, err := updateInstanceConfiguration(*req.InstanceConfiguration, group.InstanceConfiguration)
		if err != nil {
			log.Error("Update instance configuration err: %+v", err)
			return errors.NewErrorRespWithHttpCode(errors.ServerInternalError, http.StatusInternalServerError)
		}
		group.InstanceConfiguration = instanceConfig
	}
	
	return nil
}

// GetInstanceScalingGroup get instance scaling group detail
func GetInstanceScalingGroup(log *logger.FMLogger, projectId, groupId string) (*model.ScalingGroupDetail, *errors.ErrorResp) {
	if exist := db.IsScalingGroupExist(projectId, groupId); !exist {
		log.Error("The scaling group[%s] of project[%s] is not found", groupId, projectId)
		return nil, errors.NewErrorRespWithHttpCode(errors.ScalingGroupNotFound, http.StatusNotFound)
	}
	group, err := db.GetScalingGroupById(projectId, groupId)
	if err != nil {
		log.Error("Read scaling group[%s] from db err: %+v", groupId, err)
		return nil, errors.NewErrorRespWithHttpCode(errors.ServerInternalError, http.StatusInternalServerError)
	}
	detail := convertDaoScalingGroup(group)
	return &detail, nil
}

// ListInstanceScalingGroup list instance scaling group
func ListInstanceScalingGroup(log *logger.FMLogger, projectId, name string, limit, offset int) (model.ScalingGroupList,
	*errors.ErrorResp) {
	var list model.ScalingGroupList
	groups, err := db.ListScalingGroupByFilter(projectId, name, limit, offset)
	if err != nil {
		log.Error("List ScalingGroups of projectId[%s] from db err: %+v", projectId, err)
		return list, errors.NewErrorRespWithHttpCode(errors.ServerInternalError, http.StatusInternalServerError)
	}

	details := []model.ScalingGroupDetail{}
	for _, g := range groups {
		details = append(details, convertDaoScalingGroup(g))
	}
	list.Count = len(details)
	list.InstanceScalingGroups = details
	return list, nil
}

// GetInstanceConfigOfInstanceScalingGroup get instance configuration of instance scaling group
func GetInstanceConfigOfInstanceScalingGroup(log *logger.FMLogger,
	groupId string) (*model.InstanceConfiguration, *errors.ErrorResp) {
	if exist := db.IsScalingGroupExistWithoutProject(groupId); !exist {
		log.Error("The scaling group[%s] is not found", groupId)
		return nil, errors.NewErrorRespWithHttpCode(errors.ScalingGroupNotFound, http.StatusNotFound)
	}
	group, err := db.GetScalingGroupById("", groupId)
	if err != nil {
		log.Error("Read scaling group[%s] from db err: %+v", groupId, err)
		return nil, errors.NewErrorRespWithHttpCode(errors.ServerInternalError, http.StatusInternalServerError)
	}
	config, err := db.GetInstanceConfigurationById(group.InstanceConfiguration.Id)
	if err != nil {
		log.Error("Read instance configuration[%s] of scaling group[%s] from db err: %+v",
			group.InstanceConfiguration.Id, groupId, err)
		return nil, errors.NewErrorRespWithHttpCode(errors.ServerInternalError, http.StatusInternalServerError)
	}
	instanceConfig, err := convertDaoInstanceConfiguration(config)
	if err != nil {
		log.Error("Convert format for instance configuration[%s] of scaling group[%s] err: %+v",
			group.InstanceConfiguration.Id, groupId, err)
		return nil, errors.NewErrorRespWithHttpCode(errors.ServerInternalError, http.StatusInternalServerError)
	}
	return instanceConfig, nil
}

// scalingJudgmentForUpdateScalingGroup 更新伸缩组后的伸缩判断：
// 1.当EnableAutoScaling设置为true 且 实例伸缩组配置伸缩策略时，按伸缩策略进行弹性伸缩；
// 2.当实例伸缩组未配置伸缩策略 或 EnableAutoScaling设置为false时，按DesireInstanceNumber进行弹性伸缩。
func scalingJudgmentForUpdateScalingGroup(log *logger.FMLogger, projectId string,
	group *db.ScalingGroup) error {
	resCtrl, err := cloudresource.GetResourceController(projectId)
	if err != nil {
		return err
	}
	vmGroup, err := db.GetVmScalingGroupById(group.ResourceId)
	if err != nil {
		return err
	}
	// 若扩缩时，决策的结果（扩容目标数 或 缩容实例）已失效，需要重复决策
	decisionCount := 0
	for {
		curNum, err := resCtrl.GetAsGroupCurrentInstanceNum(vmGroup.AsGroupId)
		if err != nil {
			return err
		}
		log.Info("Instance scaling group[%s]: curInstanceNum is %d, desireInstanceNumber is %d,"+
			" minInstanceNumber is %d, maxInstanceNumber is %d, enableAutoScaling is %t, len(group.ScalingPolicies) is %d",
			group.Id, curNum, group.DesireInstanceNumber, group.MinInstanceNumber, group.MaxInstanceNumber,
			group.EnableAutoScaling, len(group.ScalingPolicies))
		if group.EnableAutoScaling && len(group.ScalingPolicies) != 0 {
			if group.MinInstanceNumber > curNum {
				err = taskservice.StartScaleOutGroupTask(group.Id, group.MinInstanceNumber)
			}
			if group.MaxInstanceNumber < curNum {
				err = taskservice.StartScaleInGroupTaskForRandomVms(group.Id, projectId, curNum-group.MaxInstanceNumber)
			}
		} else {
			if group.DesireInstanceNumber > curNum {
				err = taskservice.StartScaleOutGroupTask(group.Id, group.DesireInstanceNumber)
			}
			if group.DesireInstanceNumber < curNum {
				err = taskservice.StartScaleInGroupTaskForRandomVms(group.Id, projectId, curNum-group.DesireInstanceNumber)
			}
		}
		// err 为决策过期失效错误，重复决策；否则退出循环
		if !wraperrors.Is(err, common.ErrScalingDecisionExpired) {
			break
		}
		decisionCount++
		if decisionCount == maxDecisionCount {
			return wraperrors.Errorf("all [%d] decisions are expired, "+
				"there may be a logic error", maxDecisionCount)
		}
	}
	return err
}

func updateJudgmentForInstanceNumberParam(req model.UpdateScalingGroupReq, group *db.ScalingGroup) *errors.ErrorResp {
	if req.MaxInstanceNumber != nil {
		group.MaxInstanceNumber = *req.MaxInstanceNumber
		if group.DesireInstanceNumber > *req.MaxInstanceNumber {
			group.DesireInstanceNumber = *req.MaxInstanceNumber
		}
	}
	if req.MinInstanceNumber != nil {
		group.MinInstanceNumber = *req.MinInstanceNumber
		if group.DesireInstanceNumber < *req.MinInstanceNumber {
			group.DesireInstanceNumber = *req.MinInstanceNumber
		}
	}
	if req.DesireInstanceNumber != nil {
		group.DesireInstanceNumber = *req.DesireInstanceNumber
	}
	if group.MaxInstanceNumber < group.MinInstanceNumber || group.DesireInstanceNumber < group.MinInstanceNumber ||
		group.DesireInstanceNumber > group.MaxInstanceNumber {
		return errors.NewErrorRespWithHttpCode(errors.InstanceNumUpdateError, http.StatusBadRequest)
	}
	return nil
}
