// Copyright (c) Huawei Technologies Co., Ltd. 2022-2022. All rights reserved.

// 删除vm任务
package asynctask

import (
	"scase.io/application-auto-scaling-service/pkg/cloudresource"
	"scase.io/application-auto-scaling-service/pkg/db"
	"scase.io/application-auto-scaling-service/pkg/utils/logger"
)

const (
	TaskTypeDelVmTask = db.TaskTypeDeleteVm
)

// DelVmTask vm删除任务，先关机，后删除
type DelVmTask struct {
	vmId      string
	projectId string
	BaseTask
}

// NewDelVmTask ...
func NewDelVmTask(vmId string, projectId string) *DelVmTask {
	return &DelVmTask{
		vmId:      vmId,
		projectId: projectId,
	}
}

// GetKey get id of the resource corresponding to the task
func (t *DelVmTask) GetKey() string {
	return t.vmId
}

// GetType get task type
func (t *DelVmTask) GetType() string {
	return TaskTypeDelVmTask
}

// Run run task
func (t *DelVmTask) Run(log *logger.FMLogger) error {
	resCtrl, err := cloudresource.GetResourceController(t.projectId)
	if err != nil {
		return err
	}

	// 1. 等待vm关闭
	if err = resCtrl.WaitVmShutoff(log, t.vmId); err != nil {
		return err
	}

	// 2. 删除vm
	if err = resCtrl.DeleteVm(log, t.vmId); err != nil {
		return err
	}

	// 3. 删除db中记录的vm删除任务
	if err = db.DeleteDeletingVm(t.vmId); err != nil {
		return err
	}
	return db.DeleteAsyncTask(t.GetType(), t.GetKey())
}
