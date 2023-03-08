// Copyright (c) Huawei Technologies Co., Ltd. 2022-2022. All rights reserved.

// 日志定义
package log

import (
	"fmt"

	"github.com/beego/beego/v2/server/web/context"
	"go.uber.org/zap"

	"codehub-g.huawei.com/videocloud/mediaprocesscenter/application-gateway/config"
)

const (
	RunLoggerPath      = "/home/appgateway/log/run/run.log"
	MonitorLoggerPath  = "/home/appgateway/log/monitor/monitor.log"
	AccessLoggerPath   = "/home/appgateway/log/access/access.log"
	ServiceLoggerPath  = "/home/appgateway/log/service_all/service_call.log"
	SecurityLoggerPath = "/home/appgateway/log/security/security.log"
)

const (
	ResourceAppProcess            = "appprocess"
	ResourceServerSession         = "serversession"
	ResourceClientSession         = "clientsession"
	ResourceInstanceConfiguration = "instance-configuration"
	Monitor 					  = "monitor"
)

var (
	Logger      *zap.Logger
	SugarLogger *zap.SugaredLogger

	// RunLogger 用来保存微服务（appgateway）产生的运行日志
	RunLogger *FMLogger
	// MonitorLogger 用来保存微服务（appgateway）产生的统计日志，用以数据分析，暂时没有用
	MonitorLogger *FMLogger
	// AccessLogger 用来保存微服务（appgateway）被调用的日志，包括被调接口、返回值等
	AccessLogger *FMLogger
	// ServiceLogger 用来保存微服务（appgateway）访问外部服务的日志
	ServiceLogger *FMLogger
	// SecurityLogger 用来保存安全有关服务的日志
	SecurityLogger *FMLogger
)

func init() {
	logger, err := zap.NewProduction()
	if err != nil {
		fmt.Printf("[logger] failed to new zap log for %v\n", err)
	}

	Logger = logger
	SugarLogger = logger.Sugar()
}

// GetTraceLogger get trace logger from context
func GetTraceLogger(ctx *context.Context) *FMLogger {
	if tl := ctx.Input.GetData(TraceLogger); tl != nil {
		if tLogger, ok := tl.(*FMLogger); ok {
			return tLogger
		}
	}

	return RunLogger
}

func GetDefaultRunLogger() *FMLogger {
	return RunLogger
}

// InitLog init loggers
func InitLog() error {
	atomicLevel := zap.NewAtomicLevel()
	if config.GlobalConfig.LogLevel == config.LogLevelDebug {
		atomicLevel.SetLevel(zap.DebugLevel)
	}

	var err error
	RunLogger, err = initLogger(RunLoggerPath, &atomicLevel)
	if err != nil {
		return fmt.Errorf("init run logger error %v", err)
	}

	AccessLogger, err = initLogger(AccessLoggerPath, &atomicLevel)
	if err != nil {
		return fmt.Errorf("init run logger error %v", err)
	}

	SecurityLogger, err = initLogger(SecurityLoggerPath, &atomicLevel)
	if err != nil {
		return fmt.Errorf("init security logger error %v", err)
	}

	return nil
}
