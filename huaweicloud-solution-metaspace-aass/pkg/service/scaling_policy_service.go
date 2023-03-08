// Copyright (c) Huawei Technologies Co., Ltd. 2022-2022. All rights reserved.

// 弹性伸缩策略服务
package service

import (
	"net/http"

	"github.com/beego/beego/v2/client/orm"
	"github.com/google/uuid"
	pkgerrors "github.com/pkg/errors"

	"scase.io/application-auto-scaling-service/pkg/api/errors"
	"scase.io/application-auto-scaling-service/pkg/api/model"
	"scase.io/application-auto-scaling-service/pkg/cloudresource"
	"scase.io/application-auto-scaling-service/pkg/common"
	"scase.io/application-auto-scaling-service/pkg/db"
	"scase.io/application-auto-scaling-service/pkg/metricmonitor"
	"scase.io/application-auto-scaling-service/pkg/service/taskservice"
	"scase.io/application-auto-scaling-service/pkg/utils/logger"
)

// CreateScalingPolicy create scaling policy
func CreateScalingPolicy(log *logger.FMLogger, projectId string,
	req model.CreateScalingPolicyReq) (*model.CreateScalingPolicyResp, *errors.ErrorResp) {
	group, err := db.GetScalingGroupById(projectId, *req.InstanceScalingGroupID)
	if err != nil {
		if pkgerrors.Is(err, orm.ErrNoRows) {
			log.Error("The scaling group[%s] of project[%s] is not found", *req.InstanceScalingGroupID, projectId)
			return nil, errors.NewErrorRespWithHttpCode(errors.ScalingGroupNotFound, http.StatusNotFound)
		}
		log.Error("Get scaling group[%s] from db err: %+v", *req.InstanceScalingGroupID, err)
		return nil, errors.NewErrorRespWithHttpCode(errors.ServerInternalError, http.StatusInternalServerError)
	}
	if group.State == db.ScalingGroupStateDeleting {
		log.Error(errors.ScalingGroupDeleting.Msg())
		return nil, errors.NewErrorRespWithHttpCode(errors.ScalingGroupDeleting, http.StatusBadRequest)
	}
	if db.IsTargetBasedPolicyExistInScalingGroup(projectId, *req.InstanceScalingGroupID) {
		log.Error(errors.TargetBasedPolicyExist.Msg())
		return nil, errors.NewErrorRespWithHttpCode(errors.TargetBasedPolicyExist, http.StatusBadRequest)
	}
	policyId := uuid.NewString()
	policy, err := convertCreateScalingPolicyReq(req, projectId, policyId)
	if err != nil {
		log.Error("Convert CreateScalingPolicyReq err: %+v", err)
		return nil, errors.NewErrorRespWithHttpCode(errors.ServerInternalError, http.StatusInternalServerError)
	}

	task, err := metricmonitor.GetMgmt().NewTaskForPolicy(policy.ScalingGroup.Id, policy.Id,
		*req.TargetConfiguration.MetricName, *req.TargetConfiguration.TargetValue)
	if err != nil {
		log.Error("Creat metric monitor task for policy[%s] err: %+v", policy.Id, err)
		return nil, errors.NewErrorRespWithHttpCode(errors.ServerInternalError, http.StatusInternalServerError)
	}
	err = metricmonitor.GetMgmt().AddTask(task)
	if err != nil {
		log.Error("Add metric monitor task for policy[%s] : %+v", policy.Id, err)
		return nil, errors.NewErrorRespWithHttpCode(errors.ServerInternalError, http.StatusInternalServerError)
	}

	if err := db.AddScalingPolicy(policy); err != nil {
		log.Error("Add ScalingPolicy to db err: %+v", err)
		return nil, errors.NewErrorRespWithHttpCode(errors.ServerInternalError, http.StatusInternalServerError)
	}
	return &model.CreateScalingPolicyResp{ScalingPolicyId: policyId}, nil
}

// DeleteScalingPolicy delete scaling policy
func DeleteScalingPolicy(log *logger.FMLogger, projectId, policyId string) *errors.ErrorResp {
	if exist := db.IsScalingPolicyExist(projectId, policyId); !exist {
		log.Error("The scaling policy[%s] of project[%s] is not found", policyId, projectId)
		return errors.NewErrorRespWithHttpCode(errors.ScalingPolicyNotFound, http.StatusNotFound)
	}
	policy, err := db.GetScalingPolicyById(projectId, policyId)
	if err != nil {
		log.Error("Read scaling policy[%s] of project[%s] from db err: %+v", policyId, projectId, err)
		return errors.NewErrorRespWithHttpCode(errors.ServerInternalError, http.StatusInternalServerError)
	}
	group, err := db.GetScalingGroupById(projectId, policy.ScalingGroup.Id)
	if err != nil {
		log.Error("Read policy[%s] bound scaling group[%s] of project[%s] from db err: %+v",
			policyId, policy.ScalingGroup.Id, projectId, err)
		return errors.NewErrorRespWithHttpCode(errors.ServerInternalError, http.StatusInternalServerError)
	}
	if group.EnableAutoScaling == true {
		log.Error(errors.PolicyDeleteError.Msg())
		return errors.NewErrorRespWithHttpCode(errors.PolicyDeleteError, http.StatusBadRequest)
	}
	if policy.PolicyType == common.PolicyTypeTargetBased {
		taskId := metricmonitor.GetMgmt().TaskIdForPolicy(policy.Id)
		if err = metricmonitor.GetMgmt().DeleteTask(taskId); err != nil {
			log.Error("Delete metric monitor task for policy[%s] err: %+v", policy.Id, err)
			return errors.NewErrorRespWithHttpCode(errors.ServerInternalError, http.StatusInternalServerError)
		}
	}
	if err = scalingJudgmentForDeleteScalingPolicy(log, group); err != nil {
		if pkgerrors.Is(err, common.ErrScalingGroupNotStable) {
			log.Info("The policy[%s] of group[%s] cannot be deleted because group is locked", policyId, group.Id)
			return errors.NewErrorRespWithHttpCode(errors.GroupLockUpdateNumError, http.StatusBadRequest)
		}
		log.Error("Scaling judgment for group[%s] failed, err: %+v", group.Id, err)
		return errors.NewErrorRespWithHttpCode(errors.ServerInternalError, http.StatusInternalServerError)
	}
	if err = db.DeleteScalingPolicy(projectId, policyId); err != nil {
		log.Error("Delete scaling policy[%s] of project[%s] from db err: %+v", policyId, projectId, err)
		return errors.NewErrorRespWithHttpCode(errors.ServerInternalError, http.StatusInternalServerError)
	}
	return nil
}

// UpdateScalingPolicy update scaling policy
func UpdateScalingPolicy(log *logger.FMLogger, projectId, policyId string,
	req model.UpdateScalingPolicyReq) *errors.ErrorResp {
	if exist := db.IsScalingPolicyExist(projectId, policyId); !exist {
		log.Error("The scaling policy[%s] of project[%s] is not found", policyId, projectId)
		return errors.NewErrorRespWithHttpCode(errors.ScalingPolicyNotFound, http.StatusNotFound)
	}
	policy, err := db.GetScalingPolicyById(projectId, policyId)
	if err != nil {
		log.Error("Read scaling policy[%s] of project[%s] from db err: %+v", policyId, projectId, err)
		return errors.NewErrorRespWithHttpCode(errors.ServerInternalError, http.StatusInternalServerError)
	}
	err = convertUpdateScalingPolicyReq(req, policy)
	if err != nil {
		log.Error("Convert UpdateScalingPolicyReq of policy[%s] err: %+v", policyId, err)
		return errors.NewErrorRespWithHttpCode(errors.ServerInternalError, http.StatusInternalServerError)
	}
	if policy.PolicyType == common.PolicyTypeTargetBased && req.TargetConfiguration != nil {
		taskId := metricmonitor.GetMgmt().TaskIdForPolicy(policy.Id)
		if err = metricmonitor.GetMgmt().UpdateTask(taskId, *req.TargetConfiguration.TargetValue); err != nil {
			log.Error("Update metric monitor task for policy[%s] err: %+v", policy.Id, err)
			return errors.NewErrorRespWithHttpCode(errors.ServerInternalError, http.StatusInternalServerError)
		}
	}
	if err := db.UpdateScalingPolicyById(policyId, policy); err != nil {
		log.Error("Update ScalingPolicy from db err: %+v", err)
		return errors.NewErrorRespWithHttpCode(errors.ServerInternalError, http.StatusInternalServerError)
	}
	return nil
}

// scalingJudgmentForDeleteScalingPolicy 删除策略后，判断是否触发弹性伸缩：
// 1. 伸缩组的EnableAutoScaling为false, 伸缩组仍按DesireInstanceNumber进行弹性伸缩，不触发弹性伸缩；
// 2. 伸缩组的EnableAutoScaling为true 且 伸缩组存在其他伸缩策略，伸缩组仍按伸缩策略进行弹性伸缩，不触发弹性伸缩；
// 3. 伸缩组的EnableAutoScaling为true 且 伸缩组不存在其他伸缩策略，
//    伸缩组由按伸缩策略改为按DesireInstanceNumber进行弹性伸缩，可能触发弹性伸缩。
func scalingJudgmentForDeleteScalingPolicy(log *logger.FMLogger, group *db.ScalingGroup) error {
	resCtrl, err := cloudresource.GetResourceController(group.ProjectId)
	if err != nil {
		return err
	}
	if group.EnableAutoScaling == false || len(group.ScalingPolicies) != 0 {
		return nil
	}
	vmGroup, err := db.GetVmScalingGroupById(group.ResourceId)
	if err != nil {
		return err
	}
	curNum, err := resCtrl.GetAsGroupCurrentInstanceNum(vmGroup.AsGroupId)
	log.Info("instance scaling group[%s]: curInstanceNum is %d, desireInstanceNumber is %d",
		group.Id, curNum, group.DesireInstanceNumber)
	if curNum > group.DesireInstanceNumber {
		return taskservice.StartScaleInGroupTaskForRandomVms(group.Id, group.ProjectId,
			curNum-group.DesireInstanceNumber)
	} else if curNum < group.DesireInstanceNumber {
		return taskservice.StartScaleOutGroupTask(group.Id, group.DesireInstanceNumber)
	}
	log.Info("The current number of instances is equal to desireInstanceNumber")
	return nil
}
