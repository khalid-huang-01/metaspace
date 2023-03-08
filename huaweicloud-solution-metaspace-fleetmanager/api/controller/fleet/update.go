// Copyright (c) Huawei Technologies Co., Ltd. 2022-2022. All rights reserved.

// fleet更新模块
package fleet

import (
	"encoding/json"
	"fleetmanager/api/common/log"
	"fleetmanager/api/model/fleet"
	"fleetmanager/api/response"
	service "fleetmanager/api/service/fleet"
	"fleetmanager/api/validator"
	"fleetmanager/logger"
	"github.com/beego/beego/v2/server/web"
	"net/http"
)

type UpdateController struct {
	web.Controller
}

// UpdateAttributes: 更新Fleet基本属性
func (c *UpdateController) UpdateAttributes() {
	tLogger := log.GetTraceLogger(c.Ctx).WithField(logger.Stage, "update_fleet_attributes")
	s := service.NewFleetService(c.Ctx, tLogger)
	code, rsp, e := s.UpdateAttribute()
	if e != nil {
		response.ServiceError(c.Ctx, e)
		tLogger.WithField(logger.Error, e.Error()).Error("update attribute error")
		return
	}

	if code < http.StatusOK || code >= http.StatusBadRequest {
		response.TransPort(c.Ctx, code, rsp)
	} else {
		response.Success(c.Ctx, http.StatusNoContent, nil)
	}
}

// UpdateInboundPermissions: 更新应用队列入站规则
func (c *UpdateController) UpdateInboundPermissions() {
	tLogger := log.GetTraceLogger(c.Ctx).WithField(logger.Stage, "update_fleet_inbound_permissions")
	r := fleet.UpdateInboundPermissionRequest{}
	if err := json.Unmarshal(c.Ctx.Input.RequestBody, &r); err != nil {
		response.InputError(c.Ctx)
		tLogger.WithField(logger.Error, err.Error()).Error("read update inbound permissions request body error")
		return
	}

	if err := validator.Validate(&r); err != nil {
		response.ParamsError(c.Ctx, err)
		tLogger.WithField(logger.Error, err.Error()).Error("parameters invalid")
		return
	}

	s := service.NewPermissionService(c.Ctx, tLogger)
	if e := s.UpdateInboundPermission(&r); e != nil {
		response.ServiceError(c.Ctx, e)
		tLogger.WithField(logger.Error, e.Error()).Error("update inbound permissions error")
		return
	}

	response.Success(c.Ctx, http.StatusNoContent, nil)
}

// UpdateRuntimeConfiguration: 更新应用队列运行时配置
func (c *UpdateController) UpdateRuntimeConfiguration() {
	tLogger := log.GetTraceLogger(c.Ctx).WithField(logger.Stage, "update_fleet_runtime_configuration")
	r := fleet.UpdateRuntimeConfigurationRequest{}
	if err := json.Unmarshal(c.Ctx.Input.RequestBody, &r); err != nil {
		response.InputError(c.Ctx)
		tLogger.WithField(logger.Error, err.Error()).Error("read update runtime configuration request body error")
		return
	}

	if err := validator.Validate(&r); err != nil {
		response.ParamsError(c.Ctx, err)
		tLogger.WithField(logger.Error, err.Error()).Error("parameters invalid")
		return
	}

	s := service.NewConfigService(c.Ctx, tLogger)
	code, rsp, e := s.UpdateRuntimeConfiguration(&r)
	if e != nil {
		response.ServiceError(c.Ctx, e)
		tLogger.WithField(logger.Error, e.Error()).Error("update runtime configuration error")
		return
	}

	if code < http.StatusOK || code >= http.StatusBadRequest {
		response.TransPort(c.Ctx, code, rsp)
	} else {
		response.Success(c.Ctx, http.StatusNoContent, nil)
	}
}

// UpdateCapacity: 更新应用队列容量
func (c *UpdateController) UpdateCapacity() {
	tLogger := log.GetTraceLogger(c.Ctx).WithField(logger.Stage, "update_fleet_capacity")
	s := service.NewFleetService(c.Ctx, tLogger)
	code, rsp, e := s.UpdateInstanceCapacity()
	if e != nil {
		response.ServiceError(c.Ctx, e)
		tLogger.WithField(logger.Error, e.Error()).Error("update instance capacity error")
		return
	}

	if code < http.StatusOK || code >= http.StatusBadRequest {
		response.TransPort(c.Ctx, code, rsp)
	} else {
		response.Success(c.Ctx, http.StatusNoContent, nil)
	}
}
