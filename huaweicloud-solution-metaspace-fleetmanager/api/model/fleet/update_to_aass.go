// Copyright (c) Huawei Technologies Co., Ltd. 2022-2022. All rights reserved.

// 更新aass信息结构体定义
package fleet

type UpdateScalingGroupRequest struct {
	Id                    *string                      `json:"-"`
	FleetId               *string                      `json:"fleet_id,omitempty"`
	MinInstanceNumber     *int                         `json:"min_instance_number,omitempty"`
	MaxInstanceNumber     *int                         `json:"max_instance_number,omitempty"`
	DesireInstanceNumber  *int                         `json:"desire_instance_number,omitempty"`
	CoolDownTime          *int                         `json:"cool_down_time,omitempty"`
	InstanceConfiguration *UpdateInstanceConfiguration `json:"instance_configuration,omitempty"`
	EnableAutoScaling     *bool                        `json:"enable_auto_scaling,omitempty"`
	InstanceTags      	  *[]InstanceTag           	   `json:"instance_tags,omitempty"`
}

type UpdateInstanceConfiguration struct {
	RuntimeConfiguration                    *UpdateRuntimeConfiguration `json:"runtime_configuration,omitempty"`
	ServerSessionProtectionPolicy           *string                     `json:"server_session_protection_policy,omitempty"`
	ServerSessionProtectionTimeLimitMinutes *int                        `json:"server_session_protection_time_limit_minutes,omitempty"`
}

type UpdateRuntimeConfiguration struct {
	ServerSessionActivationTimeoutSeconds *int                         `json:"server_session_activation_timeout_seconds,omitempty"`
	MaxConcurrentServerSessionsPerProcess *int                         `json:"max_concurrent_server_sessions_per_process,omitempty"`
	ProcessConfigurations                 []UpdateProcessConfiguration `json:"process_configurations,omitempty"`
}

type UpdateProcessConfiguration struct {
	LaunchPath           *string `json:"launch_path,omitempty" validate:"startswith=/local/app/|startswith=c:/local/app/|startswith=C:/local/app/,min=0,max=1024"`
	Parameters           *string `json:"parameters,omitempty" validate:"min=0,max=1024"`
	ConcurrentExecutions *int    `json:"concurrent_executions,omitempty" validate:"gte=1,lte=50"`
}