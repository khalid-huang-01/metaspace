// Copyright (c) Huawei Technologies Co., Ltd. 2022-2022. All rights reserved.

// 客户端会话相关方法
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

type ClientSessionControllerImpl struct {
	web.Controller
}

var ClientSessionController = &ClientSessionControllerImpl{}

// CreateClientSession 创建一个新的client session
func (c *ClientSessionControllerImpl) CreateClientSession() {
	tLogger := log.GetTraceLogger(c.Ctx)

	// 获取 header和body
	var reqBody apis.CreateClientSessionRequest
	err := json.Unmarshal(c.Ctx.Input.RequestBody, &reqBody)
	if err != nil {
		tLogger.Errorf("[client session controller] failed to unmarshal "+
			"create client session request body for %v", err)
		Response(c.Ctx, http.StatusBadRequest, errors.NewCreateClientSessionError(fmt.Sprintf("Can not "+
			"unmarshal request body for %v", err), http.StatusBadRequest))
		return
	}

	// valid request body
	if err := validator.Validate(&reqBody); err != nil {
		tLogger.Errorf("[client session controller] invalid request body for %v", err)
		Response(c.Ctx, http.StatusBadRequest, errors.NewCreateClientSessionError(err.Error(), http.StatusBadRequest))
		return
	}
	tLogger.Infof("[client session controller] received an client session"+
		"register request %v", reqBody)

	res, errMsg := services.ClientSessionService.CreateClientSession(&reqBody, tLogger)
	if errMsg != nil {
		tLogger.Errorf("[client session controller] failed to create client session")
		Response(c.Ctx, errMsg.HttpCode, errMsg)
		return
	}
	Response(c.Ctx, http.StatusOK, res)
}

// CreateClientSessions 批量创建clientSession
func (c *ClientSessionControllerImpl) CreateClientSessions() {
	tLogger := log.GetTraceLogger(c.Ctx)
	var reqBody apis.CreateClientSessionsRequest
	err := json.Unmarshal(c.Ctx.Input.RequestBody, &reqBody)
	if err != nil {
		tLogger.Errorf("[client session controller] failed to unmarshal "+
			"create client sessions request body for %v", err)
		Response(c.Ctx, http.StatusBadRequest, errors.NewCreateClientSessionsError(fmt.Sprintf("Can not"+
			" unmarshal request body for %v", err), http.StatusBadRequest))
		return
	}

	// valida request body
	if err := validator.Validate(&reqBody); err != nil {
		tLogger.Errorf("[client session controller] invalid request body for %v", err)
		Response(c.Ctx, http.StatusBadRequest, errors.NewCreateClientSessionError(err.Error(), http.StatusBadRequest))
		return
	}

	tLogger.Infof("[client session controller] received client sessions create request %v", reqBody)
	res, errMsg := services.ClientSessionService.CreateClientSessions(&reqBody, tLogger)
	if errMsg != nil {
		tLogger.Errorf("[client session controller] failed to create client sessions")
		Response(c.Ctx, errMsg.HttpCode, errMsg)
		return
	}
	Response(c.Ctx, http.StatusOK, res)
}

// UpdateClientSession 用来更新client session
func (c *ClientSessionControllerImpl) UpdateClientSession() {
	tLogger := log.GetTraceLogger(c.Ctx)
	cid := c.GetString(":client_session_id")

	var req apis.UpdateClientSessionRequest
	err := json.Unmarshal(c.Ctx.Input.RequestBody, &req)
	if err != nil {
		tLogger.Errorf("[client session control] failed to unmarshal "+
			"client session %s request boby for %v", cid, err)
		Response(c.Ctx, http.StatusBadRequest, errors.NewUpdateClientSessionError(cid, fmt.Sprintf("Can not "+
			"unmarshal request body for %v", err), http.StatusBadRequest))
		return
	}

	// validate
	if err := validator.Validate(&req); err != nil {
		tLogger.Errorf("[Client session controller] invalid request body for %v", err)
		Response(c.Ctx, http.StatusBadRequest, errors.NewUpdateClientSessionError(cid, err.Error(), http.StatusBadRequest))
	}

	tLogger.Infof("[client session controller] received an client session request %v", req)
	res, errMsg := services.ClientSessionService.UpdateClientSession(cid, &req, tLogger)
	if errMsg != nil {
		tLogger.Errorf("[client session controller] failed to update client session")
		Response(c.Ctx, errMsg.HttpCode, errMsg)
		return
	}
	Response(c.Ctx, http.StatusNoContent, res)
}

// UpdateClientSessionState 用来更新Client session的状态
func (c *ClientSessionControllerImpl) UpdateClientSessionState() {
	tLogger := log.GetTraceLogger(c.Ctx)

	cid := c.GetString(":client_session_id")

	var req apis.UpdateClientSessionRequestForAuxProxy
	err := json.Unmarshal(c.Ctx.Input.RequestBody, &req)
	if err != nil {
		tLogger.Errorf("[client session control] failed to unmarshal "+
			"client session %s request body for %v", cid, err)
		Response(c.Ctx, http.StatusBadRequest, errors.NewUpdateClientSessionStateError(cid, fmt.Sprintf("Can "+
			"not unmarshal request body for %v", err), http.StatusBadRequest))
		return
	}

	// valid request
	if err := validator.Validate(&req); err != nil {
		tLogger.Errorf("[client session controller] invalid request body for %v", err)
		Response(c.Ctx, http.StatusBadRequest,
			errors.NewUpdateClientSessionStateError(cid, err.Error(), http.StatusBadRequest))
		return
	}

	tLogger.Infof("[client session controller] received an client session id %s request %v", cid, req)
	res, errMsg := services.ClientSessionService.UpdateClientSessionState(cid, &req, tLogger)
	if errMsg != nil {
		tLogger.Errorf("[client session controller] failed to update client session")
		Response(c.Ctx, errMsg.HttpCode, errMsg)
		return
	}
	Response(c.Ctx, http.StatusNoContent, res)
}

// ShowClientSession 用来展示client session
func (c *ClientSessionControllerImpl) ShowClientSession() {
	tLogger := log.GetTraceLogger(c.Ctx)

	cid := c.GetString(":client_session_id")

	tLogger.Infof("[client session controller] received an client session show request, "+
		"client session id %s", cid)

	res, errMsg := services.ClientSessionService.ShowClientSession(cid, tLogger)
	if errMsg != nil {
		tLogger.Errorf("[client session controller] failed to show client session")
		Response(c.Ctx, errMsg.HttpCode, errMsg)
		return
	}
	Response(c.Ctx, http.StatusOK, res)
}

// ListClientSessions 以列表的方式列举出ClientSession列表
func (c *ClientSessionControllerImpl) ListClientSessions() {
	tLogger := log.GetTraceLogger(c.Ctx)

	offset, err := common.CheckOffset(c.Ctx)
	if err != nil {
		tLogger.Errorf("[app process controller] failed to fetch offset %v", err)
		Response(c.Ctx, http.StatusBadRequest, errors.NewListAppProcessesError(err.Error(), http.StatusBadRequest))
		return
	}

	limit, err := common.CheckLimit(c.Ctx)
	if err != nil {
		tLogger.Errorf("[app process controller] failed to fetch limit %v", err)
		Response(c.Ctx, http.StatusBadRequest, errors.NewListAppProcessesError(err.Error(), http.StatusBadRequest))
		return
	}

	sort, err := common.CheckSort(c.Ctx)
	if err != nil {
		tLogger.Errorf("[client session controller] failed to fetch sort %v", err)
		Response(c.Ctx, http.StatusBadRequest, errors.NewListAppProcessesError(err.Error(), http.StatusBadRequest))
		return
	}

	serverSessionID, err := common.CheckServerSessionID(c.Ctx)
	if err != nil {
		tLogger.Errorf("[client session controller] failed to fetch serverSessionID for %v", err)
		Response(c.Ctx, http.StatusBadRequest, errors.NewListClientSessionsError(err.Error(), http.StatusBadRequest))
	}

	tLogger.Infof("[client session controller] received list client session request, "+
		"serverSession id %s offset %d, limit %d", serverSessionID, offset, limit)

	res, errMsg := services.ClientSessionService.ListClientSession(serverSessionID, offset, limit, sort, tLogger)
	if errMsg != nil {
		tLogger.Errorf("[client session controller] failed to list client session")
		Response(c.Ctx, errMsg.HttpCode, errMsg)
		return
	}

	Response(c.Ctx, http.StatusOK, res)
}
