// Copyright (c) Huawei Technologies Co., Ltd. 2022-2022. All rights reserved.

// 客户端会话结构定义
package apis

type ClientSession struct {
	ID              string `json:"client_session_id"`
	ServerSessionID string `json:"server_session_id"`
	ProcessID       string `json:"process_id"`
	InstanceID      string `json:"instance_id"`
	FleetID         string `json:"fleet_id"`
	State           string `json:"state"`
	PublicIP        string `json:"ip_address"`
	ClientPort      int    `json:"port"`
	ClientData      string `json:"client_data"`
	ClientID        string `json:"client_id"`
}

type Client struct {
	ClientData string `json:"client_data" validate:"omitempty,min=1,max=2048"`
	ClientID   string `json:"client_id" validate:"required,min=1,max=1024"`
}

type ClientSessionList struct {
	ClientSessions []ClientSession `json:"client_sessions"`
}

type CreateClientSessionRequest struct {
	ServerSessionID string `json:"server_session_id" validate:"required,min=1,max=128"`
	ClientData      string `json:"client_data" validate:"omitempty,min=1,max=128"`
	ClientID        string `json:"client_id" validate:"required,min=1,max=128"`
}

type CreateClientSessionResponse struct {
	ClientSession ClientSession `json:"client_session"`
}

type CreateClientSessionsRequest struct {
	ServerSessionID string   `json:"server_session_id" validate:"required,min=1,max=128"`
	Clients         []Client `json:"clients" validate:"required,min=1,max=25"`
}

type CreateClientSessionsResponse struct {
	ClientSessions []ClientSession `json:"client_sessions"`
}

type ShowClientSessionResponse struct {
	ClientSession ClientSession `json:"client_session"`
}

type ListClientSessionResponse struct {
	Count          int             `json:"count"`
	ClientSessions []ClientSession `json:"client_session"`
}

type UpdateClientSessionRequest struct {
	State string `json:"state" validate:"required,oneof=ACTIVE COMPLETED TIMEOUT"`
}

type UpdateClientSessionResponse struct {
	ClientSession ClientSession `json:"client_session"`
}

// UpdateClientSessionRequestForAuxProxy auxproxy调用该接口创建client session
type UpdateClientSessionRequestForAuxProxy struct {
	State string `json:"state" validate:"required,oneof=ACTIVE COMPLETED TIMEOUT RESERVED"`
}
