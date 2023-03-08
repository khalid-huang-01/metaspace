// Copyright (c) Huawei Technologies Co., Ltd. 2022-2022. All rights reserved.

// 应用包查询模块
package build

import (
	"fleetmanager/api/common/log"
	"fleetmanager/api/response"
	service "fleetmanager/api/service/build"
	"fleetmanager/logger"
	"github.com/beego/beego/v2/server/web"
	"net/http"
)

// QueryController
// @Description:
type QueryController struct {
	web.Controller
}

// @Title List
// @Description
// @Author wangnannan 2022-05-07 10:16:40 ${time}
func (c *QueryController) List() {
	tLogger := log.GetTraceLogger(c.Ctx).WithField(logger.Stage, "list_builds")
	s := service.NewBuildService(c.Ctx, tLogger)
	list, err := s.List(c.Ctx)
	if err != nil {
		response.Error(c.Ctx, http.StatusInternalServerError, err)
		tLogger.WithField(logger.Error, err).Error("query build list from db error")
		return
	}

	response.Success(c.Ctx, http.StatusOK, list)
}

// @Title Show
// @Description
// @Author wangnannan 2022-05-07 10:16:48 ${time}
func (c *QueryController) Show() {
	tLogger := log.GetTraceLogger(c.Ctx).WithField(logger.Stage, "show_build")
	s := service.NewBuildService(c.Ctx, tLogger)
	build, err := s.Show(c.Ctx)
	if err != nil {
		response.Error(c.Ctx, http.StatusInternalServerError, err)
		tLogger.WithField(logger.Error, err).Error("query build detail from db error")
		return
	}

	response.Success(c.Ctx, http.StatusOK, build)
}

// @Title GetUploadCredentials
// @Description
// @Author wangnannan 2022-05-07 10:16:53 ${time}
func (c *QueryController) GetUploadCredentials() {
	tLogger := log.GetTraceLogger(c.Ctx).WithField(logger.Stage, "get_upload_credentials")
	s := service.NewBuildService(c.Ctx, tLogger)
	uploadCredentials, err := s.GetUploadCredentials(c.Ctx)
	if err != nil {
		response.InternalError(c.Ctx, err)
		tLogger.WithField(logger.Error, err.Error()).Error("get upload credentials error")
		return
	}

	response.Success(c.Ctx, http.StatusOK, uploadCredentials)
}
