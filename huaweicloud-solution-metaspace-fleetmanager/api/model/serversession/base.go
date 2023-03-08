// Copyright (c) Huawei Technologies Co., Ltd. 2022-2022. All rights reserved.

// 服务端会话基本结构体定义
package serversession

type Property struct {
	Key   string `json:"key" validate:"required,min=1,max=32"`
	Value string `json:"value" validate:"required,min=1,max=96"`
}

type ServerSession struct {
	ServerSessionId                         string     `json:"server_session_id"`
	Name                                    string     `json:"name"`
	CreatorId                               string     `json:"creator_id"`
	FleetId                                 string     `json:"fleet_id"`
	Properties                              []Property `json:"server_session_properties"`
	ServerSessionData                       string     `json:"server_session_data"`
	CurrentClientSessionCount               int        `json:"current_client_session_count"`
	MaxClientSessionCount                   int        `json:"max_client_session_count"`
	State                                   string     `json:"state"`
	StateReason                             string     `json:"state_reason"`
	IpAddress                               string     `json:"ip_address"`
	Port                                    int        `json:"port"`
	ClientSessionCreationPolicy             string     `json:"client_session_creation_policy"`
	ServerSessionProtectionPolicy           string     `json:"server_session_protection_policy"`
	ServerSessionProtectionTimeLimitMinutes int        `json:"server_session_protection_time_limit_minutes"`
}
