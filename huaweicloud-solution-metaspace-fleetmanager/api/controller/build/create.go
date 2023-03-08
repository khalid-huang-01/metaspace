// Copyright (c) Huawei Technologies Co., Ltd. 2022-2022. All rights reserved.

// 应用包创建模块
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

// CreateController
// @Description:
type CreateController struct {
	web.Controller
}

// @Title Create
// @Description  创建应用包
// @Author wangnannan 2022-05-07 10:15:28 ${time}

func (c *CreateController) Create() {
	tLogger := log.GetTraceLogger(c.Ctx).WithField(logger.Stage, "create_build")
	s := service.NewBuildService(c.Ctx, tLogger)
	r := build.CreateRequest{}

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

	rsp, e := s.Create(c.Ctx, r)
	if e != nil {
		response.Error(c.Ctx, http.StatusInternalServerError, e)
		tLogger.WithField(logger.Error, e.Error()).Error("create build failed")
		return
	}
	response.Success(c.Ctx, http.StatusOK, rsp)
}

func (c *CreateController) CreateByImage() {
	tLogger := log.GetTraceLogger(c.Ctx).WithField(logger.Stage, "create_build")
	s := service.NewBuildService(c.Ctx, tLogger)

	r := build.CreateByImageRequest{}

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

	rsp, e := s.CreateByImage(c.Ctx, r)
	if e != nil {
		response.Error(c.Ctx, http.StatusInternalServerError, e)
		tLogger.WithField(logger.Error, e.Error()).Error("create build failed")
		return
	}
	response.Success(c.Ctx, http.StatusOK, rsp)
}
