// Copyright (c) Huawei Technologies Co., Ltd. 2022-2022. All rights reserved.

// error基本操作
package errors

import (
	"fmt"
)

// NewErrorF: 基于错误信息新建CodedError
func NewErrorF(code ErrCode, format string, args ...interface{}) *CodedError {
	return &CodedError{
		ErrC: code,
		ErrD: fmt.Sprintf("%s,%s", code.Msg(), fmt.Sprintf(format, args...)),
	}
}

// NewError: 基于错误码新建CodedError
func NewError(code ErrCode) *CodedError {
	return &CodedError{
		ErrC: code,
		ErrD: code.Msg(),
	}
}

type CodedError struct {
	ErrC ErrCode `json:"error_code,omitempty"`
	ErrD string  `json:"error_msg,omitempty"`
}

// Code: 获取错误码
func (e *CodedError) Code() ErrCode {
	return e.ErrC
}

// Desc: 获取错误描述
func (e *CodedError) Desc() string {
	return e.ErrD
}

// Error: 打印错误信息
func (e *CodedError) Error() string {
	return fmt.Sprintf("code: %s msg: %s", e.ErrC, e.ErrD)
}
