// Copyright (c) Huawei Technologies Co., Ltd. 2022-2022. All rights reserved.

// fleet更新结构体定义
package fleet

type UpdateAttributesRequest struct {
	Name                                    *string                            `json:"name,omitempty" validate:"omitempty,min=1,max=1024"`
	Description                             *string                            `json:"description,omitempty" validate:"omitempty,min=0,max=1024"`
	ServerSessionProtectionPolicy           *string                            `json:"server_session_protection_policy,omitempty" validate:"omitempty,oneof=NO_PROTECTION FULL_PROTECTION TIME_LIMIT_PROTECTION"`
	ServerSessionProtectionTimeLimitMinutes *int                               `json:"server_session_protection_time_limit_minutes,omitempty" validate:"omitempty,gte=5,lte=1440"`
	EnableAutoScaling                       *bool                              `json:"enable_auto_scaling,omitempty"`
	ScalingIntervalMinutes                  *int                               `json:"scaling_interval_minutes,omitempty" validate:"omitempty,gte=1,lte=30"`
	ResourceCreationLimitPolicy             *UpdateResourceCreationLimitPolicy `json:"resource_creation_limit_policy,omitempty" validate:"omitempty,dive"`
	InstanceTags							*[]InstanceTag					   `json:"instance_tags,omitempty" validate:"omitempty,min=0,max=10,scalingTagsNotDelicated,dive"`				
}

type UpdateAttributesDao struct {
	Name                                    *string                            `json:"name,omitempty" validate:"omitempty"`
	Description                             *string                            `json:"description,omitempty" validate:"omitempty"`
	ServerSessionProtectionPolicy           *string                            `json:"server_session_protection_policy,omitempty" validate:"omitempty,oneof=NO_PROTECTION FULL_PROTECTION TIME_LIMIT_PROTECTION"`
	ServerSessionProtectionTimeLimitMinutes *int                               `json:"server_session_protection_time_limit_minutes,omitempty" validate:"omitempty"`
	EnableAutoScaling                       *bool                              `json:"enable_auto_scaling,omitempty"`
	ScalingIntervalMinutes                  *int                               `json:"scaling_interval_minutes,omitempty" validate:"omitempty"`
	ResourceCreationLimitPolicy             *UpdateResourceCreationLimitPolicy `json:"resource_creation_limit_policy,omitempty" validate:"omitempty,dive"`
	InstanceTags							*string					   			`json:"instance_tags,omitempty" validate:"omitempty"`				
}

type UpdateResourceCreationLimitPolicy struct {
	PolicyPeriodInMinutes *int `json:"policy_period_in_minutes,omitempty" validate:"omitempty,gte=1,lte=60"`
	NewSessionsPerCreator *int `json:"new_sessions_per_creator,omitempty" validate:"omitempty,gte=1,lte=60"`
}

type UpdateInboundPermissionRequest struct {
	// A collection of port settings to be added to the fleet resource.
	InboundPermissionAuthorizations []IpPermission `json:"inbound_permission_authorizations" validate:"omitempty,dive"`

	// A collection of port settings to be removed from the fleet resource.
	InboundPermissionRevocations []IpPermission `json:"inbound_permission_revocations" validate:"omitempty,dive"`
}

type UpdateRuntimeConfigurationRequest struct {
	ServerSessionActivationTimeoutSeconds *int                         `json:"server_session_activation_timeout_seconds,omitempty" validate:"omitempty,gte=1,lte=600"`
	MaxConcurrentServerSessionsPerProcess *int                         `json:"max_concurrent_server_sessions_per_process,omitempty" validate:"omitempty,gte=1,lte=50"`
	ProcessConfigurations                 []UpdateProcessConfiguration `json:"process_configurations,omitempty" validate:"omitempty,dive,min=1,max=50"`
}

type UpdateFleetCapacityRequest struct {
	Minimum *int `json:"minimum" validate:"required,gte=0,lte=20000"`
	Desired *int `json:"desired" validate:"required,gte=0,lte=20000,gtefield=Minimum"`
	Maximum *int `json:"maximum" validate:"required,gte=0,lte=20000,gtefield=Desired"`
}
