// Copyright (c) Huawei Technologies Co., Ltd. 2022-2022. All rights reserved.

// alias创建模块
package alias

import (
	"encoding/json"
	"fleetmanager/api/common/log"
	"fleetmanager/api/model/alias"
	"fleetmanager/api/response"
	service "fleetmanager/api/service/alias"
	"fleetmanager/api/validator"
	"fleetmanager/logger"
	"github.com/beego/beego/v2/server/web"
	"net/http"
)

type CreateController struct {
	web.Controller
}

// Create: 创建alias
func (c *CreateController) Create() {
	tLogger := log.GetTraceLogger(c.Ctx).WithField(logger.Stage, "create_alias")
	r := alias.CreateRequest{}
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
	s := service.NewAliasService(c.Ctx, tLogger)
	rsp, e := s.Create(&r)
	if e != nil {
		response.ServiceError(c.Ctx, e)
		tLogger.WithField(logger.Error, e.Error()).Error("create alias error")
		return
	}
	response.Success(c.Ctx, http.StatusCreated, rsp)

}
