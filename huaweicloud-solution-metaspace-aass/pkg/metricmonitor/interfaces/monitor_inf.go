// Copyright (c) Huawei Technologies Co., Ltd. 2022-2022. All rights reserved.

// 监控接口
package interfaces

type MonitorInf interface {
	TaskIdForPolicy(policyId string) string
	DeleteTask(taskId string) error
}
