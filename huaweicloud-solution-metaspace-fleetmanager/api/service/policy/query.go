// Copyright (c) Huawei Technologies Co., Ltd. 2022-2022. All rights reserved.

// 弹性伸缩策略查询服务
package policy

import (
	"encoding/json"
	"fleetmanager/api/errors"
	"fleetmanager/api/model/policy"
	"fleetmanager/db/dao"
	"net/http"
)

func (s *Service) transPolicy(pd dao.ScalingPolicy) policy.ScalingPolicy {
	p := policy.ScalingPolicy{
		Id:            pd.Id,
		Name:          pd.Name,
		FleetId:       pd.FleetId,
		PolicyType:    pd.PolicyType,
		ScalingTarget: pd.ScalingTarget,
		State:         pd.State,
	}
	err := json.Unmarshal([]byte(pd.TargetBasedConfiguration), &p.TargetBasedConfiguration)
	if err != nil {
		s.Logger.Warn("trans policy error: %v", err)
	}

	return p
}

func (s *Service) transPolicies(pds []dao.ScalingPolicy) []policy.ScalingPolicy {
	policies := make([]policy.ScalingPolicy, len(pds))
	for i, pd := range pds {
		policies[i] = s.transPolicy(pd)
	}
	return policies
}

func (s *Service) checkOffsetLimit(count int64, offset int, limit int) *errors.CodedError {
	if int64(offset*limit) >= count {
		return errors.NewErrorF(errors.InvalidParameterValue, " offset and limit over total count")
	}
	return nil
}

// List 查询fleet policy列表
func (s *Service) List(offset int, limit int) (code int, rsp []byte, e *errors.CodedError) {
	if err := s.SetFleet(); err != nil {
		return 0, nil, err
	}

	totalCount, err := dao.GetScalingPolicyStorage().Count(dao.Filters{"FleetId": s.Fleet.Id})
	if err != nil {
		s.Logger.Error("count scaling policy by fleet id db error: %v", err)
		return 0, nil, errors.NewError(errors.DBError)
	}

	var lRsp *policy.ListResponse
	if totalCount == 0 {
		lRsp = &policy.ListResponse{
			TotalCount: int(totalCount),
			Count: 		0,
			ScalingPolicies: []policy.ScalingPolicy{},
		}
	} else {
		if err := s.checkOffsetLimit(totalCount, offset, limit); err != nil {
			return 0, nil, err
		}
		policies, err := dao.GetScalingPolicyStorage().List(dao.Filters{"FleetId": s.Fleet.Id}, (offset)*limit, limit)
		if err != nil {
			s.Logger.Error("find policies db error: %v", err)
			return 0, nil, errors.NewError(errors.DBError)
		}

		lRsp = &policy.ListResponse{
			TotalCount:         int(totalCount),
			Count: 				len(policies),
			ScalingPolicies: 	s.transPolicies(policies),
		}
	}

	rsp, err = json.Marshal(lRsp)
	if err != nil {
		s.Logger.Error("marshal response error: %v, rsp: %s", err, lRsp)
		return 0, nil, errors.NewError(errors.ServerInternalError)
	}
	return http.StatusOK, rsp, nil
}
