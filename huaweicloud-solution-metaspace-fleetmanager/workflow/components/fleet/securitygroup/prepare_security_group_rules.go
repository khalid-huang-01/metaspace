// Copyright (c) Huawei Technologies Co., Ltd. 2022-2022. All rights reserved.

// 准备安全组规则
package securitygroup

import (
	"encoding/json"
	"fleetmanager/client"
	"fleetmanager/db/dao"
	"fleetmanager/setting"
	"fleetmanager/workflow/components"
	"fleetmanager/workflow/directer"
	"fleetmanager/workflow/meta"
)

type PrepareSecurityGroupRulesTask struct {
	components.BaseTask
}

// Execute 执行准备安全组规则任务
func (t *PrepareSecurityGroupRulesTask) Execute(ctx *directer.ExecuteContext) (output interface{}, err error) {
	defer func() { t.ExecNext(output, err) }()
	resourceProjectId := t.Directer.GetContext().Get(directer.WfKeyResourceProjectId).ToString("")
	regionId := t.Directer.GetContext().Get(directer.WfKeyRegion).ToString("")
	securityGroupId := t.Directer.GetContext().Get(directer.WfKeySecurityGroupId).ToString("")
	agencyName := t.Directer.GetContext().Get(directer.WfKeyResourceAgencyName).ToString("")
	resourceDomainId := t.Directer.GetContext().Get(directer.WfKeyResourceDomainId).ToString("")

	// 创建客户指定的安全组规则，这部分安全组规则创建完成后需要写入fleet的安全组表
	inboundPermissionsString := t.Directer.GetContext().Get(directer.WfKeyInboundPermissions).ToJson("[]")
	var customerPermission []dao.InboundPermission
	if err := json.Unmarshal(inboundPermissionsString, &customerPermission); err != nil {
		return nil, err
	}
	err = createSecurityGroupRules(regionId, resourceProjectId, agencyName, resourceDomainId,
		securityGroupId, customerPermission, true)
	if err != nil {
		// 创建安全组失败，待重试
		return nil, err
	}

	// 内部使用的安全组规则，用于aux proxy 与app gateway通信使用（临时方案），这部分规则不入库
	internalPermissionStr := setting.Config.Get(setting.InternalInboundPermissions).ToJson("[]")
	var internalPermissions []dao.InboundPermission
	if err := json.Unmarshal(internalPermissionStr, &internalPermissions); err != nil {
		return nil, err
	}
	err = createSecurityGroupRules(regionId, resourceProjectId, agencyName, resourceDomainId,
		securityGroupId, internalPermissions, false)
	if err != nil {
		// 创建安全组失败，待重试
		return nil, err
	}

	return nil, nil
}

var createSecurityGroupRules = func(regionId string, projectId string, agencyName string, resDomainId string,
	securityGroupId string, inboundPermissions []dao.InboundPermission, insertDb bool) error {
	vpcClient, err := client.GetAgencyVpcClient(regionId, projectId, agencyName, resDomainId)
	if err != nil {
		return err
	}
	ingress := "ingress"
	eTherType := "IPV4"
	for _, permission := range inboundPermissions {
		id, err := client.CreateSecurityGroupRule(vpcClient, &securityGroupId, &ingress, &eTherType,
			&permission.Protocol, &permission.FromPort, &permission.ToPort, &permission.IpRange)
		if err != nil {
			return err
		}
		permission.Id = id
		permission.SecurityGroupId = securityGroupId
		if insertDb {
			if err = dao.GetPermissionStorage().InsertOrUpdate(&permission); err != nil {
				// 容许重复插入
				return err
			}
		}
	}

	return nil
}

// NewPrepareSecurityGroupRulesTask 新建准备安全组规则任务
func NewPrepareSecurityGroupRulesTask(meta meta.TaskMeta, directer directer.Directer, step int) components.Task {
	t := &PrepareSecurityGroupRulesTask{
		components.NewBaseTask(meta, directer, step),
	}

	return t
}
