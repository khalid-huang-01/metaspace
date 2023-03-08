// Copyright (c) Huawei Technologies Co., Ltd. 2022-2022. All rights reserved.

// 服务端会话更新结构体定义
package serversession

type UpdateRequest struct {
	Name                                    *string `json:"name,omitempty" validate:"omitempty,min=1,max=1024"`
	MaxClientSessionCount                   *int    `json:"max_client_session_count,omitempty" validate:"omitempty,gte=1,lte=1024"`
	ServerSessionProtectionPolicy           *string `json:"server_session_protection_policy,omitempty" validate:"omitempty,oneof=NO_PROTECTION TIME_LIMIT_PROTECTION"`
	ServerSessionProtectionTimeLimitMinutes *int    `json:"server_session_protection_time_limit_minutes,omitempty" validate:"omitempty,gte=5,lte=1440"`
	ClientSessionCreationPolicy             *string `json:"client_session_creation_policy,omitempty" validate:"omitempty,oneof=ACCEPT_ALL DENY_ALL"`
}

type UpdateRequestToAppGW struct {
	Name                                    *string `json:"name,omitempty" validate:"omitempty,min=1,max=1024"`
	MaxClientSessionNum                     *int    `json:"max_client_session_num,omitempty" validate:"omitempty,gte=1,lte=1024"`
	ServerSessionProtectionPolicy           *string `json:"server_session_protection_policy,omitempty" validate:"omitempty,oneof=NO_PROTECTION TIME_LIMIT_PROTECTION"`
	ServerSessionProtectionTimeLimitMinutes *int    `json:"server_session_protection_time_limit_minutes,omitempty" validate:"omitempty,gte=5,lte=1440"`
	ClientSessionCreationPolicy             *string `json:"client_session_creation_policy,omitempty" validate:"omitempty,oneof=ACCEPT_ALL DENY_ALL"`
}
