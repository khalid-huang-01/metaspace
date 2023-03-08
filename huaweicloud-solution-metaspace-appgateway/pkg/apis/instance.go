// Copyright (c) Huawei Technologies Co., Ltd. 2022-2022. All rights reserved.

// instance结构体定义
package apis

type InstanceResponse struct {
	InstanceId 			string		`json:"instance_id"`
	IpAddress 			string		`json:"ip_address"`
	ServerSessionCount	int			`json:"server_session_count"`
	MaxServerSessionNum	int			`json:"max_server_session_num"`
}

type ListInstanceResponse struct {
	Count			int 				`json:"count"`
	Instances		[]InstanceResponse	`json:"instances"`
}

type QueryInstanceParam struct {
	FleetID 		string
}