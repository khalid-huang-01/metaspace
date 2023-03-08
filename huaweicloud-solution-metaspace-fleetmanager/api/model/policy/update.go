// Copyright (c) Huawei Technologies Co., Ltd. 2022-2022. All rights reserved.

// 弹性伸缩策略更新结构体定义
package policy

type UpdateRequest struct {
	Name                     *string                  `json:"name,omitempty" validate:"omitempty,min=1,max=1024"`
	TargetBasedConfiguration *TargetBasedConfiguration `json:"target_based_configuration,omitempty" validate:"omitempty,dive"`
}

// NewUpdateRequest: 新建更新策略请求
func NewUpdateRequest() *UpdateRequest {
	r := &UpdateRequest{
	}

	return r
}
