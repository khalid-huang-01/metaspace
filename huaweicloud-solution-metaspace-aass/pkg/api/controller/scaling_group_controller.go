// Copyright (c) Huawei Technologies Co., Ltd. 2022-2022. All rights reserved.

// 弹性伸缩组控制
package controller

import (
	"encoding/json"
	"net/http"

	"github.com/beego/beego/v2/server/web"

	"scase.io/application-auto-scaling-service/pkg/api/errors"
	"scase.io/application-auto-scaling-service/pkg/api/model"
	"scase.io/application-auto-scaling-service/pkg/api/response"
	"scase.io/application-auto-scaling-service/pkg/api/validator"
	"scase.io/application-auto-scaling-service/pkg/common"
	"scase.io/application-auto-scaling-service/pkg/service"
	"scase.io/application-auto-scaling-service/pkg/utils/logger"
)

const (
	urlParamScalingGroupId  = ":instance_scaling_group_id"
	urlParamProjectId       = ":project_id"
	urlParamScalingPolicyId = ":scaling_policy_id"
	queryParamGroupName     = "instance_scaling_group_name"
	queryParamLimit         = "limit"
	queryParamOffset        = "offset"
)

type ScalingGroupController struct {
	web.Controller
}

// CreateScalingGroup create instance scaling group
func (c *ScalingGroupController) CreateScalingGroup() {
	tLogger := logger.GetTraceLogger(c.Ctx).WithField(logger.Stage, "create_scaling_group")
	projectId := c.GetString(urlParamProjectId)
	if errCode := validator.ErrCodeForProjectId(projectId); errCode != nil {
		response.Error(c.Ctx, http.StatusBadRequest, errors.NewErrorResp(*errCode))
		tLogger.Error("project_id verification is failed,err: %s", errCode.Msg())
		return
	}
	req := model.CreateScalingGroupReq{}
	if err := json.Unmarshal(c.Ctx.Input.RequestBody, &req); err != nil {
		response.Error(c.Ctx, http.StatusBadRequest, errors.NewErrorResp(errors.RequestParseError))
		tLogger.Error("Read req body err: %+v", err)
		return
	}
	tLogger.Info("Received scaling group create request: %s", string(c.Ctx.Input.RequestBody))
	if err := validator.Validate(&req); err != nil {
		response.Error(c.Ctx, http.StatusBadRequest,
			errors.NewErrorRespWithMessage(errors.RequestParamsError, err.Error()))
		tLogger.Error("The request parameter is err: %+v", err)
		return
	}

	group, errResp := service.CreateInstanceScalingGroup(tLogger, projectId, req)
	if errResp != nil {
		response.Error(c.Ctx, errResp.HttpCode, errResp)
		return
	}
	response.Success(c.Ctx, http.StatusCreated, group)
}

// DeleteScalingGroup delete instance scaling group
func (c *ScalingGroupController) DeleteScalingGroup() {
	tLogger := logger.GetTraceLogger(c.Ctx).WithField(logger.Stage, "delete_scaling_group")
	projectId := c.GetString(urlParamProjectId)
	if errCode := validator.ErrCodeForProjectId(projectId); errCode != nil {
		response.Error(c.Ctx, http.StatusBadRequest, errors.NewErrorResp(*errCode))
		tLogger.Error("project_id verification is failed,err: %s", errCode.Msg())
		return
	}
	groupId := c.GetString(urlParamScalingGroupId)
	tLogger.Info("Received delete request for scaling group[%s]", groupId)
	errResp := service.DeleteInstanceScalingGroup(tLogger, projectId, groupId)
	if errResp != nil {
		response.Error(c.Ctx, errResp.HttpCode, errResp)
		return
	}
	response.Success(c.Ctx, http.StatusNoContent, nil)
}

// UpdateScalingGroup update instance scaling group
func (c *ScalingGroupController) UpdateScalingGroup() {
	tLogger := logger.GetTraceLogger(c.Ctx).WithField(logger.Stage, "update_scaling_group")
	groupId := c.GetString(urlParamScalingGroupId)
	projectId := c.GetString(urlParamProjectId)
	if errCode := validator.ErrCodeForProjectId(projectId); errCode != nil {
		response.Error(c.Ctx, http.StatusBadRequest, errors.NewErrorResp(*errCode))
		tLogger.Error("project_id verification is failed,err: %s", errCode.Msg())
		return
	}
	req := model.UpdateScalingGroupReq{}
	if err := json.Unmarshal(c.Ctx.Input.RequestBody, &req); err != nil {
		response.Error(c.Ctx, http.StatusBadRequest, errors.NewErrorResp(errors.RequestParseError))
		tLogger.Error("Read req body err: %+v", err)
		return
	}
	tLogger.Info("Received update request for scaling group[%s]: %s", groupId, string(c.Ctx.Input.RequestBody))
	if err := validator.Validate(&req); err != nil {
		response.Error(c.Ctx, http.StatusBadRequest,
			errors.NewErrorRespWithMessage(errors.RequestParamsError, err.Error()))
		tLogger.Error("The request parameter is err: %+v", err)
		return
	}
	errResp := service.UpdateInstanceScalingGroup(tLogger, projectId, groupId, req)
	if errResp != nil {
		tLogger.Error("update scaling group error: %+v", errResp)
		response.Error(c.Ctx, errResp.HttpCode, errResp)
		return
	}
	response.Success(c.Ctx, http.StatusNoContent, nil)
}

// GetScalingGroup get instance scaling group detail
func (c *ScalingGroupController) GetScalingGroup() {
	tLogger := logger.GetTraceLogger(c.Ctx).WithField(logger.Stage, "get_scaling_group")
	projectId := c.GetString(urlParamProjectId)
	if errCode := validator.ErrCodeForProjectId(projectId); errCode != nil {
		response.Error(c.Ctx, http.StatusBadRequest, errors.NewErrorResp(*errCode))
		tLogger.Error("project_id verification is failed,err: %s", errCode.Msg())
		return
	}
	groupID := c.GetString(urlParamScalingGroupId)
	tLogger.Info("Received query request for scaling group[%s]", groupID)
	group, errResp := service.GetInstanceScalingGroup(tLogger, projectId, groupID)
	if errResp != nil {
		response.Error(c.Ctx, errResp.HttpCode, errResp)
		return
	}
	response.Success(c.Ctx, http.StatusOK, group)
}

// ListScalingGroup list instance scaling group
func (c *ScalingGroupController) ListScalingGroup() {
	tLogger := logger.GetTraceLogger(c.Ctx).WithField(logger.Stage, "list_scaling_group")
	projectId := c.GetString(urlParamProjectId)
	if errCode := validator.ErrCodeForProjectId(projectId); errCode != nil {
		response.Error(c.Ctx, http.StatusBadRequest, errors.NewErrorResp(*errCode))
		tLogger.Error("project_id verification is failed,err: %s", errCode.Msg())
		return
	}
	name := c.GetString(queryParamGroupName)
	if len(name) > common.MaxLengthOfParamName {
		response.Error(c.Ctx, http.StatusBadRequest, errors.NewErrorResp(errors.ScalingGroupNameError))
		tLogger.Error(errors.ScalingGroupNameError.Msg())
		return
	}
	limit, err := c.GetInt(queryParamLimit, common.MaxNumberOfParamLimit)
	if err != nil {
		response.Error(c.Ctx, http.StatusBadRequest, errors.NewErrorResp(errors.QueryParamLimitError))
		tLogger.Error("The query param limit is invalid,err: %+v", err)
		return
	}
	if limit < 0 || limit > common.MaxNumberOfParamLimit {
		response.Error(c.Ctx, http.StatusBadRequest, errors.NewErrorResp(errors.QueryParamLimitError))
		tLogger.Error("The query param limit(%d) is invalid", limit)
		return
	}
	offset, err := c.GetInt(queryParamOffset, 0)
	if err != nil {
		response.Error(c.Ctx, http.StatusBadRequest, errors.NewErrorResp(errors.QueryParamOffsetError))
		tLogger.Error("The query param offset is invalid,err: %+v", err)
		return
	}
	if offset < 0 || offset > common.MaxNumberOfParamOffset {
		response.Error(c.Ctx, http.StatusBadRequest, errors.NewErrorResp(errors.QueryParamOffsetError))
		tLogger.Error("The query param offset(%d) is invalid", offset)
		return
	}
	tLogger.Info("Received query request for scaling group list")
	list, errResp := service.ListInstanceScalingGroup(tLogger, projectId, name, limit, offset)
	if errResp != nil {
		response.Error(c.Ctx, errResp.HttpCode, errResp)
		return
	}
	response.Success(c.Ctx, http.StatusOK, list)
}

// GetInstanceConfigOfScalingGroup get instance configuration of instance scaling group
func (c *ScalingGroupController) GetInstanceConfigOfScalingGroup() {
	tLogger := logger.GetTraceLogger(c.Ctx).WithField(logger.Stage, "get_instance_config")
	groupID := c.GetString(urlParamScalingGroupId)
	tLogger.Info("Received runtime configuration query request for scaling group[%s]", groupID)
	config, errResp := service.GetInstanceConfigOfInstanceScalingGroup(tLogger, groupID)
	if errResp != nil {
		response.Error(c.Ctx, errResp.HttpCode, errResp)
		return
	}
	response.Success(c.Ctx, http.StatusOK, config)
}
