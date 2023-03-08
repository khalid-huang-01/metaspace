// Copyright (c) Huawei Technologies Co., Ltd. 2022-2022. All rights reserved.

// 服务端会话创建结构体定义
package serversession

type CreateRequest struct {
	FleetId                 string     `json:"fleet_id,omitempty" validate:"omitempty,min=0,max=64"`
	CreatorId               string     `json:"creator_id" validate:"min=0,max=1024"`
	AliasId                 string     `json:"alias_id,omitempty" validate:"omitempty,min=0,max=64"`
	Name                    string     `json:"name" validate:"min=0,max=1024"`
	MaxClientSessionCount   int        `json:"max_client_session_count" validate:"required,gte=1,lte=1024"`
	IdempotencyToken        string     `json:"idempotency_token" validate:"min=0,max=48"`
	ServerSessionData       string     `json:"server_session_data" validate:"min=0,max=4096"`
	ServerSessionProperties []Property `json:"server_session_properties" validate:"omitempty,dive,min=0,max=16"`
}

type CreateServerSessionResponse struct {
	ServerSession ServerSession `json:"server_session"`
}

type CreateRequestToAppGW struct {
	FleetId                 string     `json:"fleet_id" validate:"required,min=1,max=64"`
	CreatorId               string     `json:"creator_id" validate:"min=0,max=1024"`
	Name                    string     `json:"name" validate:"min=0,max=1024"`
	MaxClientSessionNum     int        `json:"max_client_session_num" validate:"required,gte=1,lte=1024"`
	IdempotencyToken        string     `json:"idempotency_token" validate:"min=0,max=48"`
	ServerSessionData       string     `json:"server_session_data" validate:"min=0,max=4096"`
	ServerSessionProperties []Property `json:"server_session_properties" validate:"omitempty,dive,min=0,max=16"`
}

type CreateServerSessionResponseFromAppGW struct {
	ServerSession ServerSessionFromAppGW `json:"server_session"`
}

type ServerSessionFromAppGW struct {
	ServerSessionId                         string     `json:"server_session_id"`
	Name                                    string     `json:"name"`
	CreatorId                               string     `json:"creator_id"`
	FleetId                                 string     `json:"fleet_id"`
	Properties                              []Property `json:"server_session_properties"`
	ServerSessionData                       string     `json:"server_session_data"`
	CurrentClientSessionCount               int        `json:"client_session_count"`
	MaxClientSessionCount                   int        `json:"max_client_session_num"`
	State                                   string     `json:"state"`
	StateReason                             string     `json:"state_reason"`
	IpAddress                               string     `json:"ip_address"`
	Port                                    int        `json:"port"`
	ClientSessionCreationPolicy             string     `json:"client_session_creation_policy"`
	ServerSessionProtectionPolicy           string     `json:"server_session_protection_policy"`
	ServerSessionProtectionTimeLimitMinutes int        `json:"server_session_protection_time_limit_minutes"`
	CreationTime                            string     `json:"creation_time"`
	TerminationTime                         string     `json:"termination_time"`
}
