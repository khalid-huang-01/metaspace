// Copyright (c) Huawei Technologies Co., Ltd. 2022-2022. All rights reserved.

// 服务端会话创建模块
package serversession

import (
	"encoding/json"
	"fleetmanager/api/common/log"
	"fleetmanager/api/model/serversession"
	"fleetmanager/api/response"
	service "fleetmanager/api/service/serversession"
	"fleetmanager/api/validator"
	"fleetmanager/logger"
	"github.com/beego/beego/v2/server/web"
)

type CreateController struct {
	web.Controller
}

// Create: 创建服务器会话
func (c *CreateController) Create() {
	tLogger := log.GetTraceLogger(c.Ctx).WithField(logger.Stage, "create_server_session")
	r := serversession.CreateRequest{}
	if err := json.Unmarshal(c.Ctx.Input.RequestBody, &r); err != nil {
		response.InputError(c.Ctx)
		tLogger.WithField(logger.Error, err.Error()).Error("read request body error")
		return
	}

	if err := validator.Validate(&r); err != nil {
		response.ParamsError(c.Ctx, err)
		tLogger.WithField(logger.Error, err.Error()).Error("parameters invalid")
		return
	}

	s := service.NewServerSessionService(c.Ctx, tLogger)
	code, rsp, e := s.Create(&r)
	if e != nil {
		response.ServiceError(c.Ctx, e)
		tLogger.WithField(logger.Error, e.Error()).Error("create server session error")
		return
	}

	response.TransPort(c.Ctx, code, rsp)
}
