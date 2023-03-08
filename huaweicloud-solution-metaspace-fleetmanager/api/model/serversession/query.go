// Copyright (c) Huawei Technologies Co., Ltd. 2022-2022. All rights reserved.

// 服务端会话查询结构体定义
package serversession

type ShowServerSessionResponseFromAppGW struct {
	ServerSession ServerSessionFromAppGW `json:"server_session"`
}

type ShowServerSessionResponse struct {
	ServerSession ServerSession `json:"server_session"`
}

type ListServerSessionResponse struct {
	Count          int             `json:"count"`
	ServerSessions []ServerSession `json:"server_sessions"`
}

type ListServerSessionResponseFromAppGW struct {
	Count          int                      `json:"count"`
	ServerSessions []ServerSessionFromAppGW `json:"server_sessions"`
}
