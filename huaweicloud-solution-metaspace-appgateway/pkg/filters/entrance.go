// Copyright (c) Huawei Technologies Co., Ltd. 2022-2022. All rights reserved.

// 输入过滤相关方法
package filters

import (
	"time"

	"github.com/beego/beego/v2/server/web/context"

	"codehub-g.huawei.com/videocloud/mediaprocesscenter/application-gateway/pkg/utils/log"
)

// EntranceFilter entrance filter
func EntranceFilter(ctx *context.Context) {
	var traceLogger *log.FMLogger

	requestId := ctx.Input.Header(log.RequestId)
	switch requestId {
	case "":
		traceLogger = log.RunLogger.WithField(log.RequestId, "unknown")
	default:
		traceLogger = log.RunLogger.WithField(log.RequestId, requestId)
	}

	ctx.Input.SetData(log.StartTime, time.Now())
	ctx.Input.SetData(log.RequestId, requestId)
	ctx.Input.SetData(log.TraceLogger, traceLogger)
}

// AppProcessEntranceFilter app process entrance filter
func AppProcessEntranceFilter(ctx *context.Context) {
	EntranceFilter(ctx)
	ctx.Input.SetData(log.ResourceType, log.ResourceAppProcess)
}

// ServerSessionEntranceFilter server session entrance filter
func ServerSessionEntranceFilter(ctx *context.Context) {
	EntranceFilter(ctx)
	ctx.Input.SetData(log.ResourceType, log.ResourceServerSession)
}

// ClientSessionEntranceFilter client session entrance filter
func ClientSessionEntranceFilter(ctx *context.Context) {
	EntranceFilter(ctx)
	ctx.Input.SetData(log.ResourceType, log.ResourceClientSession)
}

// InstanceConfigurationEntranceFilter instance configuration entrance filter
func InstanceConfigurationEntranceFilter(ctx *context.Context) {
	EntranceFilter(ctx)
	ctx.Input.SetData(log.ResourceType, log.ResourceInstanceConfiguration)
}

// MonitorEntranceFiltermonitor entrance filter
func MonitorEntranceFilter(ctx *context.Context) {
	EntranceFilter(ctx)
	ctx.Input.SetData(log.ResourceType, log.Monitor)
}
