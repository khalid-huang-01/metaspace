// Copyright (c) Huawei Technologies Co., Ltd. 2022-2022. All rights reserved.

// 日志模块
package log

import (
	"fmt"

	"github.com/beego/beego/v2/server/web/context"
	"go.uber.org/zap"

	"codehub-g.huawei.com/videocloud/mediaprocesscenter/auxproxy/config"
)

var (
	RunLogger *FMLogger
)

// GetTraceLogger get trace logger from context
func GetTraceLogger(ctx *context.Context) *FMLogger {
	if tl := ctx.Input.GetData(TraceLogger); tl != nil {
		if tLogger, ok := tl.(*FMLogger); ok {
			return tLogger
		}
	}

	return RunLogger
}

// InitLog init loggers
func InitLog() error {
	atomicLevel := zap.NewAtomicLevel()
	if config.Opts.LogLevel == config.LogLevelDebug {
		atomicLevel.SetLevel(zap.DebugLevel)
	}

	var err error
	RunLogger, err = initLogger(config.RunLoggerPath, &atomicLevel)
	if err != nil {
		return fmt.Errorf("init run logger error %v", err)
	}

	return nil
}
