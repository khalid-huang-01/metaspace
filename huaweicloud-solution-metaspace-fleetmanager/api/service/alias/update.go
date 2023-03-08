// Copyright (c) Huawei Technologies Co., Ltd. 2022-2022. All rights reserved.

// alias更新方法
package alias

import (
	"encoding/json"
	"fleetmanager/api/errors"
	"fleetmanager/api/model/alias"
	"fleetmanager/api/validator"
	"fleetmanager/db/dao"
	"fmt"
	"net/http"
	"time"

	"github.com/beego/beego/v2/client/orm"
)

// Update 更新alias属性
func (s *Service) Update() (code int, e *errors.CodedError) {
	// 校验alias是否存在
	if err := s.setAlias(); err != nil {
		s.Logger.Error("get alias info db error: %v", err)
		if err == orm.ErrNoRows {
			return 0, errors.NewError(errors.AliasNotFound)
		} else {
			return 0, errors.NewError(errors.DBError)
		}
	}
	// 校验请求参数
	if e := s.buildUpdateAliasReq(); e != nil {
		return 0, e
	}
	if s.alias.Type == dao.AliasTypeTerminated {
		return 0, errors.NewError(errors.AliasNotFound)
	}
	// 校验路由策略
	if s.updateReq.Type == dao.AliasTypeDeactive {
		return http.StatusNoContent, s.deactiveType()
	}
	return 0, s.activeType()
}

// TerminalType: 终止路由策略校验更新
func (s *Service) deactiveType() (e *errors.CodedError) {
	// 终止策略,alias原关联FleetId清空
	if err := s.updateToDb(); err != nil {
		s.Logger.Error("update alias terminalType to error failed: %v", err)
		return errors.NewError(errors.DBError)
	}
	return nil
}

// SimpleType: 简单路由策略类型校验更新
func (s *Service) activeType() (e *errors.CodedError) {
	// 校验Fleet是否存在以及是否为激活状态
	if s.updateReq.Type == dao.AliasTypeActive && len(s.updateReq.AssociatedFleets) == 0 {
		return errors.NewErrorF(errors.InvalidParameterValue, "active alias must associated one fleet at least")
	}
	for _, associatedFleet := range s.updateReq.AssociatedFleets {
		if err := s.SetFleetById(associatedFleet.FleetId); err != nil {
			return errors.NewErrorF(errors.FleetNotInDB, fmt.Sprintf("fleet_id: %s", associatedFleet.FleetId))
		}
		if s.Fleet.State != dao.FleetStateActive {
			return errors.NewErrorF(errors.FleetNotActive, fmt.Sprintf("fleet_id: %s", associatedFleet.FleetId))
		}
	}
	return s.updateToDb()
}

// buildUpdateAliasReq: 请求参数校验
func (s *Service) buildUpdateAliasReq() *errors.CodedError {
	req := &alias.UpdateAliasRequest{}
	if err := json.Unmarshal(s.Ctx.Input.RequestBody, req); err != nil {
		s.Logger.Error("unmarshal request body %v error: %v", s.Ctx.Input.RequestBody, err)
		return errors.NewErrorF(errors.InvalidParameterValue, " read request params error")
	}

	if err := validator.Validate(req); err != nil {
		s.Logger.Error("request params invalid, reqBody:%s, err:%+v", s.Ctx.Input.RequestBody, err)
		return errors.NewErrorF(errors.InvalidParameterValue, err.Error())
	}
	s.updateReq = req
	return nil
}

// updateToDb: 更新数据库
func (s *Service) updateToDb() *errors.CodedError {

	if s.updateReq.Name != "" {
		s.alias.Name = s.updateReq.Name
	}
	if s.updateReq.Description != "" {
		s.alias.Description = s.updateReq.Description
	}
	if s.updateReq.Type != "" {
		s.alias.Type = s.updateReq.Type
	}
	if s.updateReq.Message != "" {
		s.alias.Message = s.updateReq.Message
	}
	if s.updateReq.AssociatedFleets != nil {
		associatedFleetsByte, err := json.Marshal(s.updateReq.AssociatedFleets)
		if err != nil {
			s.Logger.Error("marshal alias data associated fleets %+v,error: %+v,", s.updateReq.AssociatedFleets, err)
		}
		s.alias.AssociatedFleets = string(associatedFleetsByte)
	}
	s.alias.UpdateTime = time.Now().UTC()
	if err := dao.GetAliasStorage().Update(s.alias, "Name", "Description", "UpdateTime", "AssociatedFleets", "Type", "Message"); err != nil {
		s.Logger.Error("update alias data to error aliasId:%s, err:%+v", s.alias.Id, err)
		return errors.NewError(errors.DBError)
	}
	return nil
}
