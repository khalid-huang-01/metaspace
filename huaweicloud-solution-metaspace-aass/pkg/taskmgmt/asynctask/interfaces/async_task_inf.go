// Copyright (c) Huawei Technologies Co., Ltd. 2022-2022. All rights reserved.

// 异步任务接口
package interfaces

import "scase.io/application-auto-scaling-service/pkg/utils/logger"

type AsyncTaskInf interface {
	Run(log *logger.FMLogger) error
	GetKey() string
	GetType() string
	SetStatusFailed(err error)
	SetStatusComplete()
	IsComplete() bool
	GetRetryTimes() int32
	GetLastErr() error
}
