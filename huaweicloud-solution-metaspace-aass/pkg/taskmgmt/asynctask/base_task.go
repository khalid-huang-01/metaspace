// Copyright (c) Huawei Technologies Co., Ltd. 2022-2022. All rights reserved.

// 基础任务
package asynctask

import "time"

const (
	// taskComplete means the task has completed its execution.
	taskComplete TaskStatus = "Complete"
	// taskFailed means the task has failed its execution.
	taskFailed TaskStatus = "Failed"
)

type TaskStatus string

type BaseTask struct {
	lastErr       error
	retryTime     int32
	taskStatus    TaskStatus
	lastStartTime time.Time
}

// IsComplete ...
func (t *BaseTask) IsComplete() bool {
	return t.taskStatus == taskComplete
}

// SetStatusComplete ...
func (t *BaseTask) SetStatusComplete() {
	t.taskStatus = taskComplete
}

// SetStatusFailed ...
func (t *BaseTask) SetStatusFailed(err error) {
	t.taskStatus = taskFailed
	t.lastErr = err
	t.retryTime++
}

// GetLastErr ...
func (t *BaseTask) GetLastErr() error {
	return t.lastErr
}

// GetRetryTimes ...
func (t *BaseTask) GetRetryTimes() int32 {
	return t.retryTime
}
