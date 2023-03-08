// Copyright (c) Huawei Technologies Co., Ltd. 2022-2022. All rights reserved.

// 重启任务
package reload

import (
	"github.com/pkg/errors"

	"scase.io/application-auto-scaling-service/pkg/db"
	"scase.io/application-auto-scaling-service/pkg/metricmonitor/interfaces"
	"scase.io/application-auto-scaling-service/pkg/taskmgmt"
	"scase.io/application-auto-scaling-service/pkg/taskmgmt/asynctask"
	"scase.io/application-auto-scaling-service/pkg/utils"
)

// AddAsyncTask 重启异步任务
func AddAsyncTask(monitor interfaces.MonitorInf, taskMgmt *taskmgmt.AsyncTaskMgmt, task *db.AsyncTask) error {
	if taskMgmt.IsTaskExist(task.TaskType, task.TaskKey) {
		return nil
	}
	switch task.TaskType {
	case db.TaskTypeScaleOutScalingGroup:
		taskConf := &db.ScaleOutTaskConf{}
		if err := utils.ToObject([]byte(task.TaskConf), taskConf); err != nil {
			return errors.Wrapf(err, "utils unmarshal task conf[%s] err", task.TaskConf)
		}
		taskMgmt.AddTask(asynctask.NewScaleOutTask(task.TaskKey, taskConf.TargetInstanceNumber))
	case db.TaskTypeScaleInScalingGroup:
		taskConf := &db.ScaleInTaskConf{}
		if err := utils.ToObject([]byte(task.TaskConf), taskConf); err != nil {
			return errors.Wrapf(err, "utils unmarshal task conf[%s] err", task.TaskConf)
		}
		taskMgmt.AddTask(asynctask.NewScaleInTask(task.TaskKey, taskConf.ScaleInInstanceIds))
	case db.TaskTypeDeleteVm:
		taskConf := &db.DeleteVmTaskConf{}
		if err := utils.ToObject([]byte(task.TaskConf), taskConf); err != nil {
			return errors.Wrapf(err, "utils unmarshal task conf[%s] err", task.TaskConf)
		}
		taskMgmt.AddTask(asynctask.NewDelVmTask(task.TaskKey, taskConf.ResProjectId))
	case db.TaskTypeDeleteScalingGroup:
		taskConf := &db.DeleteScalingGroupTaskConf{}
		if err := utils.ToObject([]byte(task.TaskConf), taskConf); err != nil {
			return errors.Wrapf(err, "utils unmarshal task conf[%s] err", task.TaskConf)
		}
		taskMgmt.AddTask(asynctask.NewDelScalingGroupTask(task.TaskKey, monitor))
	}
	return nil
}
