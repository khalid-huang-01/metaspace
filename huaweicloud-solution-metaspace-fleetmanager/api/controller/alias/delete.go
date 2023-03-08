// Copyright (c) Huawei Technologies Co., Ltd. 2022-2022. All rights reserved.

// alias删除模块
package alias

import (
	"fleetmanager/api/common/log"
	"fleetmanager/api/response"
	service "fleetmanager/api/service/alias"
	"fleetmanager/logger"
	"github.com/beego/beego/v2/server/web"
	"net/http"
)

type DeleteController struct {
	web.Controller
}

// Delete: 删除Alias,RoutingStrategy
func (c *DeleteController) Delete() {
	tLogger := log.GetTraceLogger(c.Ctx).WithField(logger.Stage, "delete_alias")
	s := service.NewAliasService(c.Ctx, tLogger)
	if e := s.Delete(c.Ctx); e != nil {
		response.ServiceError(c.Ctx, e)
		tLogger.WithField(logger.Error, e.Error()).Error("delete alias in db error")
		return
	}
	response.Success(c.Ctx, http.StatusNoContent, nil)
}
