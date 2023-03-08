// Copyright (c) Huawei Technologies Co., Ltd. 2022-2022. All rights reserved.

// 日志关键词
package log

const TraceLogger = "trace_logger"

// 通用字段
const (
	Level     = "level"
	Timestamp = "@timestamp"
	Caller    = "caller"
	Msg       = "msg"
)

// API访问日志和服务调用日志
const (
	RequestId    = "request_id"
	ResourceType = "resource_type"
	// ResponseStatus 表示
	ResponseStatus = "response_status"
	ResponseCode   = "response_code"
	ServiceName    = "service_name"
	StartTime      = "start_time"
	EndTime        = "end_time"
	DurationMs     = "duration_ms"
)

// 业务流程日志
const (
	Stage   = "stage"
	Error   = "error"
	Success = "success"
)
