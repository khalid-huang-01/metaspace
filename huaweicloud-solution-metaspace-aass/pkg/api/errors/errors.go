// Copyright (c) Huawei Technologies Co., Ltd. 2022-2022. All rights reserved.

// 异常设置
package errors

import (
	"fmt"
)

type ErrorResp struct {
	ErrCode  ErrCode `json:"error_code,omitempty"`
	ErrMsg   string  `json:"error_msg,omitempty"`
	HttpCode int     `json:"-"`
}

// NewErrorResp get new ErrorResp
func NewErrorResp(errCode ErrCode) *ErrorResp {
	return &ErrorResp{
		ErrCode: errCode,
		ErrMsg:  errCode.Msg(),
	}
}

// NewErrorRespWithHttpCode get new ErrorResp with httpCode
func NewErrorRespWithHttpCode(errCode ErrCode, httpCode int) *ErrorResp {
	return &ErrorResp{
		ErrCode:  errCode,
		ErrMsg:   errCode.Msg(),
		HttpCode: httpCode,
	}
}

// NewErrorRespWithMessage get new ErrorResp with message
func NewErrorRespWithMessage(errCode ErrCode, msg string) *ErrorResp {
	return &ErrorResp{
		ErrCode: errCode,
		ErrMsg:  fmt.Sprintf("%s. Details: %s", errCode.Msg(), msg),
	}
}

// NewErrorRespWithHttpCodeAndMessage get new ErrorResp with httpCode and message
func NewErrorRespWithHttpCodeAndMessage(errCode ErrCode, httpCode int, msg string) *ErrorResp {
	return &ErrorResp{
		ErrCode:  errCode,
		ErrMsg:   fmt.Sprintf("%s. Details: %s", errCode.Msg(), msg),
		HttpCode: httpCode,
	}
}
