// Copyright (c) Huawei Technologies Co., Ltd. 2022-2022. All rights reserved.

// 过滤出口
package export

import (
	"net/http"
	"time"

	"github.com/beego/beego/v2/server/web/context"

	"scase.io/application-auto-scaling-service/pkg/utils/logger"
)

// Filter filter response
func Filter(ctx *context.Context) {
	fields := make(map[string]interface{})
	endTime := time.Now()
	var startUnix int64 = 0
	var endUnix int64 = endTime.Unix()
	var durationMs int64 = 0
	t := ctx.Input.GetData(logger.StartTime)
	if startTime, ok := t.(time.Time); ok {
		durationMs = endTime.Sub(startTime).Milliseconds()
		startUnix = startTime.Unix()
	}

	status := 0
	statusCode := ctx.ResponseWriter.Status
	if statusCode < http.StatusBadRequest && statusCode >= http.StatusOK {
		status = 1
	}

	fields[logger.RequestId] = ctx.Input.GetData(logger.RequestId)
	fields[logger.RequestRawUri] = ctx.Input.URL()
	fields[logger.ClientIp] = ctx.Input.Context.Request.RemoteAddr
	fields[logger.ResourceName] = ctx.Input.URI()
	fields[logger.RequestMethod] = ctx.Input.Method()
	fields[logger.RequestQuery] = ctx.Input.Context.Request.URL.RawQuery
	fields[logger.RequestBody] = string(ctx.Input.RequestBody)
	fields[logger.ResponseStatus] = status
	fields[logger.ResponseCode] = statusCode
	fields[logger.ServiceName] = "AASS"
	fields[logger.StartTime] = startUnix
	fields[logger.EndTime] = endUnix
	fields[logger.DurationMs] = durationMs
	logger.A.WithFields(fields).Info("api event")
}
