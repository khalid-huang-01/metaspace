// Copyright (c) Huawei Technologies Co., Ltd. 2022-2022. All rights reserved.

// 应用进程结构定义
package apis

type AppProcess struct {
	ID                                      string `json:"app_process_id"`
	PID                                     int    `json:"pid"`
	BizPID                                  int    `json:"biz_pid"`
	InstanceID                              string `json:"instance_id"`
	ScalingGroupID                          string `json:"scaling_group_id"`
	FleetID                                 string `json:"fleet_id"`
	PublicIP                                string `json:"ip_address"`
	PrivateIP                               string `json:"private_ip"`
	ClientPort                              int    `json:"port"`
	GrpcPort                                int    `json:"grpc_port"`
	AuxProxyPort                            int    `json:"aux_proxy_port"`
	LogPath                                 string `json:"log_path"`
	State                                   string `json:"state"`
	ServerSessionCount                      int    `json:"server_session_count"`
	MaxServerSessionNum                     int    `json:"max_server_session_num"`
	NewServerSessionProtectionPolicy        string `json:"new_server_session_protection_policy"`
	ServerSessionProtectionTimeLimitMinutes int    `json:"server_session_protection_time_limit_minutes"`
	ServerSessionActivationTimeoutSeconds   int    `json:"server_session_activation_timeout_seconds"`
	LaunchPath                              string `json:"launch_path"`
	Parameters                              string `json:"parameters"`
}

type AppProcessList struct {
	AppProcesses []AppProcess `json:"app_processes"`
}

// Register app process regarding apis

type RegisterAppProcessRequest struct {
	PID                                     int    `json:"pid" validate:"required,gte=1,lte=32768"`
	BizPID                                  int    `json:"biz_pid" validate:"required,gte=1,lte=32768"`
	InstanceID                              string `json:"instance_id" validate:"required,min=1,max=128"`
	ScalingGroupID                          string `json:"scaling_group_id" validate:"required,min=1,max=128"`
	FleetID                                 string `json:"fleet_id" validate:"required,min=1,max=128"`
	PublicIP                                string `json:"ip_address" validate:"required,ip4_addr"`
	PrivateIP                               string `json:"private_ip" validate:"required,ip4_addr"`
	AuxProxyPort                            int    `json:"aux_proxy_port" validate:"required,gte=1,lte=65535"`
	MaxServerSessionNum                     int    `json:"max_server_session_num" validate:"required,gte=1,lte=50"`
	NewServerSessionProtectionPolicy        string `json:"new_server_session_protection_policy" validate:"required,oneof=NO_PROTECTION FULL_PROTECTION TIME_LIMIT_PROTECTION"`
	ServerSessionProtectionTimeLimitMinutes int    `json:"server_session_protection_time_limit_minutes" validate:"required,gte=5,lte=1440"`
	ServerSessionActivationTimeoutSeconds   int    `json:"server_session_activation_timeout_seconds" validate:"required,gte=1,lte=600"`
	LaunchPath                              string `json:"launch_path" validate:"required,gte=1,lte=2147483647"`
	Parameters                              string `json:"parameters" validate:"omitempty,min=0,max=2147483647"`
}

type RegisterAppProcessResponse struct {
	AppProcess AppProcess `json:"app_process"`
}

// Update app process regarding apis

type UpdateAppProcessRequest struct {
	ClientPort int      `json:"port" validate:"required,gte=1,lte=65535"`
	GrpcPort   int      `json:"grpc_port" validate:"required,gte=1,lte=65535"`
	LogPath    []string `json:"log_path" validate:"gte=0,lte=100"`
	State      string   `json:"state" validate:"required,oneof=AVTIVATING ACTIVE TERMINATING TERMINATED ERROR"`
}

type UpdateAppProcessResponse struct {
	AppProcess AppProcess `json:"app_process"`
}

// Update app process state

type UpdateAppProcessStateRequest struct {
	State string `json:"state" validate:"oneof=AVTIVATING ACTIVE TERMINATING TERMINATED ERROR"`
}

type UpdateAppProcessStateResponse struct {
	AppProcess AppProcess `json:"app_process"`
}

// Delete app process regarding apis

// Show app process regarding apis

type ShowAppProcessResponse struct {
	AppProcess AppProcess `json:"app_process"`
}

// List app process regarding apis

type ListAppProcessesResponse struct {
	Count        int          `json:"count"`
	AppProcesses []AppProcess `json:"app_processes"`
}

type ProcessCount struct {
	State string `json:"state"`
	Count int    `json:"count"`
}

type ShowAppProcessStatesResponse struct {
	FleetID       string         `json:"fleet_id"`
	ProcessCounts []ProcessCount `json:"process_counts"`
}
