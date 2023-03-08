// Copyright (c) Huawei Technologies Co., Ltd. 2022-2022. All rights reserved.

// fleet删除模块
package fleet

import (
	"fleetmanager/api/common/log"
	"fleetmanager/api/response"
	service "fleetmanager/api/service/fleet"
	"fleetmanager/logger"
	"github.com/beego/beego/v2/server/web"
	"net/http"
)

type DeleteController struct {
	web.Controller
}

// Delete: 删除Fleet
func (c *DeleteController) Delete() {
	tLogger := log.GetTraceLogger(c.Ctx).WithField(logger.Stage, "delete_fleet")

	s := service.NewFleetService(c.Ctx, tLogger)
	if e := s.Delete(); e != nil {
		response.ServiceError(c.Ctx, e)
		tLogger.WithField(logger.Error, e.Error()).Error("delete fleet in db error")
		return
	}

	response.Success(c.Ctx, http.StatusNoContent, nil)
}
