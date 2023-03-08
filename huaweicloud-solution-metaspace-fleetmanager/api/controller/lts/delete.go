package lts

import (
	"fleetmanager/api/common/log"
	"fleetmanager/api/response"
	service "fleetmanager/api/service/lts"
	"fleetmanager/logger"
	"net/http"

	"github.com/beego/beego/v2/server/web"
)

type DeleteLTSController struct {
	web.Controller
}

// 删除日志转储
func (c *DeleteLTSController) DeleteLtsTransfer() {
	tLogger := log.GetTraceLogger(c.Ctx).WithField(logger.Stage, "delete_log_transfer")
	s := service.NewLtsService(c.Ctx, tLogger)
	projectId := c.Ctx.Input.Param(":project_id")
	transferId := c.GetString("log_transfer_id")
	code,rsp,errNew := s.DeleteTransfer(projectId, transferId)
	if errNew != nil {
		tLogger.Error("delete log transfer err:%s", errNew.Error())
		response.Error(c.Ctx, http.StatusBadRequest, errNew.ErrD)
		return
	}
	if code < http.StatusOK || code >= http.StatusBadRequest {
		response.TransPort(c.Ctx, code, rsp)
	} else {
		response.Success(c.Ctx, http.StatusNoContent, nil)
	}
	tLogger.Info("delete log transferg success")
	response.Success(c.Ctx, http.StatusNoContent, nil)
}

// 删除日志接入
func (c *DeleteLTSController) DeleteLtsAccessConfig() {
	tLogger := log.GetTraceLogger(c.Ctx).WithField(logger.Stage, "delete_access_config")
	s := service.NewLtsService(c.Ctx, tLogger)
	projectId := c.Ctx.Input.Param(":project_id")
	accessConfigId := c.GetString("access_config_id")
	code,rsp,errNew := s.DeleteAccessConfig(projectId, accessConfigId)
	if errNew != nil {
		tLogger.Error("delete log access config err:%s", errNew.Error())
		response.Error(c.Ctx, http.StatusBadRequest, errNew.ErrD)
		return
	}
	if code < http.StatusOK || code >= http.StatusBadRequest {
		response.TransPort(c.Ctx, code, rsp)
	} else {
		response.Success(c.Ctx, http.StatusNoContent, nil)
	}
	tLogger.Info("delete log access config")
	response.Success(c.Ctx, http.StatusNoContent, nil)
}
