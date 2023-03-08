// Copyright (c) Huawei Technologies Co., Ltd. 2022-2022. All rights reserved.

// 日志入口
package logger

import (
	"context"
	"fleetmanager/utils"
	"fmt"
	"go.uber.org/zap"
)

var (
	R *FMLogger
	M *FMLogger
	A *FMLogger
	C *FMLogger
	S *FMLogger
)

var (
	AtomicLevel        *zap.AtomicLevel
	GlobalCommonFields []zap.Field
)

func initAtomicLevel() *zap.AtomicLevel {
	a := zap.NewAtomicLevel()
	return &a
}

// NewDebugLogger 获取debug logger
func NewDebugLogger() *FMLogger {
	return &FMLogger{
		logger: zap.NewNop().Sugar(),
		ctx:    context.Background(),
	}
}

// Init logger init
func Init() error {
	AtomicLevel = initAtomicLevel()

	// 初始化公共日志字段
	GlobalCommonFields = append(GlobalCommonFields,
		zap.Any(LocalIP, utils.GetLocalIP()),
	)

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
	if C, err = initLogger("./log/service_call/service_call.log",
		AtomicLevel, GlobalCommonFields...); err != nil {
		return fmt.Errorf("init service-call logger error %v", err)
	}
	if S, err = initLogger("./log/security/security.log",
		AtomicLevel, GlobalCommonFields...); err != nil {
		return fmt.Errorf("init security logger error %v", err)
	}

	return nil
}
