// Copyright (c) Huawei Technologies Co., Ltd. 2022-2022. All rights reserved

// 服务端会话异常
package errors

import "fmt"

// NewCreateServerSessionError 生成一个创建Server Session失败的错误
func NewCreateServerSessionError(message string, httpCode int) *ErrorResp {
	return NewError("SCASE.00010200", fmt.Sprintf("Create server session failed: %s.", message), httpCode)
}

// NewShowServerSessionError 生成一个查询Server Session详细信息失败的错误
func NewShowServerSessionError(id, message string, httpCode int) *ErrorResp {
	return NewError("SCASE.00010201", fmt.Sprintf("Read server session %s failed: %s.",
		id, message), httpCode)
}

// NewServerSessionNotFoundError 生成一个查询Server Session找不到的错误
func NewServerSessionNotFoundError(id string, httpCode int) *ErrorResp {
	return NewError("SCASE.00010202", fmt.Sprintf("Server session %s not found.", id), httpCode)
}

// NewListServerSessionsError 生成一个查询Server Session列表失败的错误
func NewListServerSessionsError(message string, httpCode int) *ErrorResp {
	return NewError("SCASE.00010203", fmt.Sprintf("List Server sessions failed: %s", message), httpCode)
}

// NewUpdateServerSessionError 生成一个更新Server Session失败的错误
func NewUpdateServerSessionError(id, message string, httpCode int) *ErrorResp {
	return NewError("SCASE.00010204", fmt.Sprintf("Update server session %s failed: %s",
		id, message), httpCode)
}

// NewUpdateServerSessionStateError 生成一个更新Server Session 状态失败的错误
func NewUpdateServerSessionStateError(id, message string, httpCode int) *ErrorResp {
	return NewError("SCASE.00010205", fmt.Sprintf("Update server session %s state failed: %s",
		id, message), httpCode)
}
