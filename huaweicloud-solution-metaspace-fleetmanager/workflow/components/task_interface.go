// Copyright (c) Huawei Technologies Co., Ltd. 2022-2022. All rights reserved.

// 任务流接口
package components

import "fleetmanager/workflow/directer"

type Task interface {
	TaskStep() int
	LinkPrev(prev int)
	LinkNext(next int)
	Execute(ctx *directer.ExecuteContext) (output interface{}, err error)
	Rollback(ctx *directer.ExecuteContext) (output interface{}, err error)
}
