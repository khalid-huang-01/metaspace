// Copyright (c) Huawei Technologies Co., Ltd. 2022-2022. All rights reserved.

// 客户端会话创建方法
package clientsession

import (
	"encoding/json"
	"fleetmanager/api/errors"
	"fleetmanager/api/model/clientsession"
	"fleetmanager/api/params"
	"fleetmanager/api/service/constants"
	"fleetmanager/logger"
	"fleetmanager/client"
	"net/http"
	"fmt"
)

func (s *Service) forwardCreateToAPPGW(region string) (code int, rsp []byte, err error) {
	createReq := clientsession.CreateRequestToAPPGW{
		ServerSessionId: s.Ctx.Input.Param(params.ServerSessionId),
		CreateRequest:   *s.createReq,
	}

	body, err := json.Marshal(createReq)
	if err != nil {
		return 0, nil, err
	}

	url := client.GetServiceEndpoint(client.ServiceNameAPPGW, region) + constants.ClientSessionsUrl
	req := client.NewRequest(client.ServiceNameAPPGW, url, http.MethodPost, body)
	req.SetHeader(map[string]string{
		logger.RequestId: fmt.Sprintf("%s", s.Ctx.Input.GetData(logger.RequestId)),
	})
	return req.DoRequest()
}

// Create 创建客户端会话
func (s *Service) Create(r *clientsession.CreateRequest) (code int, rsp []byte, e *errors.CodedError) {
	s.createReq = r
	if err := s.SetFleetByServerSessionId(s.Ctx.Input.Param(params.ServerSessionId)); err != nil {
		s.Logger.Error("get fleet in create client session error, serverSessionId:%s",
			s.Ctx.Input.Param(params.ServerSessionId))
		return 0, nil, err
	}

	code, rsp, err := s.forwardCreateToAPPGW(s.Fleet.Region)
	s.Logger.Info("forward create client session to app gateway, code:%d, rsp:%s, err:%v", code, rsp, err)
	code, rsp, e = s.ForwardRspCheck(code, rsp, err)
	if code < http.StatusOK || code >= http.StatusBadRequest {
		return
	}

	// 正常创建, 解析server session id
	obj := clientsession.CreateResponseFromAPPGW{}
	if newErr := json.Unmarshal(rsp, &obj); newErr != nil {
		// 解析client session id失败, 内部服务异常
		s.Logger.Error("app gateway return create client session rsp unmarshal error, rsp:%s, err:%v", rsp,
			newErr)
		return 0, nil, errors.NewError(errors.ServerInternalError)
	}

	newRsp := &clientsession.CreateResponse{
		ClientSession: *generateClientSessions(&obj.ClientSession),
	}
	rsp, err = json.Marshal(newRsp)
	if err != nil {
		s.Logger.Error("marshal response error: %+v, rsp: %s", err, newRsp)
		return 0, nil, errors.NewError(errors.ServerInternalError)
	}

	return http.StatusCreated, rsp, nil
}

func (s *Service) forwardBatchCreateToAPPGW(region string) (code int, rsp []byte, err error) {
	createReq := clientsession.BatchCreateRequestToAPPGW{
		ServerSessionId:    s.Ctx.Input.Param(params.ServerSessionId),
		BatchCreateRequest: *s.batchCreateReq,
	}

	body, err := json.Marshal(createReq)
	if err != nil {
		return 0, nil, err
	}

	url := client.GetServiceEndpoint(client.ServiceNameAPPGW, region) + constants.BatchCreateClientSessionUrl
	req := client.NewRequest(client.ServiceNameAPPGW, url, http.MethodPost, body)
	req.SetHeader(map[string]string{
		logger.RequestId: fmt.Sprintf("%s", s.Ctx.Input.GetData(logger.RequestId)),
	})
	return req.DoRequest()
}

// BatchCreate 批量创建客户端会话
func (s *Service) BatchCreate(r *clientsession.BatchCreateRequest) (code int, rsp []byte, e *errors.CodedError) {
	s.batchCreateReq = r
	if err := s.SetFleetByServerSessionId(s.Ctx.Input.Param(params.ServerSessionId)); err != nil {
		s.Logger.Error("get fleet in batch create client session error, serverSessionId:%s",
			s.Ctx.Input.Param(params.ServerSessionId))
		return 0,nil, err
	}

	code, rsp, err := s.forwardBatchCreateToAPPGW(s.Fleet.Region)
	s.Logger.Info("batch create session to appgw, code: %d, rsp: %s, error: %+v", code, rsp, err)
	code, rsp, e = s.ForwardRspCheck(code, rsp, err)
	if code < http.StatusOK || code >= http.StatusBadRequest {
		return
	}

	// 正常创建, 解析server session id
	obj := clientsession.BatchCreateResponseFromAPPGW{}
	if newErr := json.Unmarshal(rsp, &obj); newErr != nil {
		// 解析client session id失败, 内部服务异常
		s.Logger.Error("app gateway return batch create client session rsp unmarshal error, rsp:%s, err:%v", rsp,
			newErr)
		return 0, nil, errors.NewError(errors.ServerInternalError)
	}

	newRsp := &clientsession.BatchCreateResponse{}
	var clientSessions = make([]clientsession.ClientSession, obj.Count)
	for _, value := range obj.ClientSessions {
		tmp := generateClientSessions(&value)
		clientSessions = append(clientSessions, *tmp)
	}
	newRsp.ClientSessions = clientSessions
	newRsp.Count = len(clientSessions)
	rsp, err = json.Marshal(newRsp)
	if err != nil {
		s.Logger.Error("marshal response error: %+v, rsp: %s", err, newRsp)
		return 0, nil, errors.NewError(errors.ServerInternalError)
	}

	return http.StatusCreated, rsp, nil
}

func generateClientSessions(clientSessionAppGW *clientsession.ClientSessionFromAPPGW) *clientsession.ClientSession {
	return &clientsession.ClientSession{
		ServerSessionId: clientSessionAppGW.ServerSessionId,
		ClientSessionId: clientSessionAppGW.ClientSessionId,
		FleetId:         clientSessionAppGW.FleetId,
		IpAddress:       clientSessionAppGW.IpAddress,
		Port:            clientSessionAppGW.Port,
		ClientData:      clientSessionAppGW.ClientData,
		ClientId:        clientSessionAppGW.ClientId,
		State:           clientSessionAppGW.State,
	}
}
