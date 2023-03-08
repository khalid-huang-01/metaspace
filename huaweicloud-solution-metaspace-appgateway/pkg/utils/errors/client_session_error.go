// Copyright (c) Huawei Technologies Co., Ltd. 2022-2022. All rights reserved.

// 客户端会话异常定义
package errors

import (
	"fmt"
)

// NewCreateClientSessionError  client session创建错误
func NewCreateClientSessionError(message string, httpCode int) *ErrorResp {
	return NewError("SCASE.00010300", fmt.Sprintf("Create client session failed: %s.", message), httpCode)
}

// NewCreateClientSessionsError  批量创建client session的错误
func NewCreateClientSessionsError(message string, httpCode int) *ErrorResp {
	return NewError("SCASE.00010301", fmt.Sprintf("Create client sessions failed: %s.", message), httpCode)
}

// NewShowClientSessionError 展示client session时的错误
func NewShowClientSessionError(ID, message string, httpCode int) *ErrorResp {
	return NewError("SCASE.00010302", fmt.Sprintf("Read client sesssion %s failed: %s",
		ID, message), httpCode)
}

// NewClientSessionNotFoundError 没有查询到client session时的错误
func NewClientSessionNotFoundError(ID string, httpCode int) *ErrorResp {
	return NewError("SCASE.00010303", fmt.Sprintf("Client session %s not found. ",
		ID), httpCode)
}

// NewListClientSessionsError 列出client session列表时的错误
func NewListClientSessionsError(message string, httpCode int) *ErrorResp {
	return NewError("SCASE.00010304", fmt.Sprintf("List client session failed: %s", message), httpCode)
}

// NewUpdateClientSessionStateError  更新client session状态时的错误
func NewUpdateClientSessionStateError(ID, message string, httpCode int) *ErrorResp {
	return NewError("SCASE.00010305", fmt.Sprintf("Update client session %s state faile: %s",
		ID, message), httpCode)
}

// NewUpdateClientSessionError 更新client session时的错误
func NewUpdateClientSessionError(ID, message string, httpCode int) *ErrorResp {
	return NewError("SCASE.00010306", fmt.Sprintf("Update client session %s failed: %s",
		ID, message), httpCode)
}
