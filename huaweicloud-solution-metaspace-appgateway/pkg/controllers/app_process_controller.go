// Copyright (c) Huawei Technologies Co., Ltd. 2022-2022. All rights reserved.

// 应用进程相关方法
package controllers

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/beego/beego/v2/server/web"

	"codehub-g.huawei.com/videocloud/mediaprocesscenter/application-gateway/pkg/apis"
	"codehub-g.huawei.com/videocloud/mediaprocesscenter/application-gateway/pkg/common"
	"codehub-g.huawei.com/videocloud/mediaprocesscenter/application-gateway/pkg/services"
	"codehub-g.huawei.com/videocloud/mediaprocesscenter/application-gateway/pkg/utils/errors"
	"codehub-g.huawei.com/videocloud/mediaprocesscenter/application-gateway/pkg/utils/log"
	"codehub-g.huawei.com/videocloud/mediaprocesscenter/application-gateway/pkg/utils/validator"
)

type AppProcessControllerImpl struct {
	web.Controller
}

var AppProcessController = &AppProcessControllerImpl{}

// RegisterAppProcess register an app process
func (a *AppProcessControllerImpl) RegisterAppProcess() {
	tLogger := log.GetTraceLogger(a.Ctx)

	// 1. 拿到请求的内容，包括header和body
	var registerAppProcessRequestBody apis.RegisterAppProcessRequest
	err := json.Unmarshal(a.Ctx.Input.RequestBody, &registerAppProcessRequestBody)
	if err != nil {
		tLogger.Errorf("[app process controller] failed to unmarshal register process request body for %v", err)
		Response(a.Ctx, http.StatusBadRequest, errors.NewCreateAppProcessError(fmt.Sprintf("can not unmarshal request "+
			"body %v for %v", registerAppProcessRequestBody, err), http.StatusBadRequest))
		return
	}

	if err := validator.Validate(&registerAppProcessRequestBody); err != nil {
		tLogger.Errorf("[app process controller] invalid request body for %v", err)
		Response(a.Ctx, http.StatusBadRequest, errors.NewCreateAppProcessError(err.Error(), http.StatusBadRequest))
		return
	}

	tLogger.Infof("[app process controller] received an app process register request %v", registerAppProcessRequestBody)

	res, errMsg := services.AppProcessService.CreateAppProcessService(&registerAppProcessRequestBody, tLogger)
	if errMsg != nil {
		tLogger.Errorf("[app process controller] failed to create app process id %v", res.AppProcess.ID)
		Response(a.Ctx, errMsg.HttpCode, errMsg)
		return
	}

	Response(a.Ctx, http.StatusOK, res)
}

// UpdateAppProcess update an app process
func (a *AppProcessControllerImpl) UpdateAppProcess() {
	tLogger := log.GetTraceLogger(a.Ctx)

	pID := a.GetString(":process_id")

	var updateAppProcessRequest apis.UpdateAppProcessRequest
	err := json.Unmarshal(a.Ctx.Input.RequestBody, &updateAppProcessRequest)
	if err != nil {
		tLogger.Errorf(
			"[app process controller] failed to unmarshal update app process %s request body for %v", pID, err)
		Response(a.Ctx, http.StatusBadRequest, errors.NewUpdateAppProcessesError(fmt.Sprintf(
			"can not unmarshal request body %v for %v", updateAppProcessRequest, err), http.StatusBadRequest))
		return
	}

	if err := validator.Validate(&updateAppProcessRequest); err != nil {
		tLogger.Errorf("[app process controller] invalid process %v request body for %v", pID, err)
		Response(a.Ctx, http.StatusBadRequest, errors.NewUpdateAppProcessesError(err.Error(), http.StatusBadRequest))
		return
	}

	tLogger.Infof("[app process controller] received an app process %v update request %v", pID, updateAppProcessRequest)

	_, errMsg := services.AppProcessService.UpdateAppProcessService(pID, &updateAppProcessRequest, tLogger)
	if errMsg != nil {
		tLogger.Errorf("[app process controller] failed to update app process %v", pID)
		Response(a.Ctx, errMsg.HttpCode, errMsg)
		return
	}

	Response(a.Ctx, http.StatusNoContent, nil)
}

// UpdateAppProcessState update an app process state
func (a *AppProcessControllerImpl) UpdateAppProcessState() {
	tLogger := log.GetTraceLogger(a.Ctx)

	pID := a.GetString(":process_id")

	var updateAppProcessStateRequest apis.UpdateAppProcessStateRequest
	err := json.Unmarshal(a.Ctx.Input.RequestBody, &updateAppProcessStateRequest)
	if err != nil {
		tLogger.Errorf(
			"[app process controller] failed to unmarshal update app process %s request body for %v", pID, err)
		Response(a.Ctx, http.StatusBadRequest, errors.NewUpdateAppProcessStateError(fmt.Sprintf(
			"can not unmarshal request body %v for %v", updateAppProcessStateRequest, err), http.StatusBadRequest))
		return
	}

	if err := validator.Validate(&updateAppProcessStateRequest); err != nil {
		tLogger.Errorf("[app process controller] invalid request body for %v", err)
		Response(a.Ctx, http.StatusBadRequest, errors.NewUpdateAppProcessStateError(err.Error(), http.StatusBadRequest))
		return
	}

	tLogger.Debugf("[app process controller] received an app process update request %v", updateAppProcessStateRequest)

	_, errMsg := services.AppProcessService.UpdateAppProcessStateService(pID, &updateAppProcessStateRequest, tLogger)
	if errMsg != nil {
		tLogger.Errorf("[app process controller] failed to update app process")
		Response(a.Ctx, errMsg.HttpCode, errMsg)
		return
	}

	Response(a.Ctx, http.StatusNoContent, nil)
}

// ShowAppProcess show an app process
func (a *AppProcessControllerImpl) ShowAppProcess() {
	tLogger := log.GetTraceLogger(a.Ctx)

	pID := a.GetString(":process_id")

	tLogger.Infof("[app process controller] received an app process show request, process id %s", pID)

	res, errMsg := services.AppProcessService.ShowAppProcessService(pID, tLogger)
	if errMsg != nil {
		tLogger.Errorf("[app process controller] failed to update app process")
		Response(a.Ctx, errMsg.HttpCode, errMsg)
		return
	}

	Response(a.Ctx, http.StatusOK, res)
}

// ListAppProcesses list app processes
func (a *AppProcessControllerImpl) ListAppProcesses() {
	tLogger := log.GetTraceLogger(a.Ctx)

	offset, err := common.CheckOffset(a.Ctx)
	if err != nil {
		tLogger.Errorf("[app process controller] failed to fetch offset %v", err)
		Response(a.Ctx, http.StatusBadRequest, errors.NewListAppProcessesError(err.Error(), http.StatusBadRequest))
		return
	}

	limit, err := common.CheckLimit(a.Ctx)
	if err != nil {
		tLogger.Errorf("[app process controller] failed to fetch limit %v", err)
		Response(a.Ctx, http.StatusBadRequest, errors.NewListAppProcessesError(err.Error(), http.StatusBadRequest))
		return
	}

	sort, err := common.CheckSort(a.Ctx)
	if err != nil {
		tLogger.Errorf("[app process controller] failed to fetch sort %v", err)
		Response(a.Ctx, http.StatusBadRequest, errors.NewListAppProcessesError(err.Error(), http.StatusBadRequest))
		return
	}

	// fleetID和instanceID是可选的
	fleetID := a.Ctx.Input.Query("fleet_id")
	instanceID := a.Ctx.Input.Query("instance_id")

	tLogger.Infof("[app process controller] received list app processes request, "+
		"fleet id %s, instance id %s, offset %d, limit %d", fleetID, instanceID, offset, limit)

	res, errMsg := services.AppProcessService.ListAppProcessesService(fleetID, instanceID, offset, limit, sort, tLogger)
	if errMsg != nil {
		tLogger.Errorf("[app process controller] failed to list app processes")
		Response(a.Ctx, errMsg.HttpCode, errMsg)
		return
	}

	Response(a.Ctx, http.StatusOK, res)
}

func (a *AppProcessControllerImpl) ListMonitorAppProcesses() {
	tLogger := log.GetTraceLogger(a.Ctx)
	offset, err1 := common.CheckOffset(a.Ctx)
	limit, err2  := common.CheckLimit(a.Ctx)
	if err1 != nil || err2 != nil {
		Response(a.Ctx, http.StatusBadRequest, errors.NewListAppProcessesError("failed to fetch offset or limit", http.StatusBadRequest))
		return
	}
	state, err  := common.CheckStateForAppProcess(a.Ctx)
	if err != nil {
		Response(a.Ctx, http.StatusBadRequest, errors.NewListAppProcessesError(err.Error(), http.StatusBadRequest))
		return
	}
	duration_start, duration_end, err := common.CheckDuration(a.Ctx)
	if err != nil {
		Response(a.Ctx, http.StatusBadRequest, errors.NewListAppProcessesError(err.Error(), http.StatusBadRequest))
		return
	}
	if !common.CheckIpAddress(a.Ctx.Input.Query("ip_address")) {
		Response(a.Ctx, http.StatusBadRequest, errors.NewListAppProcessesError("ip address invalid", http.StatusBadRequest))
		return
	}
	query_body := &apis.QueryMonitorAppProcess{
		State: 			state,
		FleetID: 		a.Ctx.Input.Query("fleet_id"),
		ProcessID:  	a.Ctx.Input.Query("process_id"),
		InstanceID:  	a.Ctx.Input.Query("instance_id"),
		IpAddress:  	a.Ctx.Input.Query("ip_address"),
		GtServerSessionNum: duration_start,
		LtServerSessionNum: duration_end,
	}
	tLogger.Infof("[monitor app process controller] received an app process query request %s", *query_body)
	totalCount, query_res, errResp := services.AppProcessService.ListMonitorAppProcesses(query_body, offset*limit, limit, tLogger)
	if errResp != nil {
		Response(a.Ctx, errResp.HttpCode, errResp)
		return
	}
	resp := &apis.ListMonitorAppProcessResponse {
		TotalCount: 	totalCount,
		Count: 			len(*query_res),
		Processes: 		[]apis.MonitorAppProcessResponse{},
	}
	for _, ap := range *query_res {
		resp.Processes = append(resp.Processes, *services.AppProcessService.GenerateMonitorAppProcess(&ap))
	}
	Response(a.Ctx, http.StatusOK, resp)
}

// DeleteAppProcess delete an app process
func (a *AppProcessControllerImpl) DeleteAppProcess() {
	tLogger := log.GetTraceLogger(a.Ctx)

	pID := a.GetString(":process_id")

	tLogger.Infof("[app process controller] received an app process delete request, process id %s", pID)

	errMsg := services.AppProcessService.DeleteAppProcessService(pID, tLogger)
	if errMsg != nil {
		tLogger.Errorf("[app process controller] failed to delete app process")
		Response(a.Ctx, errMsg.HttpCode, errMsg)
		return
	}

	Response(a.Ctx, http.StatusOK, nil)
}

// ShowAppProcessCounts show app process counts
func (a *AppProcessControllerImpl) ShowAppProcessCounts() {
	tLogger := log.GetTraceLogger(a.Ctx)

	fleetID := a.Ctx.Input.Query("fleet_id")

	tLogger.Infof("[app process controller] received a show app process count request")

	res, errMsg := services.AppProcessService.ShowAppProcessCounts(fleetID, tLogger)
	if errMsg != nil {
		tLogger.Errorf("[app process controller] failed to show app process counts")
		Response(a.Ctx, errMsg.HttpCode, errMsg)
		return
	}

	Response(a.Ctx, http.StatusOK, res)
}
