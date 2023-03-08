package lts

import (
	"fleetmanager/api/common/log"
	"fleetmanager/api/common/query"
	"fleetmanager/api/response"
	service "fleetmanager/api/service/lts"
	"fleetmanager/logger"
	"net/http"

	"github.com/beego/beego/v2/server/web"
)

type QureyLTSController struct {
	web.Controller
}

const (
	queryAccessConfigId = "access_config_id"
	queryLogGroupId     = "log_group_id"
	queryTransferId     = "log_transfer_id"
	queryLogStreamId    = "log_stream_id"
)

// list log group
func (c *QureyLTSController) ListLogGroup() {
	tLogger := log.GetTraceLogger(c.Ctx).WithField(logger.Stage, "list_log_group")
	s := service.NewLtsService(c.Ctx, tLogger)
	projectId := c.Ctx.Input.Param(":project_id")
	respMsg, errNew := s.ListLogGroup(projectId)
	if errNew != nil {
		tLogger.Error("list lts log group err:%+v", errNew)
		response.Error(c.Ctx, http.StatusInternalServerError, errNew.ErrD)
		return
	}
	tLogger.Info("list lts log group success")
	response.Success(c.Ctx, http.StatusOK, respMsg)
}

// list log access config
func (c *QureyLTSController) ListAccessConfig() {
	tLogger := log.GetTraceLogger(c.Ctx).WithField(logger.Stage, "list_log_access_config")
	s := service.NewLtsService(c.Ctx, tLogger)
	projectId := c.Ctx.Input.Param(":project_id")
	offset, limit, err := c.queryCheck()
	if err != nil {
		response.ParamsError(c.Ctx, err)
		return
	}
	respMsg, errNew := s.ListAccessConfig(projectId, limit, offset)
	if errNew != nil {
		tLogger.Error("list lts log access_config err:%+v", errNew)
		response.Error(c.Ctx, http.StatusInternalServerError, errNew.ErrD)
		return
	}
	tLogger.Info("list lts log access_config success")
	response.Success(c.Ctx, http.StatusOK, respMsg)
}

// query log access config
func (c *QureyLTSController) QueryAccessConfig() {
	tLogger := log.GetTraceLogger(c.Ctx).WithField(logger.Stage, "qurey access config")
	s := service.NewLtsService(c.Ctx, tLogger)
	projectId := c.Ctx.Input.Param(":project_id")
	accessConfigId := c.GetString(queryAccessConfigId)
	respMsg, errNew := s.QueryAccessConfig(projectId, accessConfigId)
	if errNew != nil {
		tLogger.Error("list lts log transfer err:%+v", errNew)
		response.Error(c.Ctx, http.StatusInternalServerError, errNew.ErrD)
		return
	}
	tLogger.Info("list lts log transfer success")
	response.Success(c.Ctx, http.StatusOK, respMsg)
}

// list log transfer
func (c *QureyLTSController) ListLogTransfer() {
	tLogger := log.GetTraceLogger(c.Ctx).WithField(logger.Stage, "list_log_transfer")
	s := service.NewLtsService(c.Ctx, tLogger)
	projectId := c.Ctx.Input.Param(":project_id")
	offset, limit, err := c.queryCheck()
	if err != nil {
		response.ParamsError(c.Ctx, err)
		return
	}
	respMsg, errNew := s.ListLogTransfer(projectId, limit, offset)
	if errNew != nil {
		tLogger.Error("list lts log transfer err:%+v", errNew.Error())
		response.Error(c.Ctx, http.StatusInternalServerError, errNew.ErrD)
		return
	}
	tLogger.Info("list lts log transfer success")
	response.Success(c.Ctx, http.StatusOK, respMsg)
}

// query log transfer
func (c *QureyLTSController) QueryLogTransfer() {
	tLogger := log.GetTraceLogger(c.Ctx).WithField(logger.Stage, "qurey access config")
	s := service.NewLtsService(c.Ctx, tLogger)
	projectId := c.Ctx.Input.Param(":project_id")
	LogStreamId := c.GetString(queryLogStreamId)
	respMsg, errNew := s.QueryLogTransfer(projectId, LogStreamId)
	// 如果查询失败，返回空
	if errNew != nil {
		tLogger.Error("list lts log transfer err:%s", errNew.ErrD)
		response.Success(c.Ctx, http.StatusNoContent, respMsg)
		return
	}
	tLogger.Info("list lts log transfer success")
	response.Success(c.Ctx, http.StatusOK, respMsg)
}

// queryCheck 校验分页参数
func (c *QureyLTSController) queryCheck() (int, int, error) {
	offset, err := query.CheckOffset(c.Ctx)
	if err != nil {
		return 0, 0, err
	}
	limit, err := query.CheckLimit(c.Ctx)
	if err != nil {
		return 0, 0, err
	}
	return offset, limit, nil
}
