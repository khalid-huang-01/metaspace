// Copyright (c) Huawei Technologies Co., Ltd. 2022-2022. All rights reserved.

// 服务端会话查询服务
package serversession

import (
	"encoding/json"
	"fleetmanager/api/errors"
	"fleetmanager/api/model/serversession"
	"fleetmanager/api/params"
	"fleetmanager/api/service/constants"
	"fleetmanager/client"
	"fleetmanager/utils"
	"fleetmanager/logger"
	"fmt"
	"net/http"
)

var (
	clientGetServiceEndpoint = client.GetServiceEndpoint
	clientNewRequest         = client.NewRequest
)

func (s *Service) forwardShowToAPPGW(region string) (code int, rsp []byte, err error) {
	sessionId := s.Ctx.Input.Param(params.ServerSessionId)
	url := client.GetServiceEndpoint(client.ServiceNameAPPGW, region) +
		fmt.Sprintf(constants.ServerSessionUrlPattern, sessionId)
	req := client.NewRequest(client.ServiceNameAPPGW, url, http.MethodGet, nil)
	req.SetHeader(map[string]string{
		logger.RequestId: fmt.Sprintf("%s", s.Ctx.Input.GetData(logger.RequestId)),
	})
	return req.DoRequest()
}

// Show 查询服务端会话详情
func (s *Service) Show() (code int, rsp []byte, e *errors.CodedError) {
	if err := s.SetFleetByServerSessionId(s.Ctx.Input.Param(params.ServerSessionId)); err != nil {
		s.Logger.Error("get fleet in show server session error, serverSessionId:%s",
			s.Ctx.Input.Param(params.ServerSessionId))
		return 0, nil, err
	}

	code, rsp, err := s.forwardShowToAPPGW(s.Fleet.Region)
	s.Logger.Info("show server session to appgw, code: %d, rsp: %s, error: %+v", code, rsp, err)
	code, rsp, e = s.ForwardRspCheck(code, rsp, err)
	if code < http.StatusOK || code >= http.StatusBadRequest {
		return
	}

	// 正常创建, 解析server session id
	obj := serversession.ShowServerSessionResponseFromAppGW{}
	if newErr := json.Unmarshal(rsp, &obj); newErr != nil {
		// 解析server session id失败, 内部服务异常
		s.Logger.Error("app gateway return show server session rsp unmarshal error, rsp:%s, err:%v", rsp,
			newErr)
		return 0, nil, errors.NewError(errors.ServerInternalError)
	}

	newRsp := &serversession.ShowServerSessionResponse{
		ServerSession: *generateServerSessions(&obj.ServerSession),
	}

	rsp, err = json.Marshal(newRsp)
	if err != nil {
		s.Logger.Error("marshal response error: %+v, rsp: %s", err, newRsp)
		return 0, nil, errors.NewError(errors.ServerInternalError)
	}

	return code, rsp, nil
}

func (s *Service) forwardListToAPPGW(region string) (code int, rsp []byte, err error) {
	url := clientGetServiceEndpoint(client.ServiceNameAPPGW, region) + constants.ServerSessionsUrl
	req := clientNewRequest(client.ServiceNameAPPGW, url, http.MethodGet, nil)
	req.SetQuery(params.QueryOffset,
		utils.GetStringIfNotEmpty(s.Ctx.Input.Query(params.QueryOffset), params.DefaultOffset))
	req.SetQuery(params.QueryLimit,
		utils.GetStringIfNotEmpty(s.Ctx.Input.Query(params.QueryLimit), params.DefaultLimit))
	req.SetQuery(params.QueryFleetId, s.Ctx.Input.Query(params.QueryFleetId))
	req.SetQuery(params.QuerySort, utils.GetStringIfNotEmpty(s.Ctx.Input.Query(params.QuerySort), params.DefaultSort))
	queryState := s.Ctx.Input.Query(params.QueryState)
	if queryState != "" {
		req.SetQuery(params.QueryState, queryState)
	}
	req.SetHeader(map[string]string{
		logger.RequestId: fmt.Sprintf("%s", s.Ctx.Input.GetData(logger.RequestId)),
	})
	return req.DoRequest()
}

// List 查询服务端会话列表
func (s *Service) List() (code int, rsp []byte, e *errors.CodedError) {
	if err := s.SetFleetById(s.Ctx.Input.Query(params.QueryFleetId)); err != nil {
		s.Logger.Error("get fleet in list server session error, fleetId:%s",
			s.Ctx.Input.Query(params.QueryFleetId))
		return 0, nil, err
	}

	code, rsp, err := s.forwardListToAPPGW(s.Fleet.Region)
	s.Logger.Info("list server session to app gateway, code: %d, rsp: %s, error: %+v", code, rsp, err)
	code, rsp, e = s.ForwardRspCheck(code, rsp, err)
	if code < http.StatusOK || code >= http.StatusBadRequest {
		return
	}

	// 正常创建, 解析server session id
	obj := serversession.ListServerSessionResponseFromAppGW{}
	if newErr := json.Unmarshal(rsp, &obj); newErr != nil {
		s.Logger.Error("list server session to app gateway rsp unmarshal error, rsp:%s, err:%v", rsp,
			newErr)
		return 0, nil, errors.NewError(errors.ServerInternalError)
	}

	newRsp := &serversession.ListServerSessionResponse{
		Count: obj.Count,
	}
	var serverSessions = make([]serversession.ServerSession, 0)
	for _, value := range obj.ServerSessions {
		tmp := generateServerSessions(&value)
		serverSessions = append(serverSessions, *tmp)
	}
	newRsp.ServerSessions = serverSessions
	rsp, err = json.Marshal(newRsp)
	if err != nil {
		s.Logger.Error("marshal response error: %+v, rsp: %s", err, newRsp)
		return 0, nil, errors.NewError(errors.ServerInternalError)
	}

	return code, rsp, nil
}

func generateServerSessions(serverSessionAppGW *serversession.ServerSessionFromAppGW) *serversession.ServerSession {
	return &serversession.ServerSession{
		ServerSessionId:                         serverSessionAppGW.ServerSessionId,
		Name:                                    serverSessionAppGW.Name,
		CreatorId:                               serverSessionAppGW.CreatorId,
		FleetId:                                 serverSessionAppGW.FleetId,
		Properties:                              serverSessionAppGW.Properties,
		ServerSessionData:                       serverSessionAppGW.ServerSessionData,
		CurrentClientSessionCount:               serverSessionAppGW.CurrentClientSessionCount,
		MaxClientSessionCount:                   serverSessionAppGW.MaxClientSessionCount,
		State:                                   serverSessionAppGW.State,
		StateReason:                             serverSessionAppGW.StateReason,
		IpAddress:                               serverSessionAppGW.IpAddress,
		Port:                                    serverSessionAppGW.Port,
		ClientSessionCreationPolicy:             serverSessionAppGW.ClientSessionCreationPolicy,
		ServerSessionProtectionPolicy:           serverSessionAppGW.ServerSessionProtectionPolicy,
		ServerSessionProtectionTimeLimitMinutes: serverSessionAppGW.ServerSessionProtectionTimeLimitMinutes,
	}
}
