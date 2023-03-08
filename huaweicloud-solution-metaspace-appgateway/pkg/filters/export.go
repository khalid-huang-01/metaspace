// Copyright (c) Huawei Technologies Co., Ltd. 2022-2022. All rights reserved.

// 输出过滤相关方法
package filters

import (
	"net/http"
	"time"

	"github.com/beego/beego/v2/server/web/context"

	"codehub-g.huawei.com/videocloud/mediaprocesscenter/application-gateway/pkg/utils/log"
)

// ExportFilter export filter
func ExportFilter(ctx *context.Context) {
	fields := make(map[string]interface{}, 10)

	var startTime time.Time
	endTime := time.Now()
	durationMs := int64(0)

	t := ctx.Input.GetData(log.StartTime)
	if v, ok := t.(time.Time); ok {
		startTime = v
		durationMs = endTime.Sub(startTime).Milliseconds()
	}

	statusCode := 0
	if v, ok := ctx.Input.GetData(log.ResponseCode).(int); ok {
		statusCode = v
	}

	status := 0
	if statusCode < http.StatusBadRequest && statusCode >= http.StatusOK {
		status = 1
	}

	fields[log.RequestId] = ctx.Input.GetData(log.RequestId)
	fields[log.ResponseStatus] = status
	fields[log.ResponseCode] = statusCode
	fields[log.ResourceType] = ctx.Input.GetData(log.ResourceType)

	fields[log.StartTime] = startTime
	fields[log.EndTime] = endTime
	fields[log.DurationMs] = durationMs

	log.AccessLogger.WithFields(fields).Debugf("api event")
}
