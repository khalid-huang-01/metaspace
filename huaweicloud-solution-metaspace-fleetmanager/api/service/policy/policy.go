// Copyright (c) Huawei Technologies Co., Ltd. 2022-2022. All rights reserved.

// 弹性伸缩策略服务
package policy

import (
	"fleetmanager/api/errors"
	"fleetmanager/api/model/policy"
	"fleetmanager/api/service/base"
	"fleetmanager/db/dao"
	"fleetmanager/logger"
	"github.com/beego/beego/v2/server/web/context"
	"github.com/beego/beego/v2/client/orm"
)

type Service struct {
	base.FleetService
	resourceProjectId string
	createReq         *policy.CreateRequest
	updateReq         *policy.UpdateRequest
	scalingPolicy     *dao.ScalingPolicy
	createRsp         *policy.CreateResponseFromAASS
}

// NewPolicyService 新建fleet policy服务
func NewPolicyService(ctx *context.Context, logger *logger.FMLogger) *Service {
	s := &Service{
		FleetService: base.FleetService{
			Ctx:    ctx,
			Logger: logger,
		},
	}
	return s
}

// 设置policy对象
func (s *Service) SetPolicyById(policyId string) *errors.CodedError {
	p, err := dao.GetScalingPolicyStorage().Get(dao.Filters{"Id": policyId})
	if err != nil {
		s.Logger.Error("get policy from db error: %v", err)
		if err == orm.ErrNoRows {
			return errors.NewError(errors.PolicyNotFound)
		}
		return errors.NewError(errors.DBError)
	}

	s.scalingPolicy = p
	return nil
}
