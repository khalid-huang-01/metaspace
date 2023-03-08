// Copyright (c) Huawei Technologies Co., Ltd. 2022-2022. All rights reserved.

// fleet创建模块
package fleet

import (
	"encoding/json"
	"fleetmanager/api/common/log"
	"fleetmanager/api/model/fleet"
	"fleetmanager/api/response"
	service "fleetmanager/api/service/fleet"
	"fleetmanager/api/validator"
	"fleetmanager/logger"
	"net/http"

	"github.com/beego/beego/v2/server/web"
)

type CreateController struct {
	web.Controller
}

// Create: 创建fleet
func (c *CreateController) Create() {
	tLogger := log.GetTraceLogger(c.Ctx).WithField(logger.Stage, "create_fleet")

	r := fleet.NewCreateRequest()
	if err := json.Unmarshal(c.Ctx.Input.RequestBody, r); err != nil {
		response.InputError(c.Ctx)
		tLogger.WithField(logger.Error, err.Error()).Error("read request body error")
		return
	}

	// 字段有效性校验
	if err := validator.Validate(r); err != nil {
		response.ParamsError(c.Ctx, err)
		tLogger.WithField(logger.Error, err.Error()).Error("parameters invalid")
		return
	}

	s := service.NewFleetService(c.Ctx, tLogger)
	rsp, e := s.Create(r)
	if e != nil {
		response.ServiceError(c.Ctx, e)
		tLogger.WithField(logger.Error, e.Error()).Error("create fleet error")
		return
	}
	response.Success(c.Ctx, http.StatusCreated, rsp)
}
