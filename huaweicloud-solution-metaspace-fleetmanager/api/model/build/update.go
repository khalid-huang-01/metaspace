// Copyright (c) Huawei Technologies Co., Ltd. 2022-2022. All rights reserved.

// 应用包更新结构体构建
package build

// UpdateRequest
// @Description:
type UpdateRequest struct {
	Name        string `json:"name" validate:"required,buildName"`
	Description string `json:"description" validate:"min=0,max=100"`
	Version     string `json:"version" validate:"required,min=1,max=50"`
}
