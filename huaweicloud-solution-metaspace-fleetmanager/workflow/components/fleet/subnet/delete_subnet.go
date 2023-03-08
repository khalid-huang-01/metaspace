// Copyright (c) Huawei Technologies Co., Ltd. 2022-2022. All rights reserved.

// 删除子网
package subnet

import (
	"fleetmanager/client"
	"fleetmanager/workflow/components"
	"fleetmanager/workflow/directer"
	"fleetmanager/workflow/meta"
)

type DeleteSubnetTask struct {
	components.BaseTask
}

// Execute 执行删除子网任务
func (t *DeleteSubnetTask) Execute(ctx *directer.ExecuteContext) (output interface{}, err error) {
	defer func() { t.ExecNext(output, err) }()
	resourceProjectId := t.Directer.GetContext().Get(directer.WfKeyResourceProjectId).ToString("")
	regionId := t.Directer.GetContext().Get(directer.WfKeyRegion).ToString("")
	agencyName := t.Directer.GetContext().Get(directer.WfKeyResourceAgencyName).ToString("")
	resourceDomainId := t.Directer.GetContext().Get(directer.WfKeyResourceDomainId).ToString("")
	subnetName := t.Directer.GetContext().Get(directer.WfKeySubnetName).ToString("")
	var subnetId = t.Directer.GetContext().Get(directer.WfKeySubnetId).ToString("")
	var vpcId = t.Directer.GetContext().Get(directer.WfKeyVpcId).ToString("")

	subnetId, vpcId, err = getSubnet(regionId, resourceProjectId, agencyName, resourceDomainId, subnetName, vpcId)
	if err != nil {
		return nil, err
	}

	if subnetId == ""{
		return nil, nil
	}

	// 删除subnet
	if err := deleteSubnet(regionId, resourceProjectId, agencyName, resourceDomainId, subnetId, vpcId); err != nil {
		return nil, err
	}

	// 等待subnet删除完成
	if err := waitSubnetDeleted(regionId, resourceProjectId, agencyName, resourceDomainId, subnetId); err != nil {
		return nil, err
	}

	return nil, nil
}

var getSubnet = func(regionId string, projectId string, agencyName string, resDomainId string,
	subnetName string, vpcId string) (string, string, error) {
	vpcClient, err := client.GetAgencyVpcClient(regionId, projectId, agencyName, resDomainId)
	if err != nil {
		return "", "", err
	}
	return client.GetSubnet(vpcClient, &subnetName, vpcId)
}

var deleteSubnet = func(regionId string, projectId string, agencyName string, resDomainId string,
	subnetId string, vpcId string) error {
	vpcClient, err := client.GetAgencyVpcClient(regionId, projectId, agencyName, resDomainId)
	if err != nil {
		return err
	}
	return client.DeleteSubnet(vpcClient, &subnetId, &vpcId)
}

var waitSubnetDeleted = func(regionId string, projectId string, agencyName string, resDomainId string,
	subnetId string) error {
	vpcClient, err := client.GetAgencyVpcClient(regionId, projectId, agencyName, resDomainId)
	if err != nil {
		return err
	}
	return client.WaitSubnetDeleted(vpcClient, &subnetId)
}

// NewDeleteSubnetTask 新建删除子网任务
func NewDeleteSubnetTask(meta meta.TaskMeta, directer directer.Directer, step int) components.Task {
	t := &DeleteSubnetTask{
		components.NewBaseTask(meta, directer, step),
	}

	return t
}
