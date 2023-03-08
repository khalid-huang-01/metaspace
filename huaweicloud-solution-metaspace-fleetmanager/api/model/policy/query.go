// Copyright (c) Huawei Technologies Co., Ltd. 2022-2022. All rights reserved.

// 弹性伸缩策略查询结构体定义
package policy

type ListResponse struct {
	TotalCount		int				`json:"total_count"`
	Count           int          	`json:"count"`
	ScalingPolicies []ScalingPolicy `json:"scaling_policies"`
}
