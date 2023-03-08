// Copyright (c) Huawei Technologies Co., Ltd. 2022-2022. All rights reserved.

// 实例相关服务
package service

import (
	"net/http"
	"time"

	"scase.io/application-auto-scaling-service/pkg/api/errors"

	asmodel "github.com/huaweicloud/huaweicloud-sdk-go-v3/services/as/v1/model"
	"scase.io/application-auto-scaling-service/pkg/cloudresource"
	"scase.io/application-auto-scaling-service/pkg/db"
	"scase.io/application-auto-scaling-service/pkg/common"
	"scase.io/application-auto-scaling-service/pkg/api/model"
	"scase.io/application-auto-scaling-service/pkg/utils/logger"
)

func GenerateInstancesFromAs(tLogger *logger.FMLogger, qip *model.QueryInstanceParams) (
	*model.ListInstanceResonse, *errors.ErrorResp) {
	
	res, err := ListInstancesFromAs(tLogger, qip)
	if err != nil {
		tLogger.Error("get instances from from as err: %v", err)
		return nil, err
	}
	if int(*res.TotalNumber) < qip.Limit * qip.StartNumber {
		tLogger.Error("offset * limit larger than total count")
		return nil, errors.NewErrorRespWithHttpCode(errors.RequestParamsError, http.StatusBadRequest)
	}
	tLogger.Info("list scaling instances: %v", res)
	resp := &model.ListInstanceResonse{
		TotalNumber: int(*res.TotalNumber),
		Count: len(*res.ScalingGroupInstances),
	}
	for _, instance := range *res.ScalingGroupInstances {
		resp.Instances = append(resp.Instances, *GenerateInstanceResponse(&instance))
	}
	return resp, nil
}

func ListInstancesFromAs(tLogger *logger.FMLogger, qip *model.QueryInstanceParams) (
		*asmodel.ListScalingInstancesResponse, *errors.ErrorResp) {
	resCtrl, err := cloudresource.GetResourceController(qip.ProjectId)
	if err != nil {
		tLogger.Error("get resCtrl error: %v", err)
		return nil, errors.NewErrorRespWithHttpCode(errors.ServerInternalError, http.StatusInternalServerError)
	}

	asCtrl := resCtrl.AsClient()

	request := &asmodel.ListScalingInstancesRequest{}
	request.ScalingGroupId = qip.ScalingGroupId
	if qip.LifeCycleState != "" {

		life_cycle_state := GenerateInstanceLifeCycleState(qip.LifeCycleState)
		if life_cycle_state != nil {
			request.LifeCycleState = life_cycle_state
		} else {
			return nil, errors.NewErrorRespWithHttpCodeAndMessage(errors.RequestParamsError, http.StatusBadRequest, 
					"request param life_cycle_state invalid")
		}
	}
	if qip.HealthState != "" {
		health_state := GenerateInstanceHealthState(qip.HealthState)
		if health_state != nil {
			request.HealthStatus = health_state
		} else {
			return nil, errors.NewErrorRespWithHttpCodeAndMessage(errors.RequestParamsError, http.StatusBadRequest, 
				"request param health_state invalid")
		}
	}

	limit := int32(qip.Limit)
	if limit > 0 {
		request.Limit = &limit
	}
	start_number := int32(qip.StartNumber)
	if start_number > 0 {
		request.StartNumber = &start_number
	}

	res, err := asCtrl.ListScalingInstances(request)

	if err != nil {
		tLogger.Error("list scaling instances error: %v", err)
		return nil, errors.NewErrorRespWithHttpCode(errors.ScalingGroupNotFound, http.StatusBadRequest)
	}
	return res, nil
}

func GenerateInstanceResponse(instance *asmodel.ScalingGroupInstance) (
	*model.InstanceResponse) {
	instance_id := ""
	if instance.InstanceId != nil {
		instance_id = *instance.InstanceId
	}
	res := &model.InstanceResponse{
		InstanceId: 		instance_id,
		InstanceName: 		*instance.InstanceName,
		LifeCycleState: 	instance.LifeCycleState,
		HealthStatus: 		instance.HealthStatus,
		CreatedAt: 			time.Time(*instance.CreateTime).Local(),
	}
	return res
}

func GenerateInstanceLifeCycleState(state string) (*asmodel.ListScalingInstancesRequestLifeCycleState) {
	var life_cycle_state asmodel.ListScalingInstancesRequestLifeCycleState
	if state == common.LifeCycleStateInservice {
		life_cycle_state = asmodel.GetListScalingInstancesRequestLifeCycleStateEnum().INSERVICE
	} else if state == common.LifeCycleStatePending {
		life_cycle_state = asmodel.GetListScalingInstancesRequestLifeCycleStateEnum().PENDING
	} else if state == common.LifeCycleStateRemoving {
		life_cycle_state = asmodel.GetListScalingInstancesRequestLifeCycleStateEnum().REMOVING
	} else if state == common.LifeCycleStatePendingWait {
		life_cycle_state = asmodel.GetListScalingInstancesRequestLifeCycleStateEnum().PENDING_WAIT
	} else if state == common.LifeCycleStateRemovingWait {
		life_cycle_state = asmodel.GetListScalingInstancesRequestLifeCycleStateEnum().REMOVING_WAIT
	} else if state == common.LifeCycleStateEnteringStandby {
		life_cycle_state = asmodel.GetListScalingInstancesRequestLifeCycleStateEnum().ENTERING_STANDBY
	} else if state == common.LifeCycleStateStandby {
		life_cycle_state = asmodel.GetListScalingInstancesRequestLifeCycleStateEnum().STANDBY
	} else {
		return nil
	}
	return &life_cycle_state
}

func GenerateInstanceHealthState(state string) (*asmodel.ListScalingInstancesRequestHealthStatus) {
	var health_state asmodel.ListScalingInstancesRequestHealthStatus
	if state == common.HealthStateError {
		health_state = asmodel.GetListScalingInstancesRequestHealthStatusEnum().ERROR 
	} else if state == common.HealthStateInitailizing {
		health_state = asmodel.GetListScalingInstancesRequestHealthStatusEnum().INITIALIZING
	} else if state == common.HealthStateNormal {
		health_state = asmodel.GetListScalingInstancesRequestHealthStatusEnum().NORMAL
	} else {
		return nil
	}
	return &health_state
}

func GetInsConfIdByFleetId(fleet_id string, tLogger *logger.FMLogger) (string, *errors.ErrorResp) {
	f := db.Filters{
		"fleet_id": fleet_id,
	}
	var group db.ScalingGroup
	err := f.Filter(common.TableNameScalingGroup).RelatedSel().One(&group)
	if err != nil {
		tLogger.Error("list get instance configration id error: %v", err)
		return "", errors.NewErrorRespWithHttpCode(errors.ServerInternalError, 
			http.StatusInternalServerError)
	}
	return group.ResourceId, nil
}

func GetScalingGroupByInsConfId(ins_conf_id string, tLogger *logger.FMLogger) (string, *errors.ErrorResp) {
	f := db.Filters{
		"id":          ins_conf_id,
	}
	var vm_group db.VmScalingGroup
	err := f.Filter(common.TableNameVmScalingGroup).RelatedSel().One(&vm_group)
	if err != nil {
		tLogger.Error("list get as group id error: %v", err)
		return "", errors.NewErrorRespWithHttpCode(errors.ServerInternalError, 
			http.StatusInternalServerError)
	}
	return vm_group.AsGroupId, nil
}