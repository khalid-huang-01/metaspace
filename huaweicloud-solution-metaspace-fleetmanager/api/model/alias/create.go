// Copyright (c) Huawei Technologies Co., Ltd. 2022-2022. All rights reserved.

// alias创建结构体定义
package alias

type CreateRequest struct {
	Name            string            `json:"name" validate:"required,min=1,max=1024"`
	Description     string            `json:"description" validate:"min=1,max=1024"`
	AssociatedFleets []AssociatedFleet `json:"associated_fleets,omitempty" validate:"omitempty,min=0,max=10,associatedFleetsNotDelicated,dive"`
	Type            string            `json:"type" validate:"required,oneof=ACTIVE DEACTIVE"`
	Message         string            `json:"message,omitempty" validate:"omitempty,min=0,max=1024"`
}

// CreateResponse: 创建别名返回响应结构
type CreateResponse struct {
	Alias Alias `json:"alias"`
}