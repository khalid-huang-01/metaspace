// Copyright (c) Huawei Technologies Co., Ltd. 2022-2022. All rights reserved.

// 弹性伸缩策略删除方法
package policy

import (
	"fleetmanager/api/errors"
	"fleetmanager/api/params"
	"fleetmanager/api/service/constants"
	"fleetmanager/client"
	"fleetmanager/db/dao"
	"fleetmanager/logger"
	"fmt"
	"github.com/beego/beego/v2/client/orm"
	"net/http"
)

func (s *Service) forwardDeleteToAASS() (code int, rsp []byte, err error) {
	url := client.GetServiceEndpoint(client.ServiceNameAASS, s.Fleet.Region) +
		fmt.Sprintf(constants.ScalingPolicyUrlPattern, s.scalingPolicy.ResourceProjectId, s.scalingPolicy.Id)
	req := client.NewRequest(client.ServiceNameAASS, url, http.MethodDelete, nil)
	req.SetHeader(map[string]string{
		logger.RequestId: fmt.Sprintf("%s", s.Ctx.Input.GetData(logger.RequestId)),
	})
	return req.DoRequest()
}

func (s *Service) deletePolicyInDb() *errors.CodedError {
	policyId := s.Ctx.Input.Param(params.PolicyId)
	if err := dao.GetScalingPolicyStorage().Delete(&dao.ScalingPolicy{Id: policyId}); err != nil {
		s.Logger.Error("delete policy error: %v", err)
		if err == orm.ErrNoRows {
			return nil
		}

		return errors.NewError(errors.DBError)
	}
	return nil
}

// Delete 删除fleet策略
func (s *Service) Delete() (code int, rsp []byte, e *errors.CodedError) {
	if err := s.SetFleet(); err != nil {
		return 0, nil, err
	}

	if s.Fleet.State != dao.FleetStateActive && s.Fleet.State != dao.FleetStateError {
		s.Logger.Error("fleet %s do not support to delete policy", s.Fleet)
		return 0, nil, errors.NewError(errors.FleetStateNotSupportDeletePolicy)
	}

	policyId := s.Ctx.Input.Param(params.PolicyId)
	if err := s.SetPolicyById(policyId); err != nil {
		return 0, nil, err
	}

	var err error
	code, rsp, err = s.forwardDeleteToAASS()
	s.Logger.Info("delete policy to aass, code: %d, rsp: %v, error: %+v", code, rsp, err)
	code, rsp, e = s.ForwardRspCheck(code, rsp, err)
	if code < http.StatusOK || code >= http.StatusBadRequest {
		return
	}

	// 删除数据库
	if err := s.deletePolicyInDb(); err != nil {
		logger.M.WithField(logger.Stage, "delete_scaling_policy").
			WithField(logger.Error, err.Error()).
			Error("Delete policy to db error, aass Rsp:%s, fleet:%v", rsp, s.Fleet)
		return 0, nil, err
	}

	return http.StatusNoContent, nil, nil
}
