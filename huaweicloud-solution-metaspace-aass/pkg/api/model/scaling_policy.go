// Copyright (c) Huawei Technologies Co., Ltd. 2022-2022. All rights reserved.

// 弹性伸缩策略结构定义
package model

type CreateScalingPolicyReq struct {
	TargetConfiguration    *TargetConfiguration `json:"target_based_configuration,omitempty" validate:"required_if=Type TARGET_BASED"`
	Name                   *string              `json:"name" validate:"required,min=1,max=1024"`
	InstanceScalingGroupID *string              `json:"instance_scaling_group_id" validate:"required,uuid"`
	Type                   *string              `json:"policy_type" validate:"required,oneof=TARGET_BASED"`
}

type UpdateScalingPolicyReq struct {
	TargetConfiguration *TargetConfiguration `json:"target_based_configuration,omitempty" validate:"omitempty"`
	Name                *string              `json:"name,omitempty" validate:"omitempty,min=1,max=1024"`
}

type TargetConfiguration struct {
	MetricName  *string `json:"metric_name" validate:"required,oneof=PERCENT_AVAILABLE_SERVER_SESSIONS"`
	TargetValue *int32  `json:"target_value" validate:"required,gte=1,lte=100"`
}

type CreateScalingPolicyResp struct {
	ScalingPolicyId string `json:"scaling_policy_id"`
}
