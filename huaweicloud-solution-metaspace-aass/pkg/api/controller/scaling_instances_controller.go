// Copyright (c) Huawei Technologies Co., Ltd. 2022-2022. All rights reserved.

// 监控控制
package controller

import (
	"net/http"

	"github.com/beego/beego/v2/server/web"
	"scase.io/application-auto-scaling-service/pkg/api/response"
	"scase.io/application-auto-scaling-service/pkg/common"
	"scase.io/application-auto-scaling-service/pkg/api/model"
	"scase.io/application-auto-scaling-service/pkg/service"
	"scase.io/application-auto-scaling-service/pkg/utils/logger"
)

type InstanceController struct {
	web.Controller
}

func (i *InstanceController) ListInstances() {
	tLogger := logger.GetTraceLogger(i.Ctx).WithField(logger.Stage, "List instances")
	tLogger.Info("Receive an list instances request")
	ins_conf_id, err := service.GetInsConfIdByFleetId(
		i.Ctx.Input.Query(common.FleetId), tLogger)
	if err != nil {
		// 这里代表fleet创建流程还未进入到aass就已经失败，找不到fleet的记录, 直接返回空
		tLogger.Error("fleet %s get instance configuration id error: %v", 
			i.Ctx.Input.Query(common.FleetId), err)
		response.Success(i.Ctx, err.HttpCode, &model.ListInstanceResonse{})
		return
	}
	
	scaling_group_id, err := service.GetScalingGroupByInsConfId(ins_conf_id, tLogger)
	if err != nil {
		tLogger.Error("get scaling group id error: %v", err)
		response.Error(i.Ctx, err.HttpCode, err)
		return
	}
	req := &model.QueryInstanceParams{
		ScalingGroupId: scaling_group_id,
		Limit:			common.GetLimit(i.Ctx, tLogger),
		StartNumber: 	common.GetStartNumber(i.Ctx, tLogger),
		HealthState: 	i.Ctx.Input.Query("health_state"),
		LifeCycleState: i.Ctx.Input.Query("life_cycle_state"),
		ProjectId: 		i.GetString(urlParamProjectId),
	}
	resp, err := service.GenerateInstancesFromAs(tLogger, req)
	if err != nil {
		tLogger.Error("list scaling instance error: %v", err)
		response.Error(i.Ctx, err.HttpCode, err)
		return
	}
	response.Success(i.Ctx, http.StatusOK, resp)
}