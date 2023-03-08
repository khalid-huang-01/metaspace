// Copyright (c) Huawei Technologies Co., Ltd. 2022-2022. All rights reserved.

// 应用包删除模块
package build

import (
	"fleetmanager/api/common/log"
	"fleetmanager/api/response"
	service "fleetmanager/api/service/build"
	"fleetmanager/logger"
	"github.com/beego/beego/v2/server/web"
	"net/http"
)

// DeleteController
// @Description:
type DeleteController struct {
	web.Controller
}

// @Title Delete
// @Description
// @Author wangnannan 2022-05-07 10:16:09 ${time}
func (c *DeleteController) Delete() {
	tLogger := log.GetTraceLogger(c.Ctx).WithField(logger.Stage, "delete_build")
	s := service.NewBuildService(c.Ctx, tLogger)
	if err := s.Delete(c.Ctx); err != nil {
		response.Error(c.Ctx, http.StatusInternalServerError, err)
		tLogger.WithField(logger.Error, err.Error()).Error("delete build error")
		return
	}

	response.Success(c.Ctx, http.StatusNoContent, nil)
}
