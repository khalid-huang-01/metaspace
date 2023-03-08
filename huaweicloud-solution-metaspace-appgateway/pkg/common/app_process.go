// Copyright (c) Huawei Technologies Co., Ltd. 2022-2022. All rights reserved.

// 应用进程状态相关定义
package common

const (
	AppProcessStateActivating  = "ACTIVATING"
	AppProcessStateActive      = "ACTIVE"
	AppProcessStateTerminating = "TERMINATING"
	AppProcessStateTerminated  = "TERMINATED"
	AppProcessStateError       = "ERROR"
)

const AppProcessIDPrefix = "app-process-"

const (
	ParamName      = "name"
	ParamCreatorId = "creator_id"
	ParamFleetId   = "fleet_id"
	ParamStartTime = "start_time"
	ParamEndTime   = "end_time"
	ParamStart     = "duration_start"
	ParamEnd       = "duration_end"
	TimeLayout     = "2006-01-02T15:04:05"
)