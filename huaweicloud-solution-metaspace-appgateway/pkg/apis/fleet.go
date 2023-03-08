// Copyright (c) Huawei Technologies Co., Ltd. 2022-2022. All rights reserved.

// fleet结构体定义
package apis

type FleetResponse struct {
	FleetID				string		`json:"fleet_id"`
	ProcessCount		int			`json:"process_count"`
	ServerSessionCount	int			`json:"server_session_count"`
	MaxServerSessionNum	int			`json:"max_server_session_num"`
}

type ListFleetsResponse struct {
	Count 				int				`json:"count"`
	Fleets				[]FleetResponse	`json:"fleets"`
}

type QueryFleetsInfoParam struct {
	FleetId				string
}