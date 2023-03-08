// Copyright (c) Huawei Technologies Co., Ltd. 2022-2022. All rights reserved.

// 弹性伸缩组列表响应
package model

import (
	"github.com/huaweicloud/huaweicloud-sdk-go-v3/core/utils"
	"strings"
)

// Response Object
type ListScalingGroupResponse struct {
	Count                 int                    `json:"count"`
	InstanceScalingGroups []InstanceScalingGroup `json:"instance_scaling_groups"`
}

type InstanceScalingGroup struct {
	Id string `json:"instance_scaling_group_id"`
}

// ListScalingGroupResponse 打印ListScalingGroupResponse信息
func (o ListScalingGroupResponse) String() string {
	data, err := utils.Marshal(o)
	if err != nil {
		return "ListScalingGroupResponse struct{}"
	}

	return strings.Join([]string{"ListScalingGroupResponse", string(data)}, " ")
}
