// Copyright (c) Huawei Technologies Co., Ltd. 2022-2022. All rights reserved.

// 进程基本结构体定义
package process

type Process struct {
	Id                  string `json:"id"`
	Ip                  string `json:"ip"`
	Port                int    `json:"port"`
	State               string `json:"state"`
	ServerSessionCount  int    `json:"server_session_count"`
	MaxServerSessionNum int    `json:"max_server_session_num"`
}

// 接受appgw返回应用进程结构
type AppProcess struct {
	Id                                      string `json:"app_process_id"`
	PId                                     int    `json:"pid"`
	InstanceId                              string `json:"instance_id"`
	ScalingGroupId                          string `json:"scaling_group_id"`
	FleetId                                 string `json:"fleet_id"`
	IpAddress                               string `json:"ip_address"`
	PrivateIp                               string `json:"private_ip"`
	Port                                    int    `json:"port"`
	GrpcPort                                int    `json:"grpc_port"`
	AuxProxyPort                            int    `json:"aux_proxy_port"`
	LogPath                                 string `json:"log_path"`
	State                                   string `json:"state"`
	ServerSessionCount                      int    `json:"server_session_count"`
	MaxServerSessionNum                     int    `json:"max_server_session_num"`
	NewServerSessionProtectionPolicy        string `json:"new_server_session_protection_policy"`
	ServerSessionProtectionTimeLimitMinutes int    `json:"server_session_protection_time_limit_minutes"`
	ServerSessionActivationTimeoutSeconds   int    `json:"server_session_activation_timeout_seconds"`
}
