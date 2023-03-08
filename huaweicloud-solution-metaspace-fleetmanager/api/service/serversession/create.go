// Copyright (c) Huawei Technologies Co., Ltd. 2022-2022. All rights reserved.

// 服务端会话创建服务
package serversession

import (
	"encoding/json"
	"fleetmanager/api/errors"
	"fleetmanager/api/model/serversession"
	AliasService "fleetmanager/api/service/alias"
	"fleetmanager/api/service/constants"
	"fleetmanager/client"
	"fleetmanager/db/dao"
	"fleetmanager/logger"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"
)

func (s *Service) forwardCreateToAPPGW(region string) (code int, rsp []byte, err error) {
	createReq := serversession.CreateRequestToAppGW{
		FleetId:                 s.createReq.FleetId,
		CreatorId:               s.createReq.CreatorId,
		Name:                    s.createReq.Name,
		MaxClientSessionNum:     s.createReq.MaxClientSessionCount,
		IdempotencyToken:        s.createReq.IdempotencyToken,
		ServerSessionData:       s.createReq.ServerSessionData,
		ServerSessionProperties: s.createReq.ServerSessionProperties,
	}
	body, err := json.Marshal(createReq)
	if err != nil {
		s.Logger.Error("marshal body error: %v", err)
		return
	}
	url := client.GetServiceEndpoint(client.ServiceNameAPPGW, region) + constants.ServerSessionsUrl
	req := client.NewRequest(client.ServiceNameAPPGW, url, http.MethodPost, body)
	
	req.SetHeader(map[string]string{
		logger.RequestId: fmt.Sprintf("%s", s.Ctx.Input.GetData(logger.RequestId)),
	})
	return req.DoRequest()
}

// Create 创建服务端会话
func (s *Service) Create(r *serversession.CreateRequest) (code int, rsp []byte, e *errors.CodedError) {
	s.createReq = r
	if err := s.checkRequestParameters(); err != nil {
		return 0, nil, err
	}

	if err := s.SetFleetById(s.createReq.FleetId); err != nil {
		return 0, nil, errors.NewError(errors.FleetNotInDB)
	}
	code, rsp, err := s.forwardCreateToAPPGW(s.Fleet.Region)
	s.Logger.Info("forward create server session to app gateway, code:%d, rsp:%s, err:%v", code, rsp, err)
	code, rsp, e = s.ForwardRspCheck(code, rsp, err)
	if code < http.StatusOK || code >= http.StatusBadRequest {
		return
	}

	// 正常创建, 解析server session id
	obj := serversession.CreateServerSessionResponseFromAppGW{}
	if newErr := json.Unmarshal(rsp, &obj); newErr != nil {
		// 解析server session id失败, 内部服务异常
		s.Logger.Error("app gateway return create server session rsp unmarshal error, rsp:%s, err:%v", rsp,
			newErr)
		return 0, nil, errors.NewError(errors.ServerInternalError)
	}

	if err := s.insertDB(obj.ServerSession.ServerSessionId); err != nil {
		logger.M.Error("Insert server session to db error:%+v, appGWRsp:%s, fleet:%v", err, rsp, s.Fleet)
		return 0, nil, errors.NewError(errors.ServerInternalError)
	}

	newRsp := &serversession.CreateServerSessionResponse{
		ServerSession: *generateServerSessions(&obj.ServerSession),
	}

	rsp, err = json.Marshal(newRsp)
	if err != nil {
		s.Logger.Error("marshal response error: %+v, rsp: %s", err, newRsp)
		return 0, nil, errors.NewError(errors.ServerInternalError)
	}

	return http.StatusCreated, rsp, nil
}

// 记录FleetServerSession信息
func (s *Service) insertDB(fssId string) error {
	s.buildFleetServerSession(fssId)
	if err := dao.GetFleetServerSessionStorage().Insert(s.fleetServerSession); err != nil {
		s.Logger.Error("insert fleet server session db error: %v", err)
		return errors.NewError(errors.DBError)
	}

	return nil
}

// 构建fleet server session
func (s *Service) buildFleetServerSession(fssId string) {
	u, _ := uuid.NewUUID()
	ffs := &dao.FleetServerSession{
		Id:              u.String(),
		FleetId:         s.Fleet.Id,
		ServerSessionId: fssId,
		Region:          s.Fleet.Region,
		CreationTime:    time.Now().UTC(),
	}
	s.fleetServerSession = ffs
}

// createByAliasId: aliasId创建会话
func (s *Service) createByAliasId(r *serversession.CreateRequest) (e *errors.CodedError) {
	if err := s.setAlias(r.AliasId); err != nil {
		return err
	}
	// 校验路由策略类型
	if s.alias.Type == dao.AliasTypeDeactive {
		return errors.NewErrorF(errors.AliasIsDeactive, fmt.Sprintf(
			" message: %s", s.alias.Message))
	}

	fleetId, err := AliasService.GenerateFleetByAssociateFleetWeight(s.alias.AssociatedFleets)
	if err != nil {
		return err
	}
	s.createReq.FleetId = fleetId
	return nil
}

// checkRequestParameters: 请求参数校验
func (s *Service) checkRequestParameters() *errors.CodedError {
	// 处理字符串
	s.createReq.AliasId = strings.Replace(s.createReq.AliasId, " ", "", -1)
	s.createReq.FleetId = strings.Replace(s.createReq.FleetId, " ", "", -1)
	// AliasId，FleetId同时填充
	if len(s.createReq.AliasId) > 0 && len(s.createReq.FleetId) > 0 {
		s.Logger.Error("Create a service session,reference fleetId and aliasId error, aliasId:%v,fleetId:%v",
			s.createReq.AliasId, s.createReq.FleetId)
		return errors.NewError(errors.ReferenceFleetIdAndAliasIdNotBoth)
	}
	// AliasId，FleetId同时为空
	if len(s.createReq.AliasId) <= 0 && len(s.createReq.FleetId) <= 0 {
		s.Logger.Error("Create a service session. Both fleetId and aliasId are empty error")
		return errors.NewError(errors.FleetIdAndAliasNotBothEmpty)
	}
	// AliasId創建服务会话
	if len(s.createReq.AliasId) > 0 && len(s.createReq.FleetId) <= 0 {
		if err := s.createByAliasId(s.createReq); err != nil {
			s.Logger.Error("aliasId create server session error, err:%v", err)
			return err
		}
	}
	return nil
}

// RspCheck
func (s *Service) FleetNotExist(code int, rsp []byte, err error) (int, []byte, *errors.CodedError) {
	if code < http.StatusOK || code >= http.StatusBadRequest {
		// 创建失败的场景
		if err != nil {
			return 0, nil, errors.NewError(errors.ServerInternalError)
		}
		// 其他错误场景直接透传
		return code, rsp, nil
	}
	return code, rsp, nil
}
