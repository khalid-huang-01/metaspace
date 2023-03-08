// Copyright (c) Huawei Technologies Co., Ltd. 2022-2022. All rights reserved.

// 开始创建
package update

import (
	"encoding/json"
	"fleetmanager/db/dao"
	"fleetmanager/workflow/components"
	"fleetmanager/workflow/directer"
	"fleetmanager/workflow/meta"
)

type StartFleetCreation struct {
	components.BaseTask
}

// Rollback 执行启动fleet创建任务回滚流程
func (t *StartFleetCreation) Rollback(*directer.ExecuteContext) (output interface{}, err error) {
	defer func() { t.RollbackPrev(output, err) }()
	f := &dao.Fleet{}
	fb := t.Directer.GetContext().Get(directer.WfKeyFleet).ToJson("{}")
	if err = json.Unmarshal(fb, f); err != nil {
		return nil, err
	}
	f.State = dao.FleetStateError
	if err = dao.GetFleetStorage().Update(f, "State"); err != nil {
		return nil, err
	}
	return nil, nil
}

// NewStartFleetCreation 新建启动fleet创建任务
func NewStartFleetCreation(meta meta.TaskMeta, directer directer.Directer, step int) components.Task {
	t := &StartFleetCreation{
		components.NewBaseTask(meta, directer, step),
	}

	return t
}
