// Copyright (c) Huawei Technologies Co., Ltd. 2022-2022. All rights reserved.

// 应用包创建模块
package build

import (
	"fleetmanager/api/common/log"
	"fleetmanager/api/response"
	service "fleetmanager/api/service/build"
	"fleetmanager/logger"
	"github.com/beego/beego/v2/server/web"
	"net/http"
)

// UploadController
// @Description:
type UploadController struct {
	web.Controller
}

func (c *UploadController) Upload() {
	tLogger := log.GetTraceLogger(c.Ctx).WithField(logger.Stage, "upload_build")
	s := service.NewBuildService(c.Ctx, tLogger)

	file, information, err := c.GetFile("file")
	if err != nil {
		response.InputError(c.Ctx)
		tLogger.WithField(logger.Error, err.Error()).Error("read file error: %v", err)
		return
	}

	region := c.Ctx.Input.Query("region")
	if region == "" {
		response.InputError(c.Ctx)
		tLogger.Error("get region error")
		return
	}

	bucketName := c.Ctx.Input.Query("bucket_name")
	if bucketName == "" {
		response.InputError(c.Ctx)
		tLogger.Error("get bucket name error")
		return
	}

	rsp, obsError := s.Upload(c.Ctx, file, information, bucketName, region)
	if obsError != nil {
		response.Error(c.Ctx, http.StatusInternalServerError, obsError)
		tLogger.WithField(logger.Error, obsError.Error()).Error("Upload build error")
		return
	}

	err = file.Close()
	if err != nil {
		tLogger.WithField(logger.Error, err.Error()).Error("file close error")
	}

	response.Success(c.Ctx, http.StatusOK, rsp)
}
