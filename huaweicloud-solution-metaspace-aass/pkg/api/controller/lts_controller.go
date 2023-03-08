// Copyright (c) Huawei Technologies Co., Ltd. 2022-2022. All rights reserved.

// LTS日志控制
package controller

import (
	"encoding/json"
	"net/http"

	"github.com/beego/beego/v2/server/web"
	"scase.io/application-auto-scaling-service/pkg/api/errors"
	"scase.io/application-auto-scaling-service/pkg/api/model"
	"scase.io/application-auto-scaling-service/pkg/api/response"
	"scase.io/application-auto-scaling-service/pkg/common"

	"scase.io/application-auto-scaling-service/pkg/service"
	"scase.io/application-auto-scaling-service/pkg/utils/logger"
)

const (
	queryAccessConfigId = "access_config_id"
	queryLogGroupId     = "log_group_id"
	queryTransferId     = "log_transfer_id"
	queryLogStreamId    = "log_stream_id"
)

type LTSController struct {
	web.Controller
}

func (c *LTSController) CreateLogGroup() {
	tLogger := logger.GetTraceLogger(c.Ctx).WithField(logger.Stage, "create_log_group")
	reqFromFleetManager := model.CreateLogGroup{}
	if err := json.Unmarshal(c.Ctx.Input.RequestBody, &reqFromFleetManager); err != nil {
		response.Error(c.Ctx, http.StatusBadRequest, errors.NewErrorResp(errors.RequestParseError))
		tLogger.Error("Read req body err: %+v", err)
		return
	}
	projectId := c.GetString(urlParamProjectId)
	resp, err := service.CreateLtsLogGroup(tLogger, projectId, &reqFromFleetManager)
	if err != nil {
		response.Error(c.Ctx, http.StatusBadRequest,
			errors.NewErrorRespWithMessage(errors.ServerInternalError, err.ErrMsg))
		tLogger.Error("Create log group err: %+v", err)
		return
	}
	response.Success(c.Ctx, http.StatusOK, resp)
}

func (c *LTSController) ListLogGroups() {
	tLogger := logger.GetTraceLogger(c.Ctx).WithField(logger.Stage, "list_log_group")
	projectId := c.GetString(urlParamProjectId)
	listResp, err := service.ListLtsLogGroup(tLogger, projectId)
	if err != nil {
		response.Error(c.Ctx, http.StatusInternalServerError,
			errors.NewErrorRespWithMessage(errors.ServerInternalError, err.ErrMsg))
		tLogger.Error("list log group err: %+v", err)
		return
	}
	response.Success(c.Ctx, http.StatusOK, listResp)
}

func (c *LTSController) ListLogStream() {
	tLogger := logger.GetTraceLogger(c.Ctx).WithField(logger.Stage, "list_log_stream")
	projectId := c.GetString(urlParamProjectId)
	logGroupId := c.GetString(queryLogGroupId)
	listResp, err := service.ListLogStreams(tLogger, projectId, logGroupId)
	if err != nil {
		response.Error(c.Ctx, http.StatusInternalServerError,
			errors.NewErrorRespWithMessage(errors.ServerInternalError, err.Error()))
		tLogger.Error("list log group err: %+v", err)
		return
	}
	response.Success(c.Ctx, http.StatusOK, listResp)
}

func (c *LTSController) CreateAccessConfig() {
	tLogger := logger.GetTraceLogger(c.Ctx).WithField(logger.Stage, "create_lts_config")
	reqFromFleetManager := model.CreateAccessConfig{}
	if err := json.Unmarshal(c.Ctx.Input.RequestBody, &reqFromFleetManager); err != nil {
		response.Error(c.Ctx, http.StatusBadRequest, errors.NewErrorResp(errors.RequestParseError))
		tLogger.Error("Read req body err: %+v", err)
		return
	}
	projectId := c.GetString(urlParamProjectId)
	fleetId := reqFromFleetManager.FleetId
	accessConfig, err := service.CreateLTSAccessConfig(tLogger, projectId, fleetId, &reqFromFleetManager)
	if err != nil {
		response.Error(c.Ctx, http.StatusBadRequest,
			errors.NewErrorRespWithMessage(errors.ServerInternalError, err.Error()))
		tLogger.Error("Create accessconfig err: %+v", err)
		return
	}
	response.Success(c.Ctx, http.StatusOK, accessConfig)
}

func (c *LTSController) DeleteAccessConfig() {
	tLogger := logger.GetTraceLogger(c.Ctx).WithField(logger.Stage, "delete access config")
	projectId := c.GetString(urlParamProjectId)
	accessconfigId := c.GetString(queryAccessConfigId)
	err := service.DeleteLTSAccessConfig(tLogger, projectId, accessconfigId)
	if err != nil {
		tLogger.Error("Delete access config err:%s", err.Error())
		response.Error(c.Ctx, http.StatusInternalServerError, err.Error())
		return
	}
	response.Success(c.Ctx, http.StatusNoContent, nil)
}

func (c *LTSController) ListAccessConfig() {
	tLogger := logger.GetTraceLogger(c.Ctx).WithField(logger.Stage, "list access config")
	projectId := c.GetString(urlParamProjectId)
	offset := common.GetStartNumber(c.Ctx, tLogger)
	limit := common.GetLimit(c.Ctx, tLogger)
	accessConfigList, err := service.ListLtsAccessConfigFromDB(tLogger, limit, offset, projectId)
	if err != nil {
		tLogger.Error("list access config err:%+v", err)
		response.Error(c.Ctx, http.StatusInternalServerError, err.ErrMsg)
		return
	}
	response.Success(c.Ctx, http.StatusOK, accessConfigList)
}

func (c *LTSController) QueryAccessConfig() {
	tLogger := logger.GetTraceLogger(c.Ctx).WithField(logger.Stage, "qurey log access config")
	accessConfigId := c.GetString(queryAccessConfigId)
	resp, err := service.QueryAccessConfig(tLogger, accessConfigId)
	if err != nil {
		tLogger.Error("query access config err:%s", err)
		response.Error(c.Ctx, http.StatusInternalServerError, err.ErrMsg)
		return
	}
	response.Success(c.Ctx, http.StatusOK, resp)
}

func (c *LTSController) UpdateAccessConfig() {
	tLogger := logger.GetTraceLogger(c.Ctx).WithField(logger.Stage, "update access config")
	projectId := c.GetString(urlParamProjectId)
	config := model.UpdateAccessConfigToDB{}
	if err := json.Unmarshal(c.Ctx.Input.RequestBody, &config); err != nil {
		response.Error(c.Ctx, http.StatusBadRequest, errors.NewErrorResp(errors.RequestParseError))
		tLogger.Error("Read req body err: %+v", err)
		return
	}
	resp, err := service.UpdateAccessConfigToDB(tLogger, projectId, config)
	if err != nil {
		tLogger.Error("list access config err:%s", err.ErrMsg)
		response.Error(c.Ctx, http.StatusInternalServerError, err.ErrMsg)
		return
	}
	response.Success(c.Ctx, http.StatusOK, resp)
}

func (c *LTSController) CreateLogTransfer() {
	tLogger := logger.GetTraceLogger(c.Ctx).WithField(logger.Stage, "create log transfer")
	projectId := c.GetString(urlParamProjectId)
	config := model.LogTransferReq{}
	if err := json.Unmarshal(c.Ctx.Input.RequestBody, &config); err != nil {
		response.Error(c.Ctx, http.StatusBadRequest, errors.NewErrorResp(errors.RequestParseError))
		tLogger.Error("Read req body err: %+v", err)
		return
	}
	resp, err := service.CreateLtsTransfer(tLogger, projectId, config)
	if err != nil {
		tLogger.Error("list access config err:%s", err.ErrMsg)
		response.Error(c.Ctx, http.StatusInternalServerError, err.ErrMsg)
		return
	}
	response.Success(c.Ctx, http.StatusOK, resp)
}

func (c *LTSController) ListLogTransfer() {
	tLogger := logger.GetTraceLogger(c.Ctx).WithField(logger.Stage, "list log transfer")
	projectId := c.GetString(urlParamProjectId)
	limit := common.GetLimit(c.Ctx, tLogger)
	l := c.Ctx.Input.Query("limit")
	println(l)
	offset := common.GetStartNumber(c.Ctx, tLogger)
	resp, err := service.ListLtsTransferFromDB(tLogger, limit, offset, projectId, c.GetString(queryLogStreamId))
	if err != nil {
		tLogger.Error("list access config err:%s", err.ErrMsg)
		response.Error(c.Ctx, http.StatusInternalServerError, err.ErrMsg)
		return
	}
	response.Success(c.Ctx, http.StatusOK, resp)
}

func (c *LTSController) QureyTransfer() {
	tLogger := logger.GetTraceLogger(c.Ctx).WithField(logger.Stage, "qurey log transfer")
	transferId := c.GetString(queryTransferId)
	projectId := c.GetString(urlParamProjectId)
	resp, err := service.QueryTransfer(projectId, transferId)
	if err != nil {
		tLogger.Error("query access config err:%s", err)
		response.Error(c.Ctx, http.StatusInternalServerError, err.ErrMsg)
		return
	}
	response.Success(c.Ctx, http.StatusOK, resp)
}

func (c *LTSController) DeleteLogTransfer() {
	tLogger := logger.GetTraceLogger(c.Ctx).WithField(logger.Stage, "delete transfer")
	projectId := c.GetString(urlParamProjectId)
	transferId := c.GetString(queryTransferId)
	err := service.DeleteTransfer(tLogger, projectId, transferId)
	if err != nil {
		tLogger.Error("Delete access config err:%s", err.ErrMsg)
		response.Error(c.Ctx, http.StatusInternalServerError, err.ErrMsg)
		return
	}
	response.Success(c.Ctx, http.StatusNoContent, nil)
}

func (c *LTSController) UpdateLogTransfer() {
	tLogger := logger.GetTraceLogger(c.Ctx).WithField(logger.Stage, "update transfer")
	projectId := c.GetString(urlParamProjectId)
	config := model.UpdateTransfer{}
	if err := json.Unmarshal(c.Ctx.Input.RequestBody, &config); err != nil {
		response.Error(c.Ctx, http.StatusBadRequest, errors.NewErrorResp(errors.RequestParseError))
		tLogger.Error("Read req body err: %+v", err)
		return
	}
	resp, err := service.UpdateTransfer(tLogger, projectId, config)
	if err != nil {
		tLogger.Error("update transfer err:%s", err.ErrMsg)
		response.Error(c.Ctx, http.StatusInternalServerError, err.ErrMsg)
		return
	}
	response.Success(c.Ctx, http.StatusAccepted, resp)
}

func (c *LTSController) UpdateHost() {
	tLogger := logger.GetTraceLogger(c.Ctx).WithField(logger.Stage, "update host")
	err := service.UpdateHostGroup()
	if err != nil {
		response.Error(c.Ctx, http.StatusBadRequest,
			errors.NewErrorRespWithMessage(errors.ServerInternalError, err.Error()))
		tLogger.Error("Read req body err: %+v", err)
		return
	}
	response.Success(c.Ctx, http.StatusAccepted, "update success")
}
