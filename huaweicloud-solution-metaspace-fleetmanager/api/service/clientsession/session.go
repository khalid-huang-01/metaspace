// Copyright (c) Huawei Technologies Co., Ltd. 2022-2022. All rights reserved.

// 客户端会话服务方法
package clientsession

import (
	"fleetmanager/api/model/clientsession"
	"fleetmanager/api/service/base"
	"fleetmanager/logger"
	"github.com/beego/beego/v2/server/web/context"
)

type Service struct {
	base.FleetService
	createReq      *clientsession.CreateRequest
	batchCreateReq *clientsession.BatchCreateRequest
}

// NewClientSessionService 新建客户端会话服务
func NewClientSessionService(ctx *context.Context, logger *logger.FMLogger) *Service {
	s := &Service{
		FleetService: base.FleetService{
			Ctx:    ctx,
			Logger: logger,
		},
	}

	return s
}
