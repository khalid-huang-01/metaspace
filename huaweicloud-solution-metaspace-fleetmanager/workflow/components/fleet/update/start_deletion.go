// Copyright (c) Huawei Technologies Co., Ltd. 2022-2022. All rights reserved.

// 开始删除
package update

import (
	"encoding/json"
	"fleetmanager/db/dao"
	"fleetmanager/resdomain/service"
	"fleetmanager/utils"
	"fleetmanager/workflow/components"
	"fleetmanager/workflow/directer"
	"fleetmanager/workflow/meta"
	"github.com/beego/beego/v2/client/orm"
)

type StartFleetDeletionTask struct {
	components.BaseTask
}

// Rollback 执行启动fleet删除任务回滚流程
func (t *StartFleetDeletionTask) Rollback(*directer.ExecuteContext) (output interface{}, err error) {
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

// Execute 执行启动Fleet删除任务
func (t *StartFleetDeletionTask) Execute(*directer.ExecuteContext) (output interface{}, err error) {
	defer func() { t.ExecNext(output, err) }()

	fleetId := t.Directer.GetContext().Get(directer.WfKeyFleetId).ToString("")
	if err = t.SyncResourceInfo(); err != nil {
		return nil, err
	}
	group, err := dao.GetScalingGroupStorage().GetOne(dao.Filters{"FleetId": fleetId})
	if err != nil {
		// 没有查询到,不需要删除
		if err == orm.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	t.Directer.GetContext().SetJson(directer.WfKeyScalingGroup, utils.ToJson(group))

	return nil, nil
}

// SyncResourceInfo 同步资源信息
func (t *StartFleetDeletionTask) SyncResourceInfo() error {
	domainAPI := service.ResDomain{}
	region := t.Directer.GetContext().Get(directer.WfKeyRegion).ToString("")
	originProjectId := t.Directer.GetContext().Get(directer.WfKeyOriginProjectId).ToString("")

	project, err := domainAPI.GetProject(originProjectId, region)
	if err != nil {
		return err
	}

	agency, err := domainAPI.GetAgency(project.OriginDomainId, region)
	if err != nil {
		return err
	}

	// 设置directer.WfKeyResourceProjectId信息
	if err = t.Directer.GetContext().Set(directer.WfKeyResourceProjectId, project.ResProjectId); err != nil {
		// 任务设置异常，资源回滚再重试 TODO(wangjun)
		return err
	}
	if err = t.Directer.GetContext().Set(directer.WfKeyResourceDomainId, project.ResDomainId); err != nil {
		// 任务设置异常，资源回滚再重试 TODO(wangjun)
		return err
	}
	if err = t.Directer.GetContext().Set(directer.WfKeyResourceAgencyName, agency.AgencyName); err != nil {
		// 任务设置异常，资源回滚再重试 TODO(wangjun)
		return err
	}
	if err = t.Directer.GetContext().Set(directer.WfKeyOriginDomainId, project.OriginDomainId); err != nil {
		// 任务设置异常，资源回滚再重试 TODO(wangjun)
		return err
	}

	return nil
}

// NewStartFleetDeletionTask 新建启动fleet删除任务
func NewStartFleetDeletionTask(meta meta.TaskMeta, directer directer.Directer, step int) components.Task {
	t := &StartFleetDeletionTask{
		components.NewBaseTask(meta, directer, step),
	}

	return t
}
