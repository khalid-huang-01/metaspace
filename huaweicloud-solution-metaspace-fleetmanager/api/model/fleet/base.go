// Copyright (c) Huawei Technologies Co., Ltd. 2022-2022. All rights reserved.

// fleet信息基本结构体定义
package fleet

type ResourceCreationLimitPolicy struct {
	PolicyPeriodInMinutes int `json:"policy_period_in_minutes" validate:"gte=1,lte=60"`
	NewSessionsPerCreator int `json:"new_sessions_per_creator" validate:"gte=1,lte=60"`
}

type ProcessConfiguration struct {
	LaunchPath           string `json:"launch_path" validate:"startswith=/local/app/|startswith=c:/local/app/|startswith=C:/local/app/,min=0,max=1024"`
	Parameters           string `json:"parameters" validate:"launchParameters"`
	ConcurrentExecutions int    `json:"concurrent_executions" validate:"gte=1,lte=50"`
}

type RuntimeConfiguration struct {
	ServerSessionActivationTimeoutSeconds int                    `json:"server_session_activation_timeout_seconds" validate:"gte=1,lte=600" default:"600"`
	MaxConcurrentServerSessionsPerProcess int                    `json:"max_concurrent_server_sessions_per_process" validate:"gte=1,lte=50" default:"1"`
	ProcessConfigurations                 []ProcessConfiguration `json:"process_configurations" validate:"required,dive,min=1,max=50"`
}

type IpPermission struct {
	Protocol string `json:"protocol" validate:"oneof=TCP UDP"`
	IpRange  string `json:"ip_range" validate:"cidr"`
	FromPort int32  `json:"from_port" validate:"gte=1025,lte=60000"`
	ToPort   int32  `json:"to_port" validate:"gte=1025,lte=60000,gtefield=FromPort"`
}

type Fleet struct {
	FleetId                                 string                      `json:"fleet_id"`
	Name                                    string                      `json:"name"`
	Description                             string                      `json:"description"`
	Region                                  string                      `json:"region"`
	State                                   string                      `json:"state"`
	BuildId                                 string                      `json:"build_id"`
	Bandwidth                               int                         `json:"bandwidth"`
	InstanceSpecification                   string                      `json:"instance_specification"`
	OperatingSystem                         string                      `json:"operating_system"`
	ServerSessionProtectionPolicy           string                      `json:"server_session_protection_policy"`
	ServerSessionProtectionTimeLimitMinutes int                         `json:"server_session_protection_time_limit_minutes"`
	EnableAutoScaling                       bool                        `json:"enable_auto_scaling"`
	ScalingIntervalMinutes                  int                         `json:"scaling_interval_minutes"`
	CreationTime                            string                      `json:"creation_time"`
	ResourceCreationLimitPolicy             ResourceCreationLimitPolicy `json:"resource_creation_limit_policy"`
	EnterpriseProjectId                     string                      `json:"enterprise_project_id"`
	InstanceTags                            []InstanceTag               `json:"instance_tags,omitempty"`
}

type InstanceTag struct {
	Key   string `json:"key" validate:"required,min=1,max=36"`
	Value string `json:"value" validate:"min=0,max=36"`
}
