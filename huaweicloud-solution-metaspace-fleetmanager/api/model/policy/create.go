// Copyright (c) Huawei Technologies Co., Ltd. 2022-2022. All rights reserved.

// 弹性伸缩策略结构体定义
package policy

type CreateRequest struct {
	Name                     string                   `json:"name" validate:"required,min=1,max=1024"`
	PolicyType               string                   `json:"policy_type" validate:"required,oneof=TARGET_BASED"`
	ScalingTarget            string                   `json:"scaling_target" validate:"required,oneof=INSTANCE"`
	TargetBasedConfiguration TargetBasedConfiguration `json:"target_based_configuration" validate:"required,dive"`
}

type CreateScalingPolicyResponse struct {
	Id      string `json:"id"`
	FleetId string `json:"fleet_id"`
	State   string `json:"state"`
	CreateRequest
}

// NewCreateRequest: 新建创建策略请求
func NewCreateRequest() *CreateRequest {
	r := &CreateRequest{
		ScalingTarget: "INSTANCE",
	}

	return r
}
