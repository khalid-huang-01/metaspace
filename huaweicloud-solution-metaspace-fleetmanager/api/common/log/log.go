// Copyright (c) Huawei Technologies Co., Ltd. 2022-2022. All rights reserved.

// 创建系统日志
package log

import (
	"fleetmanager/logger"
	"github.com/beego/beego/v2/server/web/context"
)

// GetTraceLogger: 获取系统日志对象
func GetTraceLogger(ctx *context.Context) *logger.FMLogger {
	if tl := ctx.Input.GetData(logger.TraceLogger); tl != nil {
		if tLogger, ok := tl.(*logger.FMLogger); ok {
			return tLogger
		}
	}

	return logger.R
}
