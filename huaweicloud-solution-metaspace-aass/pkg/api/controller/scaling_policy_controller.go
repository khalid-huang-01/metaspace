// Copyright (c) Huawei Technologies Co., Ltd. 2022-2022. All rights reserved.

// 弹性伸缩策略控制
package controller

import (
	"encoding/json"
	"net/http"

	"github.com/beego/beego/v2/server/web"

	"scase.io/application-auto-scaling-service/pkg/api/errors"
	"scase.io/application-auto-scaling-service/pkg/api/model"
	"scase.io/application-auto-scaling-service/pkg/api/response"
	"scase.io/application-auto-scaling-service/pkg/api/validator"
	"scase.io/application-auto-scaling-service/pkg/service"
	"scase.io/application-auto-scaling-service/pkg/utils/logger"
)

type ScalingPolicyController struct {
	web.Controller
}

// CreateScalingPolicy create scaling policy
func (c *ScalingPolicyController) CreateScalingPolicy() {
	tLogger := logger.GetTraceLogger(c.Ctx).WithField(logger.Stage, "create_scaling_policy")
	projectId := c.GetString(urlParamProjectId)
	if errCode := validator.ErrCodeForProjectId(projectId); errCode != nil {
		response.Error(c.Ctx, http.StatusBadRequest, errors.NewErrorResp(*errCode))
		tLogger.Error("project_id verification is failed,err: %s", errCode.Msg())
		return
	}
	req := model.CreateScalingPolicyReq{}
	if err := json.Unmarshal(c.Ctx.Input.RequestBody, &req); err != nil {
		response.Error(c.Ctx, http.StatusBadRequest, errors.NewErrorResp(errors.RequestParseError))
		tLogger.Error("Read req body err: %v", err)
		return
	}
	if err := validator.Validate(&req); err != nil {
		response.Error(c.Ctx, http.StatusBadRequest,
			errors.NewErrorRespWithMessage(errors.RequestParamsError, err.Error()))
		tLogger.Error("The request parameter is err: %+v", err)
		return
	}
	tLogger.Info("Received scaling policy create request: %+v", string(c.Ctx.Input.RequestBody))

	policy, errResp := service.CreateScalingPolicy(tLogger, projectId, req)
	if errResp != nil {
		response.Error(c.Ctx, errResp.HttpCode, errResp)
		return
	}
	response.Success(c.Ctx, http.StatusCreated, policy)
}

// DeleteScalingPolicy delete scaling policy
func (c *ScalingPolicyController) DeleteScalingPolicy() {
	tLogger := logger.GetTraceLogger(c.Ctx).WithField(logger.Stage, "delete_scaling_policy")
	projectId := c.GetString(urlParamProjectId)
	if errCode := validator.ErrCodeForProjectId(projectId); errCode != nil {
		response.Error(c.Ctx, http.StatusBadRequest, errors.NewErrorResp(*errCode))
		tLogger.Error("project_id verification is failed,err: %s", errCode.Msg())
		return
	}
	policyId := c.GetString(urlParamScalingPolicyId)
	tLogger.Info("Received delete request for scaling policy[%s]", policyId)
	errResp := service.DeleteScalingPolicy(tLogger, projectId, policyId)
	if errResp != nil {
		response.Error(c.Ctx, errResp.HttpCode, errResp)
		return
	}
	response.Success(c.Ctx, http.StatusNoContent, nil)
}

// UpdateScalingPolicy update scaling policy
func (c *ScalingPolicyController) UpdateScalingPolicy() {
	tLogger := logger.GetTraceLogger(c.Ctx).WithField(logger.Stage, "update_scaling_policy")
	policyId := c.Ctx.Input.Param(urlParamScalingPolicyId)
	projectId := c.GetString(urlParamProjectId)
	if errCode := validator.ErrCodeForProjectId(projectId); errCode != nil {
		response.Error(c.Ctx, http.StatusBadRequest, errors.NewErrorResp(*errCode))
		tLogger.Error("project_id verification is failed,err: %s", errCode.Msg())
		return
	}
	req := model.UpdateScalingPolicyReq{}
	if err := json.Unmarshal(c.Ctx.Input.RequestBody, &req); err != nil {
		response.Error(c.Ctx, http.StatusBadRequest, errors.NewErrorResp(errors.RequestParseError))
		tLogger.Error("Read req body err: %v", err)
		return
	}
	if err := validator.Validate(&req); err != nil {
		response.Error(c.Ctx, http.StatusBadRequest,
			errors.NewErrorRespWithMessage(errors.RequestParamsError, err.Error()))
		tLogger.Error("The request parameter is err: %+v", err)
		return
	}
	tLogger.Info("Received update request for scaling policy[%s]: %s", policyId, string(c.Ctx.Input.RequestBody))
	errResp := service.UpdateScalingPolicy(tLogger, projectId, policyId, req)
	if errResp != nil {
		response.Error(c.Ctx, errResp.HttpCode, errResp)
		return
	}
	response.Success(c.Ctx, http.StatusNoContent, nil)
	return
}
