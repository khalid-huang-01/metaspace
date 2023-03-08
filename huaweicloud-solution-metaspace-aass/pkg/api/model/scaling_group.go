// Copyright (c) Huawei Technologies Co., Ltd. 2012-2018. All rights reserved.

package model
type CreateScalingGroupReq struct {
	CoolDownTime          *int32                 `json:"cool_down_time,omitempty" validate:"omitempty,gte=0,lte=30"`
	VmTemplate            *VmTemplate            `json:"vm_template,omitempty" validate:"required_if=InstanceType VM"`
	Name                  *string                `json:"instance_scaling_group_name" validate:"required,min=1,max=64"`
	SubnetId              *string                `json:"subnet_id" validate:"required,uuid"`
	VpcId                 *string                `json:"vpc_id" validate:"required,uuid"`
	InstanceType          *string                `json:"instance_type" validate:"required,oneof=VM"`
	InstanceConfiguration *InstanceConfiguration `json:"instance_configuration" validate:"required,dive"`
	EnterpriseProjectId	  *string				 `json:"enterprise_project_id" validate:"required"`
	FleetId               *string                `json:"fleet_id" validate:"required,uuid"`
	Agency                *Agency                `json:"agency" validate:"required,dive"`
	MinInstanceNumber     int32                  `json:"min_instance_number" validate:"gte=0,instanceMaximumLimit" reg_error_info:"The value is not within the valid range"`
	MaxInstanceNumber     int32                  `json:"max_instance_number" validate:"gte=0,instanceMaximumLimit" reg_error_info:"The value is not within the valid range"`
	DesireInstanceNumber  int32                  `json:"desire_instance_number" validate:"gtefield=MinInstanceNumber,ltefield=MaxInstanceNumber"`
	EnableAutoScaling     bool                   `json:"enable_auto_scaling,omitempty"`
	InstanceTags		  []InstanceTag		 	 `json:"instance_tags,omitempty" validate:"omitempty,min=0,max=10"`
	IamAgencyName		  *string				 `json:"iam_agency_name,omitempty" validate:"omitempty,min=0,max=64"`
}
 
type InstanceTag struct {
	Key		string `json:"key" validate:"min=1,max=36"`
	Value	string `json:"value" validate:"min=0,max=43"`
}

type Agency struct {
	Name     *string `json:"name" validate:"required,min=1,max=64"`
	DomainId *string `json:"domain_id" validate:"required,uuidWithoutHyphens" reg_error_info:"Incorrect format"`
}

type CreateScalingGroupResp struct {
	ScalingGroupId string `json:"instance_scaling_group_id"`
}

type UpdateScalingGroupReq struct {
	MinInstanceNumber     *int32                 `json:"min_instance_number,omitempty" validate:"omitempty,gte=0,instanceMaximumLimit" reg_error_info:"The value is not within the valid range"`
	MaxInstanceNumber     *int32                 `json:"max_instance_number,omitempty" validate:"omitempty,gte=0,instanceMaximumLimit" reg_error_info:"The value is not within the valid range"`
	DesireInstanceNumber  *int32                 `json:"desire_instance_number,omitempty" validate:"omitempty,gte=0,instanceMaximumLimit" reg_error_info:"The value is not within the valid range"`
	CoolDownTime          *int32                 `json:"cool_down_time,omitempty" validate:"omitempty,gte=0,lte=30"`
	EnableAutoScaling     *bool                  `json:"enable_auto_scaling,omitempty" validate:"omitempty"`
	InstanceConfiguration *InstanceConfiguration `json:"instance_configuration,omitempty" validate:"omitempty"`
	InstanceTags		  []InstanceTag			 `json:"instance_tags,omitempty" validate:"omitempty,min=0,max=10"`
}

type VmTemplate struct {
	KeyName            *string         `json:"key_name" validate:"required,min=1,max=64"`
	ImageID            *string         `json:"image_id" validate:"required,uuid"`
	AvailableFlavorIds []string        `json:"available_flavor_ids" validate:"required,min=1,max=10,unique,dive,flavorId" reg_error_info:"Incorrect format"`
	Disks              []Disk          `json:"disks" validate:"required,min=1,max=1,dive"`
	SecurityGroups     []SecurityGroup `json:"security_groups" validate:"required,min=1,max=10,dive"`
	Eip                Eip             `json:"eip" validate:"required,dive"`
}

type InstanceConfiguration struct {
	RuntimeConfiguration                    *RuntimeConfiguration `json:"runtime_configuration,omitempty" validate:"omitempty,dive"`
	ServerSessionProtectionPolicy           *string               `json:"server_session_protection_policy,omitempty" validate:"omitempty,oneof=NO_PROTECTION FULL_PROTECTION TIME_LIMIT_PROTECTION"`
	ServerSessionProtectionTimeLimitMinutes *int32                `json:"server_session_protection_time_limit_minutes,omitempty" validate:"omitempty,gte=5,lte=1440"`
}

type Disk struct {
	Size       *int32  `json:"size" validate:"required,gte=1,lte=1024"`
	VolumeType *string `json:"volume_type" validate:"required,supportedVolumeType" reg_error_info:"Volume type is not supported"`
	DiskType   *string `json:"disk_type" validate:"required,oneof=SYS"`
}

type SecurityGroup struct {
	Id *string `json:"id" validate:"required,uuid"`
}

type Eip struct {
	IpType    *string   `json:"ip_type,omitempty" validate:"omitempty,oneof=5_bgp 5_sbgp 5_telcom 5_union 5_g-vm"`
	Bandwidth Bandwidth `json:"bandwidth" validate:"required"`
}

type Bandwidth struct {
	Size *int32 `json:"size" validate:"required,gte=1,bandwidthMaximumLimit" reg_error_info:"The value is not within the valid range"`
}

type RuntimeConfiguration struct {
	ProcessConfigurations                 *[]ProcessConfiguration `json:"process_configurations,omitempty" validate:"omitempty,min=1,max=50,dive"`
	ServerSessionActivationTimeoutSeconds *int32                  `json:"server_session_activation_timeout_seconds,omitempty" validate:"omitempty,gte=1,lte=600"`
	MaxConcurrentServerSessionsPerProcess *int32                  `json:"max_concurrent_server_sessions_per_process,omitempty" validate:"omitempty,gte=1,lte=50"`
}

type ProcessConfiguration struct {
	LaunchPath           *string `json:"launch_path" validate:"required,min=0,max=1024"`
	Parameters           *string `json:"parameters" validate:"required,min=0,max=1024"`
	ConcurrentExecutions *int32  `json:"concurrent_executions" validate:"required,gte=1,lte=50"`
}

type ScalingGroupDetail struct {
	ID                   string `json:"instance_scaling_group_id"`
	Name                 string `json:"instance_scaling_group_name"`
	MinInstanceNumber    int32  `json:"min_instance_number"`
	MaxInstanceNumber    int32  `json:"max_instance_number"`
	DesireInstanceNumber int32  `json:"desire_instance_number"`
	CoolDownTime         int32  `json:"cool_down_time"`
	SubnetId             string `json:"subnet_id"`
	VpcId                string `json:"vpc_id"`
	FleetId              string `json:"fleet_id"`
	EnableAutoScaling    bool   `json:"enable_auto_scaling"`
}

type ScalingGroupList struct {
	Count                 int                  `json:"count"`
	InstanceScalingGroups []ScalingGroupDetail `json:"instance_scaling_groups"`
}
