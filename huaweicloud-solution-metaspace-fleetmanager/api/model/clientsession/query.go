// Copyright (c) Huawei Technologies Co., Ltd. 2022-2022. All rights reserved.

// 客户端会话查询结构体定义
package clientsession

type ShowResponse struct {
	ClientSession ClientSession `json:"client_session"`
}

type ShowResponseFromAPPGW struct {
	ClientSession ClientSessionFromAPPGW `json:"client_session"`
}

type ListResponse struct {
	Count          int             `json:"count"`
	ClientSessions []ClientSession `json:"client_sessions"`
}

type ListResponseFromAPPGW struct {
	Count          int                      `json:"count"`
	ClientSessions []ClientSessionFromAPPGW `json:"client_sessions"`
}
