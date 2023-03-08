// Copyright (c) Huawei Technologies Co., Ltd. 2022-2022. All rights reserved.

// aass创建弹性伸缩策略结构体定义
package policy

type CreateRequestToAASS struct {
	InstanceScalingGroupId string `json:"instance_scaling_group_id"`
	CreateRequest
}

type CreateResponseFromAASS struct {
	ScalingPolicyId string `json:"scaling_policy_id"`
}
