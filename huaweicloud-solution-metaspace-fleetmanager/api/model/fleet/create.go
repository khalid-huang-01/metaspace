// Copyright (c) Huawei Technologies Co., Ltd. 2022-2022. All rights reserved.

// fleet创建结构体定义
package fleet

import "fleetmanager/setting"

type CreateRequest struct {
	Name                                    string                      `json:"name" validate:"min=1,max=1024"`
	Description                             string                      `json:"description" validate:"omitempty,min=1,max=1024"`
	BuildId                                 string                      `json:"build_id" validate:"min=1,max=128"`
	Region                                  string                      `json:"region" validate:"min=1,max=64"`
	Bandwidth                               int                         `json:"bandwidth" validate:"gte=1,lte=200000"`
	InstanceSpecification                   string                      `json:"instance_specification" validate:"oneof=scase.standard.4u8g scase.standard.8u16g"`
	ServerSessionProtectionPolicy           string                      `json:"server_session_protection_policy" validate:"oneof=NO_PROTECTION FULL_PROTECTION TIME_LIMIT_PROTECTION"`
	ServerSessionProtectionTimeLimitMinutes int                         `json:"server_session_protection_time_limit_minutes" validate:"gte=5,lte=1440"`
	RuntimeConfiguration                    RuntimeConfiguration        `json:"runtime_configuration" validate:"required,dive"`
	InboundPermissions                      []IpPermission              `json:"inbound_permissions" validate:"required,dive"`
	InstanceTags                            []InstanceTag               `json:"instance_tags" validate:"omitempty,min=0,max=10,scalingTagsNotDelicated,dive"`
	ResourceCreationLimitPolicy             ResourceCreationLimitPolicy `json:"resource_creation_limit_policy"`
	EnterpriseProjectId                     string                      `json:"enterprise_project_id" validate:"omitempty,max=64"`
	VPCId                                 	string                      `json:"vpc_id" validate:"omitempty,max=64"`
	SubnetId                                string                      `json:"subnet_id" validate:"omitempty,max=64"`
}

type FleetVpcAndSubnet struct {
	VpcId			string
	VpcName			string
	VpcCidr		    string
	SubnetId		string
	SubnetName		string
	SubnetCidr		string
	GatewayIp		string		
}

type CreateResponse struct {
	Fleet Fleet `json:"fleet"`
}

// NewCreateRequest: 新建创建fleet请求体
func NewCreateRequest() *CreateRequest {
	r := &CreateRequest{
		Region:                                  setting.DefaultFleetRegion,
		Bandwidth:                               setting.DefaultFleetBandwidth,
		InstanceSpecification:                   setting.DefaultFleetSpecification,
		ServerSessionProtectionPolicy:           setting.DefaultFleetProtectPolicy,
		ServerSessionProtectionTimeLimitMinutes: setting.DefaultFleetProtectTimeLimit,
		RuntimeConfiguration: RuntimeConfiguration{
			ServerSessionActivationTimeoutSeconds: setting.DefaultFleetSessionTimeoutSeconds,
			MaxConcurrentServerSessionsPerProcess: setting.DefaultFleetMaxSessionNumPerProcess,
		},
		ResourceCreationLimitPolicy: ResourceCreationLimitPolicy{
			PolicyPeriodInMinutes: setting.DefaultFleetPolicyPeriod,
			NewSessionsPerCreator: setting.DefaultFleetNewSessionNumPerCreator,
		},
		EnterpriseProjectId: setting.DefaultEnterpriseProject,
		VPCId:		             "",
		SubnetId: 				 "",
	}

	return r
}
