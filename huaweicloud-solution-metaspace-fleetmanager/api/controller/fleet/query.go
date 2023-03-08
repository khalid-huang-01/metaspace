// Copyright (c) Huawei Technologies Co., Ltd. 2022-2022. All rights reserved.

// fleet查询模块
package fleet

import (
	"fleetmanager/api/common/log"
	"fleetmanager/api/common/query"
	"fleetmanager/api/response"
	service "fleetmanager/api/service/fleet"
	"fleetmanager/logger"
	"net/http"

	"github.com/beego/beego/v2/server/web"
)

type QueryController struct {
	web.Controller
}

func (c *QueryController) queryCheck() (int, int, error) {
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

// ListFleets: 查询Fleet列表
func (c *QueryController) ListFleets() {
	tLogger := log.GetTraceLogger(c.Ctx).WithField(logger.Stage, "list_fleets")
	s := service.NewFleetService(c.Ctx, tLogger)

	offset, limit, err := c.queryCheck()
	if err != nil {
		response.ParamsError(c.Ctx, err)
		return
	}

	list, e := s.List(offset, limit)
	if e != nil {
		response.ServiceError(c.Ctx, e)
		tLogger.WithField(logger.Error, e.Error()).Error("query fleet list from db error")
		return
	}

	response.Success(c.Ctx, http.StatusOK, list)
}

// ListMonitorFleets: 从监控维度查看fleet列表
func (c *QueryController) ListMonitorFleets() {
	tLogger := log.GetTraceLogger(c.Ctx).WithField(logger.Stage, "list_monitor_fleets")
	s := service.NewFleetService(c.Ctx, tLogger)
	offset, limit, errCheck := c.queryCheck()
	if errCheck != nil {
		response.ParamsError(c.Ctx, errCheck)
		return
	}
	rsp, err := s.ListMonitorFleets(offset, limit)
	if err != nil {
		response.ServiceError(c.Ctx, err)
		tLogger.WithField(logger.Error, err.Error()).Error("list monitor fleet info error")
		return
	}
	response.Success(c.Ctx, http.StatusOK, rsp)
}

// ShowMonitorFleet: 从监控维度查看fleet信息
func (c *QueryController) ShowMonitorFleet() {
	tLogger := log.GetTraceLogger(c.Ctx).WithField(logger.Stage, "show_monitor_fleet")
	s := service.NewFleetService(c.Ctx, tLogger)
	rsp, err := s.ShowMonitorFleet()
	if err != nil {
		response.ServiceError(c.Ctx, err)
		tLogger.WithField(logger.Error, err.Error()).Error("query monitor fleet info error")
		return
	}
	response.Success(c.Ctx, http.StatusOK, rsp)
}

// ShowFleet: 查询Fleet详情
func (c *QueryController) ShowFleet() {
	tLogger := log.GetTraceLogger(c.Ctx).WithField(logger.Stage, "show_fleet")
	s := service.NewFleetService(c.Ctx, tLogger)
	rsp, e := s.Show()
	if e != nil {
		response.ServiceError(c.Ctx, e)
		tLogger.WithField(logger.Error, e.Error()).Error("query fleet info from db error")
		return
	}

	response.Success(c.Ctx, http.StatusOK, rsp)
}

func (c *QueryController) ListInstances() {
	tLogger := log.GetTraceLogger(c.Ctx).WithField(logger.Stage, "list_monitor_instances")
	s := service.NewFleetService(c.Ctx, tLogger)
	rsp, e := s.ListMonitorInstances()
	if e != nil {
		response.ServiceError(c.Ctx, e)
		tLogger.WithField(logger.Error, e.Error()).Error("list monitor instances error")
		return
	}

	response.Success(c.Ctx, http.StatusOK, rsp)
}

func (c *QueryController) ListAppProcesses() {
	tLogger := log.GetTraceLogger(c.Ctx).WithField(logger.Stage, "list_monitor_app_process")
	s := service.NewFleetService(c.Ctx, tLogger)
	rsp, e := s.ListMonitorAppProcesses()
	if e != nil {
		response.ServiceError(c.Ctx, e)
		tLogger.WithField(logger.Error, e.Error()).Error("list minitor app process error")
		return
	}

	response.Success(c.Ctx, http.StatusOK, rsp)
}

func (c *QueryController) ListServerSessions() {
	tLogger := log.GetTraceLogger(c.Ctx).WithField(logger.Stage, "list_monitor_server_sessions")
	s := service.NewFleetService(c.Ctx, tLogger)
	rsp, e := s.ListMonitorServerSessions()
	if e != nil {
		response.ServiceError(c.Ctx, e)
		tLogger.WithField(logger.Error, e.Error()).Error("list minitor server session error")
		return
	}

	response.Success(c.Ctx, http.StatusOK, rsp)
}
// ShowInboundPermissions: 查询Fleet的入站规则
func (c *QueryController) ShowInboundPermissions() {
	tLogger := log.GetTraceLogger(c.Ctx).WithField(logger.Stage, "show_fleet_inbound_permissions")

	s := service.NewPermissionService(c.Ctx, tLogger)
	fleetInboundPermissions, e := s.ShowInboundPermissions()
	if e != nil {
		response.ServiceError(c.Ctx, e)
		tLogger.WithField(logger.Error, e.Error()).Error("query fleet inbound permissions from db error")
		return
	}

	response.Success(c.Ctx, http.StatusOK, fleetInboundPermissions)
}

// ShowRuntimeConfiguration: 查询Fleet的运行时配置
func (c *QueryController) ShowRuntimeConfiguration() {
	tLogger := log.GetTraceLogger(c.Ctx).WithField(logger.Stage, "show_fleet_runtime_configuration")

	s := service.NewConfigService(c.Ctx, tLogger)
	fleetRuntimeConfiguration, err := s.ShowRuntimeConfiguration()
	if err != nil {
		response.ServiceError(c.Ctx, err)
		tLogger.WithField(logger.Error, err.Error()).Error("query fleet runtime configuration from db error")
		return
	}

	response.Success(c.Ctx, http.StatusOK, fleetRuntimeConfiguration)
}

// ListFleetEvents: 查询Fleet事件列表
func (c *QueryController) ListFleetEvents() {
	tLogger := log.GetTraceLogger(c.Ctx).WithField(logger.Stage, "list_fleet_events")

	s := service.NewEventService(c.Ctx, tLogger)
	list, e := s.ListFleetEvents()
	if e != nil {
		response.ServiceError(c.Ctx, e)
		tLogger.WithField(logger.Error, e.Error()).Error("query fleet events from db error")
		return
	}

	response.Success(c.Ctx, http.StatusOK, list)
}

// ShowInstanceCapacity: 查询Fleet的容量详情
func (c *QueryController) ShowInstanceCapacity() {
	tLogger := log.GetTraceLogger(c.Ctx).WithField(logger.Stage, "show_instance_capacity")

	s := service.NewFleetService(c.Ctx, tLogger)

	rsp, err := s.GetInstanceCapacity()
	if err != nil {
		response.ServiceError(c.Ctx, err)
		tLogger.WithField(logger.Error, err.Error()).Error("get instance capacity failed")
		return
	}

	response.Success(c.Ctx, http.StatusOK, rsp)
}

// ShowProcessCounts: 查询Fleet的进程统计信息
func (c *QueryController) ShowProcessCounts() {
	tLogger := log.GetTraceLogger(c.Ctx).WithField(logger.Stage, "show_process_counts")

	s := service.NewFleetService(c.Ctx, tLogger)

	code, rsp, e := s.ShowProcessCounts()
	if e != nil {
		response.ServiceError(c.Ctx, e)
		return
	}

	response.TransPort(c.Ctx, code, rsp)
}
