// Copyright (c) Huawei Technologies Co., Ltd. 2022-2022. All rights reserved.

// 常量参数定义
package params

const (
	ProjectId             = ":project_id"
	FleetId               = ":fleet_id"
	BuildId               = ":build_id"
	PolicyId              = ":policy_id"
	AliasId               = ":alias_id"
	ServerSessionId       = ":server_session_id"
	ClientSessionId       = ":client_session_id"
	QueryRegionId         = "region_id"
	QueryBucketKey        = "bucket_key"
	QueryOffset           = "offset"
	QueryLimit            = "limit"
	QueryFleetId          = "fleet_id"
	QueryFleetName        = "fleet_name"
	QueryServerSessionId  = "server_session_id"
	HeaderParameterToken  = "X-Auth-Token"
	QuerySort             = "sort"
	QueryState            = "state"
	QueryScalingGroupName = "instance_scaling_group_name"
	QueryName             = "name"
	QueryType             = "type"
	QueryAccessConfigId   = "access_config_id"
	QueryLogStreamId      = "log_stream_id"
)

const (
	DefaultSort    = "CREATED_AT%3Adesc"
	DefaultOffset  = "0"
	DefaultLimit   = "100"
	DefaultNumber  = 0
	MaxBuildNumber = 100
	MaxBuildSize   = 5 * 1024 * 1024 * 1024 // obs流式上传最大限制为5GB
)

const (
	Name          = "name"
	CreatorId     = "creator_id"
	State         = "state"
	StartTime     = "start_time"
	EndTime       = "end_time"
	Id            = "id"
	InstanceId    = "instance_id"
	ProcessId     = "process_id"
	IpAddress     = "ip_address"
	DurationStart = "duration_start"
	DurationEnd   = "duration_end"
)

const (
	HealthState    = "health_state"
	LifeCycleState = "life_cycle_state"
	Limit          = "limit"
)

const (
	ParamOffset = "offset"
	ParamLimit  = "limit"
	ParamStart  = "duration_start"
	ParamEnd    = "duration_end"
)
