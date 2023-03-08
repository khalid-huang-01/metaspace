// Copyright (c) Huawei Technologies Co., Ltd. 2022-2022. All rights reserved.

// alias创建方法
package alias

import (
	"fleetmanager/api/errors"
	"fleetmanager/api/model/alias"
	"fleetmanager/db/dao"
	"fmt"
)

// Create: 创建Alias
func (s *Service) Create(r *alias.CreateRequest) (*alias.CreateResponse, *errors.CodedError) {
	s.createRequest = r
	// 校验路由策略
	if r.Type == dao.AliasTypeActive && len(s.createRequest.AssociatedFleets) == 0 {
		return nil, errors.NewErrorF(errors.InvalidParameterValue, "active alias must associated one fleet at least")
	}
	for _, associatedFleet := range s.createRequest.AssociatedFleets {
		if err := s.SetFleetById(associatedFleet.FleetId); err != nil {
			return nil, errors.NewErrorF(errors.FleetNotInDB, fmt.Sprintf("fleet_id: %s", associatedFleet.FleetId))
		}
		if s.Fleet.State != dao.FleetStateActive {
			return nil, errors.NewErrorF(errors.FleetNotActive, fmt.Sprintf("fleet_id: %s", associatedFleet.FleetId))
		}
	}
	if err := s.insertDb(); err != nil {
		return nil, err
	}
	aliasM, err := buildAliasModel(s.alias)
	if err != nil {
		return nil, errors.NewErrorF(errors.ServerInternalError, err.Error())
	}
	rsp := &alias.CreateResponse{
		Alias: *aliasM,
	}
	return rsp, nil
}

// insertDb: 创建别名信息入库
func (s *Service) insertDb() *errors.CodedError {
	err := s.buildAlias()
	if err != nil {
		s.Logger.Error("build alias error: %v", err)
		return errors.NewErrorF(errors.ServerInternalError, err.Error())
	}
	if err := dao.GetAliasStorage().Insert(s.alias); err != nil {
		s.Logger.Error("insert alias error: %v", err)
		return errors.NewError(errors.DBError)
	}
	return nil
}
