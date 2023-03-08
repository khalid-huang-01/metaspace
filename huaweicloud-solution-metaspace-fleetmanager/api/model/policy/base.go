// Copyright (c) Huawei Technologies Co., Ltd. 2022-2022. All rights reserved.

// 弹性伸缩策略基本结构体定义
package policy

type TargetBasedConfiguration struct {
	MetricName  string `json:"metric_name" validate:"required,oneof=PERCENT_AVAILABLE_SERVER_SESSIONS"`
	TargetValue int    `json:"target_value" validate:"required,gte=1,lte=100"`
}

type ScalingPolicy struct {
	Id                       string                   `json:"policy_id"`
	Name                     string                   `json:"name"`
	FleetId                  string                   `json:"fleet_id"`
	PolicyType               string                   `json:"policy_type"`
	ScalingTarget            string                   `json:"scaling_target"`
	State                    string                   `json:"state"`
	TargetBasedConfiguration TargetBasedConfiguration `json:"target_based_configuration"`
}
