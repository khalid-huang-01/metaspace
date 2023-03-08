// Copyright (c) Huawei Technologies Co., Ltd. 2022-2022. All rights reserved.

// 客户端错误响应
package model

import (
	"github.com/huaweicloud/huaweicloud-sdk-go-v3/core/utils"
	"strings"
)

// Response Object
type ErrorResponse struct {
	ErrorCode string `json:"error_code"`
	ErrorMsg  string `json:"error_msg"`
}

// String 打印ErrorResponse信息
func (o ErrorResponse) String() string {
	data, err := utils.Marshal(o)
	if err != nil {
		return "ErrorResponse struct{}"
	}

	return strings.Join([]string{"ErrorResponse", string(data)}, " ")
}
