// Copyright (c) Huawei Technologies Co., Ltd. 2022-2022. All rights reserved

// 服务端会话相关方法
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
	"codehub-g.huawei.com/videocloud/mediaprocesscenter/application-gateway/pkg/utils/statechange"
	"codehub-g.huawei.com/videocloud/mediaprocesscenter/application-gateway/pkg/utils/validator"
)

type ServerSessionControllerImpl struct {
	web.Controller
}

var ServerSessionController = &ServerSessionControllerImpl{}

// CreateServerSession 创建服务端会话
func (a *ServerSessionControllerImpl) CreateServerSession() {
	tLogger := log.GetTraceLogger(a.Ctx)

	// 1. 获取header和body
	var reqBody apis.CreateServerSessionRequest
	err := json.Unmarshal(a.Ctx.Input.RequestBody, &reqBody)
	if err != nil {
		tLogger.Errorf("[server session controller] failed to unmarshal create server session "+
			"request body for %v", err)
		Response(a.Ctx, http.StatusBadRequest, errors.NewCreateServerSessionError(
			fmt.Sprintf("can not unmarshal request body for %v", err), http.StatusBadRequest))
		return
	}

	// valid request body
	if err := validator.Validate(&reqBody); err != nil {
		tLogger.Errorf("[server session controller] invalid request body for %v", err)
		Response(a.Ctx, http.StatusBadRequest, errors.NewCreateServerSessionError(err.Error(),
			http.StatusBadRequest))
		return
	}

	tLogger.Infof("[server session controller] received an server session register request %v", reqBody)

	res, errResp := services.ServerSessionService.CreateServerSession(&reqBody, tLogger)
	if errResp != nil {
		tLogger.Errorf("[server session controller] failed to create server session %v", res.ServerSession.ID)
		Response(a.Ctx, errResp.HttpCode, errResp)
		return
	}
	Response(a.Ctx, http.StatusOK, statechange.ChangeCreateServerSessionResponseState(res))
}

// UpdateServerSession 更新服务端会话
func (a *ServerSessionControllerImpl) UpdateServerSession() {
	tLogger := log.GetTraceLogger(a.Ctx)

	sid := a.GetString(":server_session_id")

	var reqBody apis.UpdateServerSessionRequest
	err := json.Unmarshal(a.Ctx.Input.RequestBody, &reqBody)
	if err != nil {
		tLogger.Errorf("[server session control] failed to unmarshal server session %s request body "+
			"for %v", sid, err)
		Response(a.Ctx, http.StatusBadRequest, errors.NewUpdateServerSessionError(sid,
			fmt.Sprintf("can not unmarshal request for %v", err), http.StatusBadRequest))
		return
	}

	if err := validator.Validate(&reqBody); err != nil {
		tLogger.Errorf("[server session controller] invalid server session %v request body for %v", sid, err)
		Response(a.Ctx, http.StatusBadRequest, errors.NewUpdateServerSessionError(sid,
			err.Error(), http.StatusBadRequest))
		return
	}

	tLogger.Infof("[server session controller] received an server session %v request %v", sid, reqBody)
	errResp := services.ServerSessionService.UpdateServerSession(sid, &reqBody, tLogger)
	if errResp != nil {
		tLogger.Errorf("[server session controller] failed to update server session %v for %v", sid, errResp)
		Response(a.Ctx, errResp.HttpCode, errResp)
		return
	}

	Response(a.Ctx, http.StatusNoContent, nil)
}

// UpdateServerSessionState 更新服务端会话状态
func (a *ServerSessionControllerImpl) UpdateServerSessionState() {
	tLogger := log.GetTraceLogger(a.Ctx)

	sid := a.GetString(":server_session_id")

	var reqBody apis.UpdateServerSessionStateRequest
	err := json.Unmarshal(a.Ctx.Input.RequestBody, &reqBody)
	if err != nil {
		tLogger.Errorf("[server session control] failed to unmarshal server session %s request body "+
			"for %v", sid, err)
		Response(a.Ctx, http.StatusBadRequest, errors.NewUpdateServerSessionStateError(sid,
			fmt.Sprintf("can not unmarshal request for %v", err), http.StatusBadRequest))
		return
	}

	if err := validator.Validate(&reqBody); err != nil {
		tLogger.Errorf("[server session controller] invalid server session %v request body for %v", sid, err)
		Response(a.Ctx, http.StatusBadRequest, errors.NewUpdateServerSessionStateError(sid,
			err.Error(), http.StatusBadRequest))
		return
	}

	tLogger.Infof("[server session controller] received an server session id %s request %v", sid, reqBody)
	errResp := services.ServerSessionService.UpdateServerSessionState(sid, &reqBody, tLogger)
	if errResp != nil {
		tLogger.Errorf("[server session controller] failed to update server session %v for %v", sid, errResp)
		Response(a.Ctx, http.StatusBadRequest, errResp)
		return
	}

	Response(a.Ctx, http.StatusNoContent, nil)
}

// ShowServerSession 根据服务端会话ID查询服务端会话详情
func (a *ServerSessionControllerImpl) ShowServerSession() {
	tLogger := log.GetTraceLogger(a.Ctx)

	sid := a.GetString(":server_session_id")

	tLogger.Infof("[server session controller] received an server session show request, "+
		"server session id %s", sid)
	res, errResp := services.ServerSessionService.ShowServerSession(sid, tLogger)

	if errResp != nil {
		tLogger.Errorf("[server session controller] failed to show server session %v for %v", sid, errResp)
		Response(a.Ctx, errResp.HttpCode, errResp)
		return
	}

	Response(a.Ctx, http.StatusOK, statechange.ChangeShowServerSessionResponseState(res))
}

// ListServerSessions 查询服务端会话列表
func (a *ServerSessionControllerImpl) ListServerSessions() {
	tLogger := log.GetTraceLogger(a.Ctx)

	offset, err := common.CheckOffset(a.Ctx)
	if err != nil {
		tLogger.Errorf("[server session controller] failed to fetch offset %v", err)
		Response(a.Ctx, http.StatusBadRequest, errors.NewListAppProcessesError(err.Error(), http.StatusBadRequest))
		return
	}

	limit, err := common.CheckLimit(a.Ctx)
	if err != nil {
		tLogger.Errorf("[server session controller] failed to fetch limit %v", err)
		Response(a.Ctx, http.StatusBadRequest, errors.NewListAppProcessesError(err.Error(), http.StatusBadRequest))
		return
	}

	sort, err := common.CheckSort(a.Ctx)
	if err != nil {
		tLogger.Errorf("[server session controller] failed to fetch sort %v", err)
		Response(a.Ctx, http.StatusBadRequest, errors.NewListAppProcessesError(err.Error(), http.StatusBadRequest))
		return
	}

	state, err := common.CheckState(a.Ctx)
	if err != nil {
		tLogger.Errorf("[server session controller] failed to fetch state %v", err)
		Response(a.Ctx, http.StatusBadRequest, errors.NewListAppProcessesError(err.Error(), http.StatusBadRequest))
		return
	}

	fleetID := a.Ctx.Input.Query("fleet_id")
	instanceID := a.Ctx.Input.Query("instance_id")
	processID := a.Ctx.Input.Query("process_id")

	tLogger.Infof("[server session controller] received list server session request, fleet id %s,"+
		" instance id %s, process id %s, offset %d, limit %d", fleetID, instanceID, processID, offset, limit)

	res, errResp := services.ServerSessionService.ListServerSessions(fleetID, instanceID, processID, state, offset,
		limit, sort, tLogger)
	if errResp != nil {
		tLogger.Errorf("[server session controller] failed to list server session")
		Response(a.Ctx, errResp.HttpCode, errResp)
		return
	}

	Response(a.Ctx, http.StatusOK, statechange.ChangeListServerSessionResponseState(res))
}

func (a *ServerSessionControllerImpl) ListMonitorServerSessions() {
	tLogger := log.GetTraceLogger(a.Ctx)
	offset, err1 := common.CheckOffset(a.Ctx)
	limit, err2 := common.CheckLimit(a.Ctx)
	if err1 != nil || err2 != nil {
		Response(a.Ctx, http.StatusBadRequest, errors.NewListServerSessionsError("failed to fetch offset or limit", http.StatusBadRequest))
		return
	}
	state, err := common.CheckStateforServerSession(a.Ctx)
	if err != nil {
		Response(a.Ctx, http.StatusBadRequest, errors.NewListServerSessionsError(err.Error(), http.StatusBadRequest))
		return
	}
	start_time, end_time, err := common.CheckTime(a.Ctx)
	if err != nil {
		Response(a.Ctx, http.StatusBadRequest, errors.NewListServerSessionsError(err.Error(), http.StatusBadRequest))
		return
	}
	if !common.CheckIpAddress(a.Ctx.Input.Query("ip_address")) {
		Response(a.Ctx, http.StatusBadRequest, errors.NewListServerSessionsError("ip address invalid", http.StatusBadRequest))
		return
	}
	query_body := &apis.QueryServerSession {
		FleetId: 			a.Ctx.Input.Query("fleet_id"),
		InstanceID: 		a.Ctx.Input.Query("instance_id"),
		ProcessID:  		a.Ctx.Input.Query("process_id"),
		IpAddress:  		a.Ctx.Input.Query("ip_address"),
		ServerSessionID: 	a.Ctx.Input.Query("server_session_id"),
		State:				state,
		StartTime: 			start_time,
		EndTime: 			end_time,
	}
	tLogger.Infof("[monitor server session controller] received an server session query request %s", *query_body)
	totalCount, query_res, errResp := services.ListServerSessions(query_body, offset*limit, limit, tLogger)
	if errResp != nil {
		Response(a.Ctx, errResp.HttpCode, errResp)
		return
	}
	resp := &apis.ListMonitorServerSessionsResponse {
		TotalCount: 	totalCount,
		Count: 			len(*query_res),
		ServerSessions: []apis.MonitorServerSessionResponse{},
	}
	for _, ss := range *query_res {
		resp.ServerSessions = append(resp.ServerSessions, *services.GenerateServerSession(&ss))
	}
	Response(a.Ctx, http.StatusOK, resp)
}

// TerminateAllRelativeResources 终止与指定服务端会话相关的资源，包括客户端会话和自身
func (a *ServerSessionControllerImpl) TerminateAllRelativeResources() {
	tLogger := log.GetTraceLogger(a.Ctx)

	sid := a.GetString(":server_session_id")

	tLogger.Infof("[server session controller] received an terminate all relative resource for server "+
		"session %s request ", sid)
	errResp := services.ServerSessionService.TerminateAllRelativeResources(sid, tLogger)
	if errResp != nil {
		tLogger.Errorf("[server session controller] failed to exec terminate all relative resource for "+
			"server session %s, err %s", sid, errResp.ErrorMsg)
		Response(a.Ctx, errResp.HttpCode, errResp)
		return
	}

	Response(a.Ctx, http.StatusOK, nil)
}

// FetchAllRelativeResources 根据服务端会话ID获取服务端会话信息和与该服务端会话相关的所有客户端会话列表
func (a *ServerSessionControllerImpl) FetchAllRelativeResources() {
	tLogger := log.GetTraceLogger(a.Ctx)

	sid := a.GetString(":server_session_id")

	tLogger.Infof("[server session controller] received an fetch all relative resource for server "+
		"session %s request ", sid)
	resp, errResp := services.ServerSessionService.FetchAllRelativeResources(sid, tLogger)
	if errResp != nil {
		tLogger.Errorf("[server session controller] failed to exec terminate all relative resource for "+
			"server session %s, err %s", sid, errResp.ErrorMsg)
		Response(a.Ctx, errResp.HttpCode, errResp)
		return
	}

	Response(a.Ctx, http.StatusOK, resp)
}
