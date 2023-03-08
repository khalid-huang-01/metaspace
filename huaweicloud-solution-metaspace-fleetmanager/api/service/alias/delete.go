// Copyright (c) Huawei Technologies Co., Ltd. 2022-2022. All rights reserved.

// alias删除方法
package alias

import (
	"fleetmanager/api/errors"
	"fleetmanager/api/params"
	"fleetmanager/db/dao"
	"github.com/beego/beego/v2/client/orm"
	"github.com/beego/beego/v2/server/web/context"
)

// Delete: 删除alias
func (s *Service) Delete(ctx *context.Context) *errors.CodedError {

	if err := s.setAlias(); err != nil {
		s.Logger.Error("get alias info db error: %v", err)
		if err == orm.ErrNoRows {
			return errors.NewError(errors.AliasNotFound)
		} else {
			return errors.NewError(errors.DBError)
		}
	}

	aliasId := ctx.Input.Param(params.AliasId)
	s.alias.Type = dao.AliasTypeTerminated
	s.Logger.Info("receive one delete alias: %s request", aliasId)
	if err := dao.GetAliasStorage().Update(s.alias, "Type"); err != nil {
		s.Logger.Error("delete alias db error:%v", err)
		return errors.NewError(errors.DBError)
	}
	return nil
}
