// Copyright (c) Huawei Technologies Co., Ltd. 2022-2022. All rights reserved.

// 服务相关常量定义
package constants

const (
	TimeFormatLayout = "2006-01-02T15:04:05Z"
)

// Build
const (
	BuildStateInitialized = "INITIALIZED"
	BuildStateReady       = "READY"
	BuildImageInitialized = "IMAGE_CREATING"
	BuildStateFailed      = "FAILED"
	BuildDeleted          = "DELETED"
)

// Fleet
const (
	InstanceTypeVm   = "VM"
	DefaultNameSpace = "Default"
)

const (
	UpdateScalingGroupUrlPattern   = "/v1/%s/instance-scaling-groups/%s"
	CreateScalingGroupUrlPattern   = "/v1/%s/instance-scaling-groups"
	DeleteScalingGroupUrlPattern   = "/v1/%s/instance-scaling-groups/%s"
	ListScalingGroupUrlPattern     = "/v1/%s/instance-scaling-groups"
	ShowScalingGroupUrlPattern     = "/v1/%s/instance-scaling-groups/%s"
	CreateScalingPolicyUrl         = "/v1/%s/scaling-policies"
	ScalingPolicyUrlPattern        = "/v1/%s/scaling-policies/%s"
	ScalingGroupLtsAccessConfig    = "/v1/%s/lts-access-config"
	ScalingGroupListAccessConfig   = "/v1/%s/list-lts-access-config"
	ScalingGroupLtsLogGroup        = "/v1/%s/lts-log-group"
	ScalingGroupListLogGroup       = "/v1/%s/list-lts-log-group"
	ScalingGroupLtsLogTransfer     = "/v1/%s/lts-transfer"
	ScalingGroupListLogTransfer    = "/v1/%s/list-lts-transfer"
	ServerSessionsUrl              = "/v1/server-sessions"
	ServerSessionUrlPattern        = "/v1/server-sessions/%s"
	ClientSessionsUrl              = "/v1/client-sessions"
	BatchCreateClientSessionUrl    = "/v1/client-sessions/batch-create"
	ClientSessionUrlPattern        = "/v1/client-sessions/%s"
	ProcessCountsUrl               = "/v1/app-process-counts"
	CreateResDomainUrl             = "/v3.0/OS-OPDomain/resdomain"
	CreateTokenUrl                 = "/v3/auth/tokens"
	CreateResUserUrl               = "/v3.0/OS-OPDomain/owner_user"
	ProcessesUrl                   = "/v1/app-processes"
	ServerProcessUrlPattern        = "/v1/app-processes/%s"
	APPGWMonitorFleetsUrl          = "/v1/monitor-fleets"
	APPGWMonitorInstancesUrl       = "/v1/monitor-instances"
	APPGWMonitorAppProcessesUrl    = "/v1/monitor-app-processes"
	APPGWMonitorServerSessionsUrl  = "/v1/monitor-server-sessions"
	AASSMonitorInstancesUrlPattern = "/v1/%s/monitor-instances"
)

const (
	NormalConsoleEndpoint  = "https://console.huaweicloud.com"
	UlanqabConsoleEndpoint = "https://console.ulanqab.huawei.com"
	ObsListPattern         = "/obs/manage/%s/object/list"
	LogStreamPattern       = "groupId=%s&groupName=%s&topidId=%s&topicName=%s"
)
