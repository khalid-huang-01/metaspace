// Copyright (c) Huawei Technologies Co., Ltd. 2022-2022. All rights reserved.

// alias更新模块
package alias

import (
	"fleetmanager/api/common/log"
	"fleetmanager/api/response"
	service "fleetmanager/api/service/alias"
	"fleetmanager/logger"
	"github.com/beego/beego/v2/server/web"
	"net/http"
)

type UpdateController struct {
	web.Controller
}

// Update: 更新Alias信息
func (c *UpdateController) Update() {
	tLogger := log.GetTraceLogger(c.Ctx).WithField(logger.Stage, "update_alias")
	s := service.NewAliasService(c.Ctx, tLogger)
	_, e := s.Update()
	if e != nil {
		response.ServiceError(c.Ctx, e)
		tLogger.WithField(logger.Error, e.Error()).Error("update alias error")
		return
	}
	response.Success(c.Ctx, http.StatusNoContent, nil)
}
