// Copyright (c) Huawei Technologies Co., Ltd. 2022-2022. All rights reserved.

// API出口过滤
package export

import (
	"fleetmanager/logger"
	"github.com/beego/beego/v2/server/web/context"
	"time"
)

const (
	BadRequestCode    = 400
	NormalRequestCode = 200
)

// Filter: API出口过滤器
func Filter(ctx *context.Context) {
	fields := make(map[string]interface{}, 0)

	endTime := time.Now()
	var startUnix int64 = 0
	var endUnix = endTime.Unix()
	var durationMs int64 = 0
	t := ctx.Input.GetData(logger.StartTime)
	if startTime, ok := t.(time.Time); ok {
		durationMs = endTime.Sub(startTime).Milliseconds()
		startUnix = startTime.Unix()
	}

	status := 0
	code := ctx.Output.Status
	if code == 0 {
		code = ctx.Output.Context.ResponseWriter.Status
	}
	if code < BadRequestCode && code >= NormalRequestCode {
		status = 1
	}

	routerPattern := ""
	rp := ctx.Input.GetData("RouterPattern")
	if rp != nil {
		s, ok := rp.(string)
		if ok {
			routerPattern = s
		}
	}

	fields[logger.RequestId] = ctx.Input.GetData(logger.RequestId)
	fields[logger.RequestRawUri] = ctx.Input.URL()
	fields[logger.ClientIp] = ctx.Input.Context.Request.RemoteAddr
	fields[logger.ResourceName] = routerPattern
	fields[logger.RequestMethod] = ctx.Input.Method()
	fields[logger.RequestQuery] = ctx.Input.Context.Request.URL.RawQuery
	fields[logger.RequestBody] = string(ctx.Input.RequestBody)
	fields[logger.ResponseStatus] = status
	fields[logger.ResponseCode] = code
	fields[logger.ServiceName] = "fleet_manager"
	fields[logger.StartTime] = startUnix
	fields[logger.EndTime] = endUnix
	fields[logger.DurationMs] = durationMs
	logger.A.WithFields(fields).Info("api event")
}
