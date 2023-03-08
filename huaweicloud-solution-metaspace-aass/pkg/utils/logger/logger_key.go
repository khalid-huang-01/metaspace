// Copyright (c) Huawei Technologies Co., Ltd. 2022-2022. All rights reserved.

// 日志键值
package logger

const TraceLogger = "trace_logger"

// 通用字段
const (
	Level      = "level"
	Timestamp  = "@timestamp"
	Caller     = "caller"
	Msg        = "msg"
	Stacktrace = "stacktrace"
	WorkNodeId = "work_node_id"
)

// 访问日志
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
	ServiceName    = "service_name"
	DurationMs     = "duration_ms"
	StartTime      = "start_time"
	EndTime        = "end_time"
)

// 业务流程日志
const (
	Stage          = "stage"
	Error          = "error"
	Success        = "success"
	AsyncTask      = "async_task"
	TaskRetryTimes = "task_retry_times"
	TaskLastError  = "task_last_err"
	MetricMonTask  = "metric_monitor_task"
)
