// Copyright (c) Huawei Technologies Co., Ltd. 2022-2022. All rights reserved.

// 完成删除
package update

import (
	"fleetmanager/db/dao"
	"fleetmanager/workflow/components"
	"fleetmanager/workflow/directer"
	"fleetmanager/workflow/meta"
	"github.com/beego/beego/v2/client/orm"
	"time"
)

type FinishFleetDeleteTask struct {
	components.BaseTask
}

// Execute 执行结束删除fleet任务
func (t *FinishFleetDeleteTask) Execute(*directer.ExecuteContext) (output interface{}, err error) {
	defer func() { t.ExecNext(output, err) }()
	fleetId := t.Directer.GetContext().Get(directer.WfKeyFleetId).ToString("")

	f := &dao.Fleet{
		Id:              fleetId,
		State:           dao.FleetStateTerminated,
		Terminated:      true,
		TerminationTime: time.Now(),
	}

	// 释放vpc cidr
	if err := t.DeleteFleetVpcCidr(fleetId); err != nil {
		return nil, err
	}

	return nil, dao.GetFleetStorage().Update(f, "State", "Terminated", "TerminationTime")
}

// DeleteFleetVpcCidr 删除fleet vpc cidr
func (t *FinishFleetDeleteTask) DeleteFleetVpcCidr(fleetId string) error {
	fvc, err := dao.GetFleetVpcCidrByFleetId(fleetId)
	if err != nil {
		if err == orm.ErrNoRows {
			return nil
		}
		return err
	}

	// 删除vpc cidr
	if err := dao.DeleteFleetVpcCidr(fvc); err != nil {
		return err
	}

	return nil
}

// NewFinishFleetDeleteTask 新建删除fleet任务
func NewFinishFleetDeleteTask(meta meta.TaskMeta, directer directer.Directer, step int) components.Task {
	t := &FinishFleetDeleteTask{
		components.NewBaseTask(meta, directer, step),
	}

	return t
}
