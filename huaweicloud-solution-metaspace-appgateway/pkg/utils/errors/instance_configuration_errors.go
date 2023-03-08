// Copyright (c) Huawei Technologies Co., Ltd. 2022-2022. All rights reserved.

// 实例配置异常
package errors

import "fmt"

// NewGetInstanceConfigurationError new get instance configuration error
func NewGetInstanceConfigurationError(message string, httpCode int) *ErrorResp {
	return NewError("SCASE.00010400", fmt.Sprintf("Get instance configuration failed: %s.", message), httpCode)
}
