// Copyright (c) Huawei Technologies Co., Ltd. 2022-2022. All rights reserved.

// 准备vpc资源
package vpc

import (
	"fleetmanager/client"
	"fleetmanager/workflow/components"
	"fleetmanager/workflow/directer"
	"fleetmanager/workflow/meta"
)

var (
	clientCreateVpc          = client.CreateVpc
	clientGetAgencyVpcClient = client.GetAgencyVpcClient
)

type PrepareVpcTask struct {
	components.BaseTask
}

// Execute 执行vpc准备任务
func (t *PrepareVpcTask) Execute(ctx *directer.ExecuteContext) (output interface{}, err error) {
	defer func() { t.ExecNext(output, err) }()
	resourceProjectId := t.Directer.GetContext().Get(directer.WfKeyResourceProjectId).ToString("")
	regionId := t.Directer.GetContext().Get(directer.WfKeyRegion).ToString("")
	vpcName := t.Directer.GetContext().Get(directer.WfKeyVpcName).ToString("")
	vpcCidr := t.Directer.GetContext().Get(directer.WfKeyVpcCidr).ToString("")
	agencyName := t.Directer.GetContext().Get(directer.WfKeyResourceAgencyName).ToString("")
	resourceDomainId := t.Directer.GetContext().Get(directer.WfKeyResourceDomainId).ToString("")
	var vpcId = t.Directer.GetContext().Get(directer.WfKeyVpcId).ToString("")
	EnterpriseProject := t.Directer.GetContext().Get(directer.WfKeyEnterpriseProjectId).ToString("")
	if vpcId == "" {
		// 查询vpcId
		vpc, err := getVpc(regionId, resourceProjectId, agencyName, resourceDomainId, vpcName)
		if err != nil {
			return nil, err
		}
		if vpc != nil {
			vpcId = vpc.Id
		}
	}

	if vpcId == "" {
		// 1. 创建vpc
		vpcId, err = createVpc(regionId, resourceProjectId, agencyName, resourceDomainId, vpcName, vpcCidr, EnterpriseProject)
		if err != nil {
			return nil, err
		}
	}

	if err = t.Directer.GetContext().Set(directer.WfKeyVpcId, vpcId); err != nil {
		// 任务设置异常，资源回滚再重试 TODO(wangjun)
		return nil, err
	}

	// 3. 等待vpc创建完成
	if err := waitVpcReady(regionId, resourceProjectId, agencyName, resourceDomainId, vpcId); err != nil {
		return nil, err
	}

	return nil, nil
}

var createVpc = func(regionId string, projectId string, agencyName string, resDomainId string,
	vpcName string, vpcCidr string, enterpriseProject string) (string, error) {
	vpcClient, err := clientGetAgencyVpcClient(regionId, projectId, agencyName, resDomainId)
	if err != nil {
		return "", err
	}
	return clientCreateVpc(vpcClient, &vpcCidr, &vpcName, &enterpriseProject)
}

var waitVpcReady = func(regionId string, projectId string, agencyName string, resDomainId string, vpcId string) error {
	vpcClient, err := client.GetAgencyVpcClient(regionId, projectId, agencyName, resDomainId)
	if err != nil {
		return err
	}
	return client.WaitVpcReady(vpcClient, &vpcId)
}

// NewPrepareVpcTask 新建vpc任务
func NewPrepareVpcTask(meta meta.TaskMeta, directer directer.Directer, step int) components.Task {
	t := &PrepareVpcTask{
		components.NewBaseTask(meta, directer, step),
	}

	return t
}
