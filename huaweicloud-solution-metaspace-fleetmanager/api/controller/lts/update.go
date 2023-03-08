package lts

import (
	"encoding/json"
	"fleetmanager/api/common/log"
	"fleetmanager/api/model/lts"
	"fleetmanager/api/response"
	service "fleetmanager/api/service/lts"
	"fleetmanager/api/validator"
	"fleetmanager/logger"
	"net/http"

	"github.com/beego/beego/v2/server/web"
)

type UpdateLTSController struct {
	web.Controller
}

func (c *UpdateLTSController) UpdateAccessConfig() {
	tLogger := log.GetTraceLogger(c.Ctx).WithField(logger.Stage, "update_access_config")
	projectId := c.Ctx.Input.Param(":project_id")
	s := service.NewLtsService(c.Ctx, tLogger)
	config := lts.UpdateAccessConfigToDB{}
	if err := json.Unmarshal(c.Ctx.Input.RequestBody, &config); err != nil {
		response.InputError(c.Ctx)
		tLogger.WithField(logger.Error, err.Error()).Error("read request body error")
		return
	}
	if err := validator.Validate(&config); err != nil {
		tLogger.Error("validate request body err:%+v", err.Error())
		response.ParamsError(c.Ctx, err)
		return
	}
 	code, rsp, errNew := s.UpdateAccessConfig(projectId, config)
	if errNew != nil {
		tLogger.Error("update access config err:%s", errNew.Error())
		response.Error(c.Ctx, http.StatusInternalServerError, errNew.ErrD)
		return
	}
	if code < http.StatusOK || code >= http.StatusBadRequest {
		response.TransPort(c.Ctx, code, rsp)
	} else {
		response.Success(c.Ctx, http.StatusNoContent, nil)
	}
	tLogger.Info("update access config success")
}
