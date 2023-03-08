// Copyright (c) Huawei Technologies Co., Ltd. 2022-2022. All rights reserved.

// directer
package directer

import (
	"fleetmanager/logger"
)

type Directer interface {
	Process(ctx *ExecuteContext)
	GetLogger() *logger.FMLogger
	GetContext() *WorkflowContext
}
