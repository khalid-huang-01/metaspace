package build

import (
	"fleetmanager/api/service/constants"
	"fleetmanager/db/dao"
	"fleetmanager/db/dbm"
	"fleetmanager/workflow/components"
	"fleetmanager/workflow/directer"
	"fleetmanager/workflow/meta"
)

type CreateBuildFinishTask struct {
	components.BaseTask
}

func (t *CreateBuildFinishTask) Execute(ctx *directer.ExecuteContext) (output interface{}, err error) {
	defer func() { t.ExecNext(output, err) }()
	w := &dao.Workflow{}
	buildId := t.Directer.GetContext().Get(directer.WfKeyBuildUuid).ToString("")
	err = dbm.Ormer.QueryTable(dao.WorkflowTable).Filter("resourceId", buildId).One(w)
	if err != nil {
		return nil, err
	}

	bd := &dao.Build{}
	err = dbm.Ormer.QueryTable(dao.BuildTable).Filter("Id", buildId).One(bd)
	if err != nil {
		return nil, err
	}

	// build状态为ready但workflow失败，则考虑为删除eip或ecs时失败，资源泄露，但镜像可用
	if w.State == dao.WorkflowStateError && bd.State != constants.BuildStateReady {
		bd.Id = buildId
		bd.State = constants.BuildStateFailed
		_, err := dbm.Ormer.Update(bd, "State")
		if err != nil {
			return nil, err
		}
	}

	return nil, nil
}

func NewCreateBuildFinishTask(meta meta.TaskMeta, directer directer.Directer, step int) components.Task {
	t := &CreateBuildFinishTask{
		components.NewBaseTask(meta, directer, step),
	}

	return t
}
