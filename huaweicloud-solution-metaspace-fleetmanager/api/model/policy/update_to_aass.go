// Copyright (c) Huawei Technologies Co., Ltd. 2022-2022. All rights reserved.

// aass更新弹性伸缩策略结构体定义
package policy

type UpdateRequestToAASS struct {
	Name                     string                   `json:"name"`
	PolicyType               string                   `json:"policy_type"`
	TargetBasedConfiguration TargetBasedConfiguration `json:"target_based_configuration"`
}
