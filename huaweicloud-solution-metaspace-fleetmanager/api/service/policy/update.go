// Copyright (c) Huawei Technologies Co., Ltd. 2022-2022. All rights reserved.

// 弹性伸缩策略更新服务
package policy

import (
	"encoding/json"
	"fleetmanager/api/errors"
	"fleetmanager/api/model/policy"
	"fleetmanager/api/params"
	"fleetmanager/api/service/constants"
	"fleetmanager/client"
	"fleetmanager/db/dao"
	"fleetmanager/logger"
	"fleetmanager/utils"
	"fmt"
	"net/http"
)

func (s *Service) forwardUpdateToAASS() (code int, rsp []byte, err error) {
	body, err := json.Marshal(s.updateReq)
	if err != nil {
		return
	}

	url := client.GetServiceEndpoint(client.ServiceNameAASS, s.Fleet.Region) +
		fmt.Sprintf(constants.ScalingPolicyUrlPattern, s.scalingPolicy.ResourceProjectId, s.scalingPolicy.Id)
	req := client.NewRequest(client.ServiceNameAASS, url, http.MethodPut, body)
	req.SetHeader(map[string]string{
		logger.RequestId: fmt.Sprintf("%s", s.Ctx.Input.GetData(logger.RequestId)),
	})
	return req.DoRequest()
}

func (s *Service) updateToDb(p *dao.ScalingPolicy) *errors.CodedError {
	if s.updateReq.TargetBasedConfiguration != nil {
		p.TargetBasedConfiguration = utils.ToJson(*s.updateReq.TargetBasedConfiguration)
	}

	if s.updateReq.Name != nil {
		p.Name = *s.updateReq.Name
	}

	s.Logger.Info("update policy db p:%+v", p)
	err := dao.GetScalingPolicyStorage().Update(p, "Name", "TargetBasedConfiguration")
	if err != nil {
		s.Logger.Error("update scaling policy in db error: %v", err)
		return errors.NewError(errors.DBError)
	}
	return nil
}

// Update 更新fleet policy
func (s *Service) Update(r *policy.UpdateRequest) (code int, rsp []byte, e *errors.CodedError) {
	s.Logger.Info("update policy:%+v", r)

	// policyId此处不会为空(route策略机制保证s.Ctx.Input.Param不会为空)
	policyId := s.Ctx.Input.Param(params.PolicyId)
	if err := s.SetPolicyById(policyId); err != nil {
		return 0, nil, err
	}

	s.updateReq = r
	if err := s.SetFleetById(s.scalingPolicy.FleetId); err != nil {
		s.Logger.Error("get fleet in update policy error, policyId:%s, fleetId:%s",
			policyId, s.scalingPolicy.FleetId)
		return 0, nil, err
	}

	var err error
	code, rsp, err = s.forwardUpdateToAASS()
	s.Logger.Info("update policy to aass, code: %d, rsp: %s, error: %+v", code, rsp, err)
	code, rsp, e = s.ForwardRspCheck(code, rsp, err)
	if code < http.StatusOK || code >= http.StatusBadRequest {
		return
	}

	// 更新数据库:TODO(nkaptx):存在上下不一致的情况
	if err := s.updateToDb(s.scalingPolicy); err != nil {
		logger.M.
			WithField(logger.Stage, "update_scaling_policy").
			WithField(logger.Error, err.Error()).
			Error("Update policy to db error, aass Rsp:%s, fleet:%v", rsp, s.Fleet)
		return 0, nil, errors.NewError(errors.ServerInternalError)
	}
	return
}
