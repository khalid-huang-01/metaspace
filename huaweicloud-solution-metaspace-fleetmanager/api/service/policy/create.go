// Copyright (c) Huawei Technologies Co., Ltd. 2022-2022. All rights reserved.

// 弹性伸缩策略创建方法
package policy

import (
	"encoding/json"
	"fleetmanager/api/errors"
	"fleetmanager/api/model/policy"
	"fleetmanager/api/service/constants"
	"fleetmanager/client"
	"fleetmanager/db/dao"
	"fleetmanager/logger"
	"fleetmanager/utils"
	"fmt"
	"net/http"
)

func (s *Service) allowCreate() *errors.CodedError {
	policies, err := dao.GetScalingPolicyStorage().List(dao.Filters{"FleetId": s.Fleet.Id}, 0, -1)
	if err != nil {
		s.Logger.Error("get policies db error: %v", err)
		return errors.NewError(errors.DBError)
	}

	for _, p := range policies {
		if p.PolicyType == dao.TargetBasedPolicy && s.createReq.PolicyType == dao.TargetBasedPolicy {
			s.Logger.Error("one fleet can have only one target based policy")
			return errors.NewError(errors.DuplicatePolicy)
		}
	}

	return nil
}

func (s *Service) makeCreateReq() (*policy.CreateRequestToAASS, string, error) {
	group, err := dao.GetScalingGroupStorage().GetOne(dao.Filters{"FleetId": s.Fleet.Id})
	if err != nil {
		s.Logger.Error("get scaling group db error: %v", err)
		return nil, "", fmt.Errorf("get scaling group in create policy error, fleetId:%s", s.Fleet.Id)
	}

	createReq := &policy.CreateRequestToAASS{
		InstanceScalingGroupId: group.Id,
		CreateRequest:          *s.createReq,
	}

	return createReq, group.ResourceProjectId, nil
}

func (s *Service) forwardCreateToAASS() (code int, rsp []byte, err error) {
	createReq, resProjectId, err := s.makeCreateReq()
	if err != nil {
		return 0, nil, err
	}

	body, err := json.Marshal(createReq)
	if err != nil {
		return 0, nil, err
	}

	s.resourceProjectId = resProjectId
	url := client.GetServiceEndpoint(client.ServiceNameAASS, s.Fleet.Region) +
		fmt.Sprintf(constants.CreateScalingPolicyUrl, s.resourceProjectId)
	req := client.NewRequest(client.ServiceNameAASS, url, http.MethodPost, body)
	req.SetHeader(map[string]string{
		logger.RequestId: fmt.Sprintf("%s", s.Ctx.Input.GetData(logger.RequestId)),
	})
	return req.DoRequest()
}

func (s *Service) insertDb(rsp *policy.CreateResponseFromAASS, resProjectId string) (*dao.ScalingPolicy,
	*errors.CodedError) {
	p := &dao.ScalingPolicy{
		Id:                       rsp.ScalingPolicyId,
		Name:                     s.createReq.Name,
		FleetId:                  s.Fleet.Id,
		PolicyType:               s.createReq.PolicyType,
		State:                    dao.PolicyStateActive,
		ScalingTarget:            s.createReq.ScalingTarget,
		TargetBasedConfiguration: utils.ToJson(s.createReq.TargetBasedConfiguration),
		ResourceProjectId:        resProjectId,
	}

	if err := dao.GetScalingPolicyStorage().Insert(p); err != nil {
		s.Logger.Error("insert policy to db error: %+v", err)
		return nil, errors.NewError(errors.DBError)
	}

	return p, nil
}

// Create 创建fleet策略
func (s *Service) Create(r *policy.CreateRequest) (code int, rsp []byte, e *errors.CodedError) {
	s.createReq = r
	if err := s.SetFleet(); err != nil {
		return 0, nil, err
	}

	if s.Fleet.State != dao.FleetStateActive {
		s.Logger.Error("fleet %s is not active, not ready for create policy", s.Fleet)
		return 0, nil, errors.NewError(errors.FleetStateNotSupportCreatePolicy)
	}

	if err := s.allowCreate(); err != nil {
		return 0, nil, err
	}

	var err error
	code, rsp, err = s.forwardCreateToAASS()
	s.Logger.Info("create policy to aass, code: %d, rsp: %s, error: %+v", code, rsp, err)
	code, rsp, e = s.ForwardRspCheck(code, rsp, err)
	if code < http.StatusOK || code >= http.StatusBadRequest {
		return
	}

	cRsp := &policy.CreateResponseFromAASS{}
	if err := json.Unmarshal(rsp, cRsp); err != nil {
		s.Logger.Error("marshal response error: %v, rsp: %s", err, rsp)
		return 0, nil, errors.NewError(errors.ServerInternalError)

	}

	// 更新数据库:TODO(nkaptx):存在上下不一致的情况
	policyDao, newErr := s.insertDb(cRsp, s.resourceProjectId)
	if newErr != nil {
		logger.M.Error("Insert policy to db error:%+v, aass rsp:%s, fleet:%v", newErr, rsp, s.Fleet)
		return 0, nil, newErr
	}

	newRsp := &policy.ScalingPolicy{
		Id:                       policyDao.Id,
		Name:                     policyDao.Name,
		FleetId:                  policyDao.FleetId,
		PolicyType:               policyDao.PolicyType,
		ScalingTarget:            policyDao.ScalingTarget,
		State:                    policyDao.State,
		TargetBasedConfiguration: s.createReq.TargetBasedConfiguration,
	}
	rsp, err = json.Marshal(newRsp)
	if err != nil {
		s.Logger.Error("marshal response error: %v, rsp: %s", err, newRsp)
		return 0, nil, errors.NewError(errors.ServerInternalError)
	}
	return http.StatusCreated, rsp, nil
}
