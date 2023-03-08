// Copyright (c) Huawei Technologies Co., Ltd. 2022-2022. All rights reserved.

// 准备资源账号
package resdomain

import (
	"fleetmanager/api/errors"
	"fleetmanager/resdomain/service"
	"fleetmanager/workflow/components"
	"fleetmanager/workflow/directer"
	"fleetmanager/workflow/meta"
	"github.com/beego/beego/v2/client/orm"
)

type PrepareResDomainTask struct {
	components.BaseTask
}

// Execute 执行资源账号同步任务
func (t *PrepareResDomainTask) Execute(ctx *directer.ExecuteContext) (output interface{}, err error) {
	defer func() { t.ExecNext(output, err) }()
	domainAPI := service.ResDomain{}
	region := t.Directer.GetContext().Get(directer.WfKeyRegion).ToString("")
	originProjectId := t.Directer.GetContext().Get(directer.WfKeyOriginProjectId).ToString("")

	// 判断是否需要创建资源租户
	user, err := domainAPI.GetUser(originProjectId, region)
	if err != nil {
		// 如果没有找到User
		if err == orm.ErrNoRows {
			// TODO(wj)：创建user
			return nil, errors.NewErrorF(errors.ServerInternalError,
				"do not support to create res domain")
		}
		return nil, err
	}

	// 获取资源租户ProjectID
	project, err := domainAPI.GetProject(originProjectId, region)
	if err != nil {
		return nil, err
	}

	// 获取资源租户委托名称
	agency, err := domainAPI.GetAgency(user.OriginDomainId, region)
	if err != nil {
		return nil, err
	}

	// 获取资源租户Keypair名称
	keypair, err := domainAPI.GetKeypair(user.OriginDomainId, region)
	if err != nil {
		return nil, err
	}

	// 设置directer.WfKeyResourceProjectId信息
	if err = t.Directer.GetContext().Set(directer.WfKeyResourceProjectId, project.ResProjectId); err != nil {
		// 任务设置异常，资源回滚再重试 TODO(wangjun)
		return nil, err
	}
	if err = t.Directer.GetContext().Set(directer.WfKeyResourceDomainId, project.ResDomainId); err != nil {
		// 任务设置异常，资源回滚再重试 TODO(wangjun)
		return nil, err
	}
	if err = t.Directer.GetContext().Set(directer.WfKeyResourceAgencyName, agency.AgencyName); err != nil {
		// 任务设置异常，资源回滚再重试 TODO(wangjun)
		return nil, err
	}
	if err = t.Directer.GetContext().Set(directer.WfKeyIamAgencyName, agency.IamAgencyName); err != nil {
		// 任务设置异常，资源回滚再重试 TODO(wangjun)
		return nil, err
	}
	if err = t.Directer.GetContext().Set(directer.WfKeyOriginDomainId, project.OriginDomainId); err != nil {
		// 任务设置异常，资源回滚再重试 TODO(wangjun)
		return nil, err
	}
	if err = t.Directer.GetContext().Set(directer.WfKeyResourceKeypairName, keypair.KeypairName); err != nil {
		// 任务设置异常，资源回滚再重试 TODO(wangjun)
		return nil, err
	}

	return nil, nil
}

// NewPrepareResDomainTask 新建准备资源租户任务
func NewPrepareResDomainTask(meta meta.TaskMeta, directer directer.Directer, step int) components.Task {
	t := &PrepareResDomainTask{
		components.NewBaseTask(meta, directer, step),
	}

	return t
}
