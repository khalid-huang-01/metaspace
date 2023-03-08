// Copyright (c) Huawei Technologies Co., Ltd. 2022. All rights reserved.

// 错误构造
package errors

import (
	"fmt"
	"net/http"
)

// ErrorResp error resp
type ErrorResp struct {
	ErrorCode string `json:"error_code"`
	ErrorMsg  string `json:"error_msg"`
	HttpCode  int    `json:"-"`
}

// NewSystemError new system error
func NewSystemError() *ErrorResp {
	return &ErrorResp{
		ErrorCode: fmt.Sprintf("%d", http.StatusInternalServerError),
		ErrorMsg:  "internal server error",
		HttpCode:  http.StatusInternalServerError,
	}
}

// NewBadRequestError new bad request error
func NewBadRequestError() *ErrorResp {
	return &ErrorResp{
		ErrorCode: fmt.Sprintf("%d", http.StatusInternalServerError),
		ErrorMsg:  "bad request",
		HttpCode:  http.StatusBadRequest,
	}
}

// NewError new error
func NewError(code string, msg string, httpCode int) *ErrorResp {
	return &ErrorResp{
		ErrorCode: code,
		ErrorMsg:  msg,
		HttpCode:  httpCode,
	}
}

// NewStartServerSessionError 创建激活server session失败的错误
func NewStartServerSessionError(message string, httpCode int) *ErrorResp {
	return NewError("SCASE.00020200", message, httpCode)
}

func NewAuthenticationError() *ErrorResp {
	return NewError("SCASE.00020009", "authentication error", http.StatusForbidden)
}
