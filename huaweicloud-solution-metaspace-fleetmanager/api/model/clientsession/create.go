// Copyright (c) Huawei Technologies Co., Ltd. 2022-2022. All rights reserved.

// 客户端会话创建结构体定义
package clientsession

type CreateRequest struct {
	Session
}

type BatchCreateRequest struct {
	Clients []Session `json:"clients" validate:"required,dive,min=1,max=25"`
}

type CreateRequestToAPPGW struct {
	ServerSessionId string `json:"server_session_id"`
	CreateRequest
}

type BatchCreateRequestToAPPGW struct {
	ServerSessionId string `json:"server_session_id"`
	BatchCreateRequest
}

type BatchCreateResponseFromAPPGW struct {
	Count          int                      `json:"count"`
	ClientSessions []ClientSessionFromAPPGW `json:"client_sessions"`
}

type CreateResponseFromAPPGW struct {
	ClientSession ClientSessionFromAPPGW `json:"client_session"`
}

type ClientSessionFromAPPGW struct {
	ClientSession
	ProcessId string `json:"process_id"`
}

type CreateResponse struct {
	ClientSession ClientSession `json:"client_session"`
}

type BatchCreateResponse struct {
	Count          int             `json:"count"`
	ClientSessions []ClientSession `json:"client_sessions"`
}

type ClientSession struct {
	ClientSessionId string `json:"client_session_id"`
	ServerSessionId string `json:"server_session_id"`
	FleetId         string `json:"fleet_id"`
	IpAddress       string `json:"ip_address"`
	Port            int    `json:"port"`
	ClientData      string `json:"client_data"`
	ClientId        string `json:"client_id"`
	State           string `json:"state"`
}
