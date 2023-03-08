// Copyright (c) Huawei Technologies Co., Ltd. 2022-2022. All rights reserved.

// 查询检查方法
package common

import (
	"fmt"
	"net"
	"strconv"
	"time"

	"github.com/beego/beego/v2/server/web/context"

	"codehub-g.huawei.com/videocloud/mediaprocesscenter/application-gateway/pkg/utils/log"
)

const (
	MiniOffset    = 0
	MaxiOffset    = 2147483647
	DefaultOffset = 0
	MiniLimit     = 1
	MaxiLimit     = 100
	DefaultLimit  = 100
	MaxServerSessionNum = 250

	// DefaultSort 按创建时间降序
	DefaultSort = "-CREATED_AT"
	ASCSort     = "CREATED_AT"
)

const (
	ParamOffset          = "offset"
	ParamLimit           = "limit"
	ParamState           = "state"
	ParamServerSessionID = "server_session_id"
)

// CheckOffset 检查offset的合法性
func CheckOffset(ctx *context.Context) (int, error) {
	offset, err := strconv.Atoi(ctx.Input.Query(ParamOffset))
	if err != nil {
		log.RunLogger.Infof("[query checker] offset is not valid, use default offset %d", DefaultOffset)
		offset = DefaultOffset
	}

	if offset < MiniOffset || offset > MaxiOffset {
		return offset, fmt.Errorf("offset query must between %v and %v", MiniOffset, MaxiOffset)
	}

	return offset, nil
}

// CheckLimit 检查limit的合法性
func CheckLimit(ctx *context.Context) (int, error) {
	limit, err := strconv.Atoi(ctx.Input.Query(ParamLimit))
	if err != nil {
		log.RunLogger.Infof("[query checker] limit is not valid, use default limit %d", DefaultLimit)
		limit = DefaultLimit
	}

	if limit < MiniLimit || limit > MaxiLimit {
		return limit, fmt.Errorf("limit query must between %v and %v", MiniLimit, MaxiLimit)
	}

	return limit, nil
}

// CheckSort 检查sort的合法性
func CheckSort(ctx *context.Context) (string, error) {
	return DefaultSort, nil
}

func CheckState(ctx *context.Context) (string, error) {
	state := ctx.Input.Query(ParamState)
	if state == "" || state == ServerSessionStateError ||
		state == ServerSessionStateActive ||
		state == ServerSessionStateActivating ||
		state == ServerSessionStateTerminated {
		return state, nil
	}
	return "", fmt.Errorf("invalid state, please check")
}

// CheckServerSessionID 检查server session id的合法性
func CheckServerSessionID(ctx *context.Context) (string, error) {
	serverSessionID := ctx.Input.Query(ParamServerSessionID)
	if serverSessionID != "" {
		return serverSessionID, nil
	}
	return "", fmt.Errorf("invalid server session id, server session id is null ")
}

func CheckStateForAppProcess(ctx *context.Context) (string, error) {
	state := ctx.Input.Query(ParamState)
	if state == "" {
		log.RunLogger.Infof("[query checker] state is not set")
	} else if state != AppProcessStateActivating &&
		state != AppProcessStateActive &&
		state != AppProcessStateError &&
		state != AppProcessStateTerminated &&
		state != AppProcessStateTerminating &&
		state != "" {
		return "", fmt.Errorf("invalid state, please check")
	}
	return state, nil
}

func CheckStateforServerSession(ctx *context.Context) (string, error) {
	state := ctx.Input.Query(ParamState)
	if state == "" {
		log.RunLogger.Infof("[query checker] state is not set")
	} else if state != ServerSessionStateActive &&
		state != ServerSessionStateTerminated &&
		state != ServerSessionStateActivating &&
		state != ServerSessionStateError {
		return "", fmt.Errorf("invalid state, please check")
	}
	if state == ServerSessionStateActivating {
		state = ServerSessionStateActivating + "," + ServerSessionStateCreating
	}
	return state, nil
}

func CheckDuration(ctx *context.Context) (int, int, error) {
	duration_start := ctx.Input.Query(ParamStart)
	duration_end := ctx.Input.Query(ParamEnd)
	var start, end = 0, MaxServerSessionNum
	var err error
	if duration_start == "" {
		log.RunLogger.Infof("[query checker] duration start is not valid, not set")
	} else {
		start, err = strconv.Atoi(duration_start)
		if err != nil {
			return start, end, fmt.Errorf("invalid duration start, please check")
		}
	}
	if duration_end == "" {
		log.RunLogger.Infof("[query checker] duration end is not valid, not set")
	} else {
		end, err = strconv.Atoi(duration_end)
		if err != nil {
			return start, end, fmt.Errorf("invalid duration end, please check")
		}
	}
	if start < 0 || end < 0 || start > MaxServerSessionNum || end > MaxServerSessionNum {
		return start, end, fmt.Errorf("duration strat and end should be in [0, %d], please check", MaxServerSessionNum)
	}

	if start > end {
		return start, end, fmt.Errorf("duration start larger duration end, please check")
	}
	
	return start, end, nil
}

func CheckTime(ctx *context.Context) (time.Time, time.Time, error) {
	var start, end = ctx.Input.Query(ParamStartTime), ctx.Input.Query(ParamEndTime)
	var start_time = time.Time{}
	var end_time = time.Now().Local()
	var err error
	if start != "" {
		start_time, err = time.ParseInLocation(TimeLayout, start, time.Local)
		if err != nil {
			return start_time, end_time, fmt.Errorf("invalid start or end time, please check")
		}
	}
	if end != "" {
		end_time, err = time.ParseInLocation(TimeLayout, end, time.Local)
		if err != nil {
			return start_time, end_time, fmt.Errorf("invalid start or end time, please check")
		}
	}

	if start_time.After(end_time) {
		return start_time, end_time, fmt.Errorf("start time after end time, please check")
	}
	return start_time, end_time, nil
}

func CheckIpAddress(ipv4 string) bool {
	IpAddress := net.ParseIP(ipv4)
	return IpAddress != nil || ipv4 == ""
}

func StateChangeForServerSession(source_str string) string {
	if source_str == ServerSessionStateActivating ||
		source_str == ServerSessionStateCreating {
		return ServerSessionStateActivating
	}
	return source_str
}