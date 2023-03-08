// Copyright (c) Huawei Technologies Co., Ltd. 2022-2022. All rights reserved.

// 日志常量
package logger

const TraceLogger = "trace_logger"

// 通用字段
const (
	Level      = "level"
	Timestamp  = "@timestamp"
	Caller     = "caller"
	Msg        = "msg"
	Stacktrace = "stacktrace"
	LocalIP    = "local_ip"
)

// API访问日志和服务调用日志
const (
	RequestId      = "request_id"
	RequestRawUri  = "request_raw_uri"
	ClientIp       = "client_ip"
	ResourceName   = "resource_name"
	RequestMethod  = "request_method"
	RequestQuery   = "request_query"
	RequestBody    = "request_body"
	ResponseStatus = "response_status"
	ResponseCode   = "response_code"
	ResponseBody   = "response_body"
	ServiceName    = "service_name"
	DurationMs     = "duration_ms"
	StartTime      = "start_time"
	EndTime        = "end_time"
	WorkflowId     = "workflow_id"
	WorkNodeId     = "worknode_id"
)

// 业务流程日志
const (
	Stage             = "stage"
	Error             = "error"
	Success           = "success"
	WorkflowDirection = "workflow_direction"
	TaskStep          = "task_step"
	TaskType          = "task_type"
	TaskName          = "task_name"
	TaskDescription   = "task_description"
	TaskRetryTimes    = "task_retry_times"
	TaskRetryDelay    = "task_retry_delay"
)
