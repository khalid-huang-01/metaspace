// Copyright (c) Huawei Technologies Co., Ltd. 2022-2022. All rights reserved.

// 日志
package logger

import (
	"fmt"

	"github.com/beego/beego/v2/server/web/context"
	"go.uber.org/zap"
)

var (
	R *FMLogger // 用来保存微服务产生的运行日志
	M *FMLogger // 用来保存微服务产生的统计日志，用以数据分析，暂时没有用
	A *FMLogger // 用来保存微服务被调用的日志，包括被调接口、返回值等
	C *FMLogger // 用来保存微服务访问外部服务的日志
	S *FMLogger // 用来保存微服务产生的安全日志
)

var (
	AtomicLevel        *zap.AtomicLevel
	GlobalCommonFields []zap.Field
)

func initAtomicLevel() *zap.AtomicLevel {
	a := zap.NewAtomicLevel()
	return &a
}

// Init inti logger
func Init() error {
	AtomicLevel = initAtomicLevel()
	// 初始化公共日志字段
	GlobalCommonFields = append(GlobalCommonFields)

	var err error
	if R, err = initLogger("./log/run/run.log", AtomicLevel, GlobalCommonFields...); err != nil {
		return fmt.Errorf("init run logger error %v", err)
	}
	if M, err = initLogger("./log/monitor/monitor.log", AtomicLevel, GlobalCommonFields...); err != nil {
		return fmt.Errorf("init monitor logger error %v", err)
	}
	if A, err = initLogger("./log/access/access.log", AtomicLevel, GlobalCommonFields...); err != nil {
		return fmt.Errorf("init access logger error %v", err)
	}
	if C, err = initLogger("./log/service-call/service_call.log",
		AtomicLevel, GlobalCommonFields...); err != nil {
		return fmt.Errorf("init service-call logger error %v", err)
	}
	if S, err = initLogger("./log/security/security.log",
		AtomicLevel, GlobalCommonFields...); err != nil {
		return fmt.Errorf("init security logger error %v", err)
	}

	return nil
}

// GetTraceLogger get trace logger from context
func GetTraceLogger(ctx *context.Context) *FMLogger {
	if tl := ctx.Input.GetData(TraceLogger); tl != nil {
		if tLogger, ok := tl.(*FMLogger); ok {
			return tLogger
		}
	}

	return R
}
