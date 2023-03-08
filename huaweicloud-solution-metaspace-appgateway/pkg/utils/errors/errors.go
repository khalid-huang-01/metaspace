// Copyright (c) Huawei Technologies Co., Ltd. 2022-2022. All rights reserved.

// 错误定义
package errors

import (
	"fmt"
	"net/http"
)

// Application Gateway的业务错误码从SCASE.00010000开始
// SCASE代表表示服务名
// “."连接服务名与八位数字
// 八位数字表示具体的错误类型，其中前四位0001表示application gateway组件，后四位表示具体的错误
// 后四位划分：前两位表示资源类型，00表示系统类型的错误，01表示app process，02表示server session，03表示client session

// 综上所述
// application gateway的app process的错误码占用范围为：SCASE.00010100到SCASE.00010199，共100位
// application gateway的server session的错误码占用范围为：SCASE.00010200到SCASE.00010299，共100位
// application gateway的client session的错误码占用范围为：SCASE.00010300到SCASE.00010399，共100位

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

func NewAuthenticationError() *ErrorResp {
	return NewError("SCASE.00010009", "authentication error", http.StatusForbidden)
}

// NewError new error
func NewError(code string, msg string, httpCode int) *ErrorResp {
	return &ErrorResp{
		ErrorCode: code,
		ErrorMsg:  msg,
		HttpCode:  httpCode,
	}
}

type noEffect interface {
	NoEffect() bool
}

// IsNoEffect 判断是否是noeffect类型的错误
func IsNoEffect(err error) bool {
	to, ok := err.(noEffect)
	if ok {
		return to.NoEffect()
	} else {
		return false
	}
}
