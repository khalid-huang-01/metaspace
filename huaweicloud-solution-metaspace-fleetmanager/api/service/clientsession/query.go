// Copyright (c) Huawei Technologies Co., Ltd. 2022-2022. All rights reserved.

// 客户端会话查询方法
package clientsession

import (
	"encoding/json"
	"fleetmanager/api/errors"
	"fleetmanager/api/model/clientsession"
	"fleetmanager/api/params"
	"fleetmanager/api/service/constants"
	"fleetmanager/logger"
	"fleetmanager/client"
	"fleetmanager/utils"
	"fmt"
	"net/http"
)

func (s *Service) forwardShowToAPPGW(region string) (code int, rsp []byte, err error) {
	clientSessionId := s.Ctx.Input.Param(params.ClientSessionId)
	url := client.GetServiceEndpoint(client.ServiceNameAPPGW, region) +
		fmt.Sprintf(constants.ClientSessionUrlPattern, clientSessionId)
	req := client.NewRequest(client.ServiceNameAPPGW, url, http.MethodGet, nil)
	req.SetQuery(params.QueryServerSessionId, s.Ctx.Input.Query(params.ServerSessionId))
	req.SetHeader(map[string]string{
		logger.RequestId: fmt.Sprintf("%s", s.Ctx.Input.GetData(logger.RequestId)),
	})
	return req.DoRequest()
}

// Show 查询客户端会话详情
func (s *Service) Show() (code int, rsp []byte, e *errors.CodedError) {
	if err := s.SetFleetByServerSessionId(s.Ctx.Input.Param(params.ServerSessionId)); err != nil {
		s.Logger.Error("get fleet in show client session error, serverSessionId:%s",
			s.Ctx.Input.Param(params.ServerSessionId))
		return 0, nil, err
	}

	code, rsp, err := s.forwardShowToAPPGW(s.Fleet.Region)
	s.Logger.Info("forward show client session to app gateway, code:%d, rsp:%s, err:%v", code, rsp, err)
	code, rsp, e = s.ForwardRspCheck(code, rsp, err)
	if code < http.StatusOK || code >= http.StatusBadRequest {
		return
	}

	obj := clientsession.ShowResponseFromAPPGW{}
	if newErr := json.Unmarshal(rsp, &obj); newErr != nil {
		s.Logger.Error("app gateway return show client session rsp unmarshal error, rsp:%s, err:%v", rsp,
			newErr)
		return 0, nil, errors.NewError(errors.ServerInternalError)
	}

	newRsp := &clientsession.ShowResponse{
		ClientSession: *generateClientSessions(&obj.ClientSession),
	}

	rsp, err = json.Marshal(newRsp)
	if err != nil {
		s.Logger.Error("marshal response error: %+v, rsp: %s", err, newRsp)
		return 0, nil, errors.NewError(errors.ServerInternalError)
	}

	return http.StatusOK, rsp, nil
}

func (s *Service) forwardListToAPPGW(region string) (code int, rsp []byte, err error) {
	url := client.GetServiceEndpoint(client.ServiceNameAPPGW, region) + constants.ClientSessionsUrl
	req := client.NewRequest(client.ServiceNameAPPGW, url, http.MethodGet, nil)
	req.SetQuery(params.QueryOffset,
		utils.GetStringIfNotEmpty(s.Ctx.Input.Query(params.QueryOffset), params.DefaultOffset))
	req.SetQuery(params.QueryLimit,
		utils.GetStringIfNotEmpty(s.Ctx.Input.Query(params.QueryLimit), params.DefaultLimit))
	req.SetQuery(params.QueryServerSessionId, s.Ctx.Input.Query(params.ServerSessionId))
	req.SetQuery(params.QuerySort, params.DefaultSort)
	req.SetHeader(map[string]string{
		logger.RequestId: fmt.Sprintf("%s", s.Ctx.Input.GetData(logger.RequestId)),
	})
	return req.DoRequest()
}

// List 查询客户端会话列表
func (s *Service) List() (code int, rsp []byte, e *errors.CodedError) {
	if err := s.SetFleetByServerSessionId(s.Ctx.Input.Param(params.ServerSessionId)); err != nil {
		s.Logger.Error("get fleet in list client session error, serverSessionId:%s",
			s.Ctx.Input.Param(params.ServerSessionId))
		return 0, nil, err
	}

	code, rsp, err := s.forwardListToAPPGW(s.Fleet.Region)
	s.Logger.Info("list client session to appgw, code: %d, rsp: %s, error: %+v", code, rsp, err)
	code, rsp, e = s.ForwardRspCheck(code, rsp, err)
	if code < http.StatusOK || code >= http.StatusBadRequest {
		return
	}

	obj := clientsession.ListResponseFromAPPGW{}
	if newErr := json.Unmarshal(rsp, &obj); newErr != nil {
		s.Logger.Error("app gateway return list client session rsp unmarshal error, rsp:%s, err:%v", rsp,
			newErr)
		return 0, nil, errors.NewError(errors.ServerInternalError)
	}

	newRsp := &clientsession.ListResponse{
		Count: obj.Count,
	}
	var clientSessions = make([]clientsession.ClientSession, 0)
	for _, value := range obj.ClientSessions {
		tmp := generateClientSessions(&value)
		clientSessions = append(clientSessions, *tmp)
	}
	newRsp.ClientSessions = clientSessions
	rsp, err = json.Marshal(newRsp)
	if err != nil {
		s.Logger.Error("marshal response error: %+v, rsp: %s", err, newRsp)
		return 0, nil, errors.NewError(errors.ServerInternalError)
	}

	return http.StatusOK, rsp, nil
}
