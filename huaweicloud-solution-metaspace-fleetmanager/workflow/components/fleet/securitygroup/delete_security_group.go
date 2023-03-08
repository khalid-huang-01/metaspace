// Copyright (c) Huawei Technologies Co., Ltd. 2022-2022. All rights reserved.

// 删除安全组
package securitygroup

import (
	"fleetmanager/client"
	"fleetmanager/workflow/components"
	"fleetmanager/workflow/directer"
	"fleetmanager/workflow/meta"
)

type DeleteSecurityGroupTask struct {
	components.BaseTask
}

// Execute 执行删除安全组任务
func (t *DeleteSecurityGroupTask) Execute(ctx *directer.ExecuteContext) (output interface{}, err error) {
	defer func() { t.ExecNext(output, err) }()
	resourceProjectId := t.Directer.GetContext().Get(directer.WfKeyResourceProjectId).ToString("")
	regionId := t.Directer.GetContext().Get(directer.WfKeyRegion).ToString("")
	securityGroupName := t.Directer.GetContext().Get(directer.WfKeySecurityGroupName).ToString("")
	agencyName := t.Directer.GetContext().Get(directer.WfKeyResourceAgencyName).ToString("")
	resourceDomainId := t.Directer.GetContext().Get(directer.WfKeyResourceDomainId).ToString("")
	var securityGroupId = t.Directer.GetContext().Get(directer.WfKeySecurityGroupId).ToString("")

	if securityGroupId == "" {
		// 查询securityGroupId
		securityGroupId, err = getSecurityGroup(regionId, resourceProjectId, agencyName, resourceDomainId,
			securityGroupName)
		if err != nil {
			return nil, err
		}
	}

	if securityGroupId == "" {
		return nil, nil
	}

	// 删除securityGroup
	if err := deleteSecurityGroup(regionId, resourceProjectId, agencyName, resourceDomainId,
		securityGroupId); err != nil {
		return nil, err
	}

	return nil, nil
}

var getSecurityGroup = func(regionId string, projectId string, agencyName string, resDomainId string,
	securityGroupName string) (string, error) {
	vpcClient, err := client.GetAgencyVpcClient(regionId, projectId, agencyName, resDomainId)
	if err != nil {
		return "", err
	}
	return client.GetSecurityGroup(vpcClient, &securityGroupName)
}

var deleteSecurityGroup = func(regionId string, projectId string, agencyName string, resDomainId string,
	securityGroupId string) error {
	vpcClient, err := client.GetAgencyVpcClient(regionId, projectId, agencyName, resDomainId)
	if err != nil {
		return err
	}
	return client.DeleteSecurityGroup(vpcClient, &securityGroupId)
}

// NewDeleteSecurityGroupTask 新建删除安全组任务
func NewDeleteSecurityGroupTask(meta meta.TaskMeta, directer directer.Directer, step int) components.Task {
	t := &DeleteSecurityGroupTask{
		components.NewBaseTask(meta, directer, step),
	}

	return t
}
