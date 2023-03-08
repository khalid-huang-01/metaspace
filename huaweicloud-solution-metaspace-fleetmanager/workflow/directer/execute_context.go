// Copyright (c) Huawei Technologies Co., Ltd. 2022-2022. All rights reserved.

// execute_context
package directer

type Direction int

type WorkflowDirection int

const (
	PositiveDirection WorkflowDirection = iota
	NegativeDirection
)

type ExecuteContext struct {
	From       int
	Next       int
	FromType   string
	FromOutput interface{}
	Err        error
	Direction  WorkflowDirection
	RetryTimes int
}

// Ended ctx结束
func (ctx *ExecuteContext) Ended() bool {
	return ctx.Next == 0
}
