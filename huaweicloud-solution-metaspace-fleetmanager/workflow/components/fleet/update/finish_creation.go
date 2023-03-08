// Copyright (c) Huawei Technologies Co., Ltd. 2022-2022. All rights reserved.

// 完成创建
package update

import (
	"fleetmanager/db/dao"
	"fleetmanager/workflow/components"
	"fleetmanager/workflow/directer"
	"fleetmanager/workflow/meta"
)

type FinishFleetCreationTask struct {
	components.BaseTask
}

// Execute 执行fleet创建结束任务
func (t *FinishFleetCreationTask) Execute(*directer.ExecuteContext) (output interface{}, err error) {
	defer func() { t.ExecNext(output, err) }()
	f := &dao.Fleet{
		Id:    t.Directer.GetContext().Get(directer.WfKeyFleetId).ToString(""),
		State: dao.FleetStateActive,
	}

	return nil, dao.GetFleetStorage().Update(f, "State")
}

// NewFinishFleetCreationTask 新建fleet创建结束任务
func NewFinishFleetCreationTask(meta meta.TaskMeta, directer directer.Directer, step int) components.Task {
	t := &FinishFleetCreationTask{
		components.NewBaseTask(meta, directer, step),
	}

	return t
}
