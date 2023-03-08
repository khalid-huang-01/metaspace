// Copyright (c) Huawei Technologies Co., Ltd. 2022-2022. All rights reserved.

// 服务端会话更新服务
package serversession

import (
	"encoding/json"
	"fleetmanager/api/errors"
	"fleetmanager/api/model/serversession"
	"fleetmanager/api/params"
	"fleetmanager/api/service/constants"
	"fleetmanager/client"
	"fleetmanager/logger"
	"fmt"
	"net/http"
)

func (s *Service) forwardUpdateToAPPGW(region string) (code int, rsp []byte, err error) {
	updateReq := serversession.UpdateRequestToAppGW{
		Name:                                    s.updateReq.Name,
		MaxClientSessionNum:                     s.updateReq.MaxClientSessionCount,
		ServerSessionProtectionPolicy:           s.updateReq.ServerSessionProtectionPolicy,
		ServerSessionProtectionTimeLimitMinutes: s.updateReq.ServerSessionProtectionTimeLimitMinutes,
		ClientSessionCreationPolicy:             s.updateReq.ClientSessionCreationPolicy,
	}
	body, err := json.Marshal(updateReq)
	if err != nil {
		return 0, nil, err
	}

	sessionId := s.Ctx.Input.Param(params.ServerSessionId)
	url := client.GetServiceEndpoint(client.ServiceNameAPPGW, region) +
		fmt.Sprintf(constants.ServerSessionUrlPattern, sessionId)
	req := client.NewRequest(client.ServiceNameAPPGW, url, http.MethodPut, body)
	req.SetHeader(map[string]string{
		logger.RequestId: fmt.Sprintf("%s", s.Ctx.Input.GetData(logger.RequestId)),
	})
	return req.DoRequest()
}

// Update 更新服务端会话
func (s *Service) Update(r *serversession.UpdateRequest) (code int, rsp []byte, e *errors.CodedError) {
	s.updateReq = r
	if err := s.SetFleetByServerSessionId(s.Ctx.Input.Param(params.ServerSessionId)); err != nil {
		s.Logger.Error("get fleet in show server session error, serverSessionId:%s",
			s.Ctx.Input.Param(params.ServerSessionId))
		return 0, nil, err
	}

	code, rsp, newErr := s.forwardUpdateToAPPGW(s.Fleet.Region)
	s.Logger.Info("forward update server session to app gateway, code:%d, rsp:%s, err:%v", code, rsp, newErr)
	return s.ForwardRspCheck(code, rsp, newErr)
}
