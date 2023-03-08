// Copyright (c) Huawei Technologies Co., Ltd. 2022-2022. All rights reserved.

// task_meta
package meta

import "math"

const (
	LogicFixed    = "fixed"
	LogicExponent = "exponent"
)

type RetryPolicy struct {
	Repeat       int    `json:"repeat"`
	Logic        string `json:"logic"`
	DelaySeconds int    `json:"delay_seconds"`
}

type OnFailure struct {
	RetryPolicy RetryPolicy `json:"retry_policy"`
	Ignore      bool        `json:"ignore"`
}

type TaskMeta struct {
	Name            string    `json:"name"`
	Description     string    `json:"description"`
	TaskType        string    `json:"task_type"`
	ExecuteFailure  OnFailure `json:"execute_failure"`
	RollbackFailure OnFailure `json:"rollback_failure"`
}

// GetRetryDelay 获取重试时间
func (p *RetryPolicy) GetRetryDelay(retryTimes int) int {
	switch p.Logic {
	case LogicFixed:
		return p.DelaySeconds
	case LogicExponent:
		return int(math.Pow(float64(p.DelaySeconds), float64(retryTimes)))
	default:
		return 0
	}
}
