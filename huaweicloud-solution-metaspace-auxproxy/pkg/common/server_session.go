// Copyright (c) Huawei Technologies Co., Ltd. 2022. All rights reserved.

// 服务端会话定义
package common

const (
	ServerSessionStateActivating = "ACTIVATING"
	ServerSessionStateActive     = "ACTIVE"
	ServerSessionStateTerminated = "TERMINATED"
	ServerSessionStateError      = "ERROR"
)

const (
	ClientSessionCreationPolicyAcceptAll = "ACCEPT_ALL"
	ClientSessionCreationPolicyDenyAll   = "DENY_ALL"
)

const ServerSessionIDPrefix = "server-session-"

const (
	ProtectionPolicyNoProtection        = "NO_PROTECTION"
	ProtectionPolicyTimeLimitProtection = "TIME_LIMIT_PROTECTION"
	ProtectionPolicyFullProtection      = "FULL_PROTECTION"
)
