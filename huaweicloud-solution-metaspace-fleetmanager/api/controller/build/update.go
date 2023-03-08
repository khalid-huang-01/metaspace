// Copyright (c) Huawei Technologies Co., Ltd. 2022-2022. All rights reserved.

// 应用包更新模块
package build

import (
	"encoding/json"
	"fleetmanager/api/common/log"
	"fleetmanager/api/model/build"
	"fleetmanager/api/response"
	service "fleetmanager/api/service/build"
	"fleetmanager/api/validator"
	"fleetmanager/logger"
	"github.com/beego/beego/v2/server/web"
	"net/http"
)

// UpdateController
// @Description:
type UpdateController struct {
	web.Controller
}

// @Title Update
// @Description
// @Author wangnannan 2022-05-07 10:17:26 ${time}
func (c *UpdateController) Update() {
	tLogger := log.GetTraceLogger(c.Ctx).WithField(logger.Stage, "update_build")
	s := service.NewBuildService(c.Ctx, tLogger)
	r := build.UpdateRequest{}
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

	rep, err := s.Update(c.Ctx, r)
	if err != nil {
		response.Error(c.Ctx, http.StatusInternalServerError, err)
		tLogger.WithField(logger.Error, err.Error()).Error("update build in db error")
		return
	}

	response.Success(c.Ctx, http.StatusOK, rep)
}
