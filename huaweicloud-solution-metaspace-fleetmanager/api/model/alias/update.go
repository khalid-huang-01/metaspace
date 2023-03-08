// Copyright (c) Huawei Technologies Co., Ltd. 2022-2022. All rights reserved.

// alias查询结构体定义
package alias

type UpdateAliasRequest struct {
	Name            string                  `json:"name,omitempty" validate:"omitempty,min=1,max=1024"`
	Description     string                  `json:"description,omitempty" validate:"omitempty,min=0,max=1024"`
	AssociatedFleets []AssociatedFleet 		`json:"associated_fleets,omitempty" validate:"omitempty,min=0,max=10,associatedFleetsNotDelicated,dive"`
	Type            string                  `json:"type" validate:"required,oneof=ACTIVE DEACTIVE"`
	Message         string                  `json:"message,omitempty" validate:"omitempty,min=0,max=1024"`
}
