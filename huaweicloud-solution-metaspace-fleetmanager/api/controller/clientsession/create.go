// Copyright (c) Huawei Technologies Co., Ltd. 2022-2022. All rights reserved.

// 客户端会话创建模块
package clientsession

import (
	"encoding/json"
	"fleetmanager/api/common/log"
	"fleetmanager/api/model/clientsession"
	"fleetmanager/api/response"
	service "fleetmanager/api/service/clientsession"
	"fleetmanager/api/validator"
	"fleetmanager/logger"
	"github.com/beego/beego/v2/server/web"
)

type CreateController struct {
	web.Controller
}

// Create: 创建client session
func (c *CreateController) Create() {
	tLogger := log.GetTraceLogger(c.Ctx).WithField(logger.Stage, "create_client_session")
	r := clientsession.CreateRequest{}
	if err := json.Unmarshal(c.Ctx.Input.RequestBody, &r); err != nil {
		response.InputError(c.Ctx)
		tLogger.WithField(logger.Error, err.Error()).Error("read create request body error")
		return
	}

	if err := validator.Validate(&r); err != nil {
		response.ParamsError(c.Ctx, err)
		tLogger.WithField(logger.Error, err.Error()).Error("parameters invalid")
		return
	}

	s := service.NewClientSessionService(c.Ctx, tLogger)
	code, rsp, e := s.Create(&r)
	if e != nil {
		response.ServiceError(c.Ctx, e)
		return
	}

	response.TransPort(c.Ctx, code, rsp)
}

// BatchCreate: 批量创建client session
func (c *CreateController) BatchCreate() {
	tLogger := log.GetTraceLogger(c.Ctx).WithField(logger.Stage, "batch_create_client_session")
	r := clientsession.BatchCreateRequest{}
	if err := json.Unmarshal(c.Ctx.Input.RequestBody, &r); err != nil {
		response.InputError(c.Ctx)
		tLogger.WithField(logger.Error, err.Error()).Error("read batch create request body error")
		return
	}

	if err := validator.Validate(&r); err != nil {
		response.ParamsError(c.Ctx, err)
		tLogger.WithField(logger.Error, err.Error()).Error("parameters invalid")
		return
	}

	s := service.NewClientSessionService(c.Ctx, tLogger)
	code, rsp, e := s.BatchCreate(&r)
	if e != nil {
		response.ServiceError(c.Ctx, e)
		return
	}

	response.TransPort(c.Ctx, code, rsp)
}