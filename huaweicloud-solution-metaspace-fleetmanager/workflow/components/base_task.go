// Copyright (c) Huawei Technologies Co., Ltd. 2022-2022. All rights reserved.

// 任务流基本方法
package components

import (
	"encoding/json"
	"fleetmanager/logger"
	"fleetmanager/workflow/directer"
	"fleetmanager/workflow/meta"
	"fmt"
	"time"
)

type BaseTask struct {
	Logger             *logger.FMLogger
	Directer           directer.Directer
	Step               int
	Prev               int
	Next               int
	meta               meta.TaskMeta
	retryTimes         int
	rollbackRetryTimes int
}

// TaskStep 获取任务步骤
func (t *BaseTask) TaskStep() int {
	return t.Step
}

// LinkPrev 链接上一个任务
func (t *BaseTask) LinkPrev(prev int) {
	t.Prev = prev
}

// LinkNext 链接下一个任务
func (t *BaseTask) LinkNext(next int) {
	t.Next = next
}

// Execute 执行任务
func (t *BaseTask) Execute(*directer.ExecuteContext) (interface{}, error) {
	t.ExecNext(nil, nil)
	return nil, nil
}

// Rollback 回滚任务
func (t *BaseTask) Rollback(*directer.ExecuteContext) (output interface{}, err error) {
	t.RollbackPrev(output, err)
	return nil, nil
}

// NewBaseTask 新建基础任务
func NewBaseTask(meta meta.TaskMeta, directer directer.Directer, step int) BaseTask {
	log := directer.GetLogger().WithFields(map[string]interface{}{
		logger.TaskStep:        step,
		logger.TaskType:        meta.TaskType,
		logger.TaskName:        meta.Name,
		logger.TaskDescription: meta.Description,
	})
	t := BaseTask{
		Step:     step,
		Directer: directer,
		meta:     meta,
		Logger:   log,
	}

	return t
}

func (t *BaseTask) needRetry(d directer.WorkflowDirection) bool {
	if d == directer.PositiveDirection {
		if t.retryTimes < t.meta.ExecuteFailure.RetryPolicy.Repeat {
			t.retryTimes++
			return true
		}
	}

	if d == directer.NegativeDirection {
		if t.rollbackRetryTimes < t.meta.RollbackFailure.RetryPolicy.Repeat {
			t.rollbackRetryTimes++
			return true
		}
	}

	return false
}

func (t *BaseTask) getRetryDelay(d directer.WorkflowDirection) int {
	if d == directer.PositiveDirection {
		return t.meta.ExecuteFailure.RetryPolicy.GetRetryDelay(t.retryTimes)
	}

	if d == directer.NegativeDirection {
		return t.meta.RollbackFailure.RetryPolicy.GetRetryDelay(t.rollbackRetryTimes)
	}

	return 0
}

func (t *BaseTask) stepLog(ctx *directer.ExecuteContext) {
	success := 1
	if ctx.Err != nil {
		success = 0
	}
	t.Logger.WithFields(map[string]interface{}{
		logger.WorkflowDirection: ctx.Direction,
		logger.TaskRetryTimes:    t.retryTimes,
		logger.TaskRetryDelay:    t.getRetryDelay(ctx.Direction),
		logger.Success:           success,
		logger.Error:             fmt.Sprintf("%v", ctx.Err),
	}).Info("task step log")
}

// ExecNext 执行下一个任务
func (t *BaseTask) ExecNext(output interface{}, err error) {
	retryDelay := 0
	ctx := &directer.ExecuteContext{
		FromOutput: output,
		Err:        err,
		FromType:   t.meta.TaskType,
		From:       t.Step,
		RetryTimes: t.retryTimes,
		Direction:  directer.PositiveDirection,
	}

	if err != nil {
		if t.needRetry(ctx.Direction) {
			ctx.Next = t.Step
			retryDelay = t.getRetryDelay(ctx.Direction)
		} else {
			// err不为空，且重试达到最大次数，此时判断该任务是否允许跳过，若允许跳过，执行下个任务，否则执行回滚
			if t.meta.ExecuteFailure.Ignore {
				ctx.Next = t.Next
			} else {
				ctx.Next = t.Step
				ctx.Direction = directer.NegativeDirection
			}
		}
	} else {
		ctx.Next = t.Next
	}

	// 记录一下任务的运行日志
	t.stepLog(ctx)

	// 任务重试场景支持延迟重试
	if retryDelay > 0 {
		time.Sleep(time.Duration(retryDelay) * time.Second)
	}

	t.Directer.Process(ctx)
}

// RollbackPrev 回滚任务
func (t *BaseTask) RollbackPrev(output interface{}, err error) {
	retryDelay := 0
	ctx := &directer.ExecuteContext{
		FromOutput: output,
		Err:        err,
		FromType:   t.meta.TaskType,
		From:       t.Step,
		RetryTimes: t.retryTimes,
		Direction:  directer.NegativeDirection,
	}

	if err != nil {
		if t.needRetry(ctx.Direction) {
			ctx.Next = t.Step
			retryDelay = t.getRetryDelay(ctx.Direction)
		} else {
			// err不为空，且重试达到最大次数，此时判断该任务是否允许跳过，若允许跳过，执行下个回滚流程，否则结束工作流
			if t.meta.RollbackFailure.Ignore {
				ctx.Next = t.Prev
			} else {
				ctx.Next = 0
			}
		}
	} else {
		ctx.Next = t.Prev
	}

	// 记录一下任务的运行日志
	t.stepLog(ctx)

	// 任务重试场景支持延迟重试
	if retryDelay > 0 {
		time.Sleep(time.Duration(retryDelay) * time.Second)
	}

	t.Directer.Process(ctx)
}

// ParseInput 解析任务输入
func (t *BaseTask) ParseInput(input interface{}, parameter interface{}) error {
	b, ok := input.([]byte)
	if ok {
		return json.Unmarshal(b, parameter)
	}

	b, err := json.Marshal(input)
	if err != nil {
		return err
	}

	return json.Unmarshal(b, parameter)
}
