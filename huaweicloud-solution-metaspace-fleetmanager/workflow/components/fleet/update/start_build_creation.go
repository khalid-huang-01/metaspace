// Copyright (c) Huawei Technologies Co., Ltd. 2022-2022. All rights reserved.

// 开始创建
package update

import (
	"encoding/json"
	"fleetmanager/api/service/constants"
	"fleetmanager/db/dao"
	"fleetmanager/db/dbm"
	"fleetmanager/workflow/components"
	"fleetmanager/workflow/directer"
	"fleetmanager/workflow/meta"
	"github.com/beego/beego/v2/client/orm"
)

type StartBuildCreation struct {
	components.BaseTask
}

// Rollback 执行启动镜像打包任务回滚流程
func (t *StartBuildCreation) Rollback(*directer.ExecuteContext) (output interface{}, err error) {
	defer func() { t.RollbackPrev(output, err) }()
	b := &dao.Build{}
	bd := t.Directer.GetContext().Get(directer.WfKeyBuild).ToJson("{}")
	err = json.Unmarshal(bd, b)
	if err != nil {
		return nil, err
	}
	_, err = dbm.Ormer.QueryTable(dao.BuildTable).Filter("id", b.Id).Update(orm.Params{
		"State": constants.BuildStateFailed,
	})
	if err != nil {
		return nil, err
	}
	return nil, nil
}

// NewStartBuildCreation 新建应用包创建任务
func NewStartBuildCreation(meta meta.TaskMeta, directer directer.Directer, step int) components.Task {
	t := &StartBuildCreation{
		components.NewBaseTask(meta, directer, step),
	}

	return t
}
