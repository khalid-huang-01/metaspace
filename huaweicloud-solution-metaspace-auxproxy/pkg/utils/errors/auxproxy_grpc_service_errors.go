// Copyright (c) Huawei Technologies Co., Ltd. 2022-2022. All rights reserved.

// grpc通信错误
package errors

import auxproxyservice "codehub-g.huawei.com/videocloud/mediaprocesscenter/auxproxy/pkg/utils/sdk/auxproxy_service"

// NewProcessReadyError 上报进程可用的错误
func NewProcessReadyError(message string) *auxproxyservice.Error {
	return &auxproxyservice.Error{
		ErrorCode: "SCASE.00020100",
		ErrorMsg:  message,
	}
}

// NewProcessEndingError 上报进程结束的错误
func NewProcessEndingError(message string) *auxproxyservice.Error {
	return &auxproxyservice.Error{
		ErrorCode: "SCASE.00020101",
		ErrorMsg:  message,
	}
}

// NewActivateServerSessionError 上报server session激活成功的错误
func NewActivateServerSessionError(message string) *auxproxyservice.Error {
	return &auxproxyservice.Error{
		ErrorCode: "SCASE.00020200",
		ErrorMsg:  message,
	}
}

// NewTerminateServerSessionError 上报server session结束的错误
func NewTerminateServerSessionError(message string) *auxproxyservice.Error {
	return &auxproxyservice.Error{
		ErrorCode: "SCASE.00020201",
		ErrorMsg:  message,
	}
}

// NewAcceptClientSessionError 上报接入新的client session的错误
func NewAcceptClientSessionError(message string) *auxproxyservice.Error {
	return &auxproxyservice.Error{
		ErrorCode: "SCASE.00020300",
		ErrorMsg:  message,
	}
}

// NewRemoveClientSessionError 上报client session离开的错误
func NewRemoveClientSessionError(message string) *auxproxyservice.Error {
	return &auxproxyservice.Error{
		ErrorCode: "SCASE.00020301",
		ErrorMsg:  message,
	}
}

// NewDescribeClientSessionsError 获取client session列表的错误
func NewDescribeClientSessionsError(message string) *auxproxyservice.Error {
	return &auxproxyservice.Error{
		ErrorCode: "SCASE.00020302",
		ErrorMsg:  message,
	}
}

// NewUpdateClientSessionCreationPolicyError 更新client session创建策略的错误
func NewUpdateClientSessionCreationPolicyError(message string) *auxproxyservice.Error {
	return &auxproxyservice.Error{
		ErrorCode: "SCASE.00020303",
		ErrorMsg:  message,
	}
}
