// Copyright (c) Huawei Technologies Co., Ltd. 2022-2022. All rights reserved.

// 应用进程异常定义
package errors

import "fmt"

// NewCreateAppProcessError new create app process error
func NewCreateAppProcessError(message string, httpCode int) *ErrorResp {
	return NewError("SCASE.00010100", fmt.Sprintf("Create app process failed: %s.", message), httpCode)
}

// NewShowAppProcessError new show app process error
func NewShowAppProcessError(processID, message string, httpCode int) *ErrorResp {
	return NewError("SCASE.00010101", fmt.Sprintf("Read app process %s failed: %s.", processID, message), httpCode)
}

// NewAppProcessNotFoundError new app process not found error
func NewAppProcessNotFoundError(processID string, httpCode int) *ErrorResp {
	return NewError("SCASE.00010102", fmt.Sprintf("App process %s not found.", processID), httpCode)
}

// NewListAppProcessesError new list app process error
func NewListAppProcessesError(message string, httpCode int) *ErrorResp {
	return NewError("SCASE.00010103", fmt.Sprintf("List app processes failed: %s.", message), httpCode)
}

// NewUpdateAppProcessesError new update app process error
func NewUpdateAppProcessesError(message string, httpCode int) *ErrorResp {
	return NewError("SCASE.00010104", fmt.Sprintf("Update app processes failed: %s.", message), httpCode)
}

// NewUpdateAppProcessStateError new update app process state error
func NewUpdateAppProcessStateError(message string, httpCode int) *ErrorResp {
	return NewError("SCASE.00010105", fmt.Sprintf("Update app process state failed: %s.", message), httpCode)
}

// NewDeleteAppProcessError new delete app process error
func NewDeleteAppProcessError(message string, httpCode int) *ErrorResp {
	return NewError("SCASE.00010106", fmt.Sprintf("Update app process state failed: %s.", message), httpCode)
}
