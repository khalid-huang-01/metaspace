// Copyright (c) Huawei Technologies Co., Ltd. 2022-2022. All rights reserved.

// 实例配置定义
package apis

type InstanceConfiguration struct {
	RuntimeConfiguration                    RuntimeConfiguration `json:"runtime_configuration"`
	ServerSessionProtectionPolicy           string               `json:"server_session_protection_policy"`
	ServerSessionProtectionTimeLimitMinutes int                  `json:"server_session_protection_time_limit_minutes"`
}

type RuntimeConfiguration struct {
	ServerSessionActivationTimeoutSeconds int                    `json:"server_session_activation_timeout_seconds"`
	ProcessConfiguration                  []ProcessConfiguration `json:"process_configurations"`
	MaxConcurrentServerSessionsPerProcess int                    `json:"max_concurrent_server_sessions_per_process"`
}

type ProcessConfiguration struct {
	LaunchPath           string `json:"launch_path"`
	Parameters           string `json:"parameters"`
	ConcurrentExecutions int    `json:"concurrent_executions"`
}

// Show runtime configuration regarding apis

type ShowInstanceConfigurationResponse struct {
	InstanceConfiguration
}
