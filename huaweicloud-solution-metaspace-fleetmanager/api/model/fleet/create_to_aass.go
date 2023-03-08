// Copyright (c) Huawei Technologies Co., Ltd. 2022-2022. All rights reserved.

// aass相关信息构建结构体定义
package fleet

type Disk struct {
	Size       int    `json:"size"`
	VolumeType string `json:"volume_type"`
	DiskType   string `json:"disk_type"`
}

type SecurityGroup struct {
	Id string `json:"id"`
}

type BandWidth struct {
	Size      int    `json:"size"`
	ShareType string `json:"share_type"`
}

type Eip struct {
	BandWidth BandWidth `json:"bandwidth"`
	IpType string `json:"ip_type"`
}

type VmTemplate struct {
	AvailableFlavorIds []string        `json:"available_flavor_ids"`
	ImageId            string          `json:"image_id"`
	Disks              []Disk          `json:"disks"`
	SecurityGroups     []SecurityGroup `json:"security_groups"`
	Eip                Eip             `json:"eip"`
	KeyName            string          `json:"key_name"`
}

type InstanceConfiguration struct {
	RuntimeConfiguration                    RuntimeConfiguration `json:"runtime_configuration"`
	ServerSessionProtectionPolicy           string               `json:"server_session_protection_policy"`
	ServerSessionProtectionTimeLimitMinutes int                  `json:"server_session_protection_time_limit_minutes"`
}

type InstanceScalingGroup struct {
	FleetId               string                `json:"fleet_id"`
	MinInstanceNumber     int                   `json:"min_instance_number"`
	MaxInstanceNumber     int                   `json:"max_instance_number"`
	EnterpriseProjectId	  string				`json:"enterprise_project_id"`
	DesireInstanceNumber  int                   `json:"desire_instance_number"`
	CoolDownTime          int                   `json:"cool_down_time"`
	VpcId                 string                `json:"vpc_id"`
	SubnetId              string                `json:"subnet_id"`
	InstanceType          string                `json:"instance_type"`
	VmTemplate            VmTemplate            `json:"vm_template"`
	InstanceConfiguration InstanceConfiguration `json:"instance_configuration"`
	EnableAutoScaling     bool                  `json:"enable_auto_scaling"`
	Agency                Agency                `json:"agency"`
	Name                  string                `json:"instance_scaling_group_name"`
	IamAgencyName		  string			    `json:"iam_agency_name"`
	InstanceTags		  []InstanceTag			`json:"instance_tags"`
}

type Agency struct {
	Name     string `json:"name"`
	DomainId string `json:"domain_id"`
}

type CreateScalingGroupResponse struct {
	InstanceScalingGroupId string `json:"instance_scaling_group_id"`
}
