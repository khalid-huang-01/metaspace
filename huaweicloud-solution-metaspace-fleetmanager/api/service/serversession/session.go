// Copyright (c) Huawei Technologies Co., Ltd. 2022-2022. All rights reserved.

// 服务端会话相关服务
package serversession

import (
	"fleetmanager/api/errors"
	"fleetmanager/api/model/serversession"
	"fleetmanager/api/service/base"
	"fleetmanager/db/dao"
	"fleetmanager/logger"
	"github.com/beego/beego/v2/client/orm"
	"github.com/beego/beego/v2/server/web/context"
)

type Service struct {
	base.FleetService
	fleetServerSession *dao.FleetServerSession
	createReq          *serversession.CreateRequest
	updateReq          *serversession.UpdateRequest
	alias              *dao.Alias
}

// NewServerSessionService 新建服务端会话管理服务
func NewServerSessionService(ctx *context.Context, logger *logger.FMLogger) *Service {
	s := &Service{
		FleetService: base.FleetService{
			Ctx:    ctx,
			Logger: logger,
		},
	}

	return s
}

// setAlias 校验alias是否存在
func (s *Service) setAlias(aliasId string) *errors.CodedError {
	filter := dao.Filters{
		"Id":        aliasId,
	}
	f, err := dao.GetAliasStorage().Get(filter)
	if err != nil {
		if err == orm.ErrNoRows {
			return errors.NewError(errors.AliasNotFound)
		}
		s.Logger.Error("get alias info db error: %v", err)
		return errors.NewError(errors.DBError)
	}

	s.alias = f
	return nil

}
