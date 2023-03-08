// Copyright (c) Huawei Technologies Co., Ltd. 2022-2022. All rights reserved.

// 删除vpc
package vpc

import (
	"fleetmanager/client"
	"fleetmanager/workflow/components"
	"fleetmanager/workflow/directer"
	"fleetmanager/workflow/meta"

	vpcmodel "github.com/huaweicloud/huaweicloud-sdk-go-v3/services/vpc/v2/model"
)

type DeleteVpcTask struct {
	components.BaseTask
}

// Execute 执行删除vpc任务
func (t *DeleteVpcTask) Execute(ctx *directer.ExecuteContext) (output interface{}, err error) {
	defer func() { t.ExecNext(output, err) }()
	resourceProjectId := t.Directer.GetContext().Get(directer.WfKeyResourceProjectId).ToString("")
	regionId := t.Directer.GetContext().Get(directer.WfKeyRegion).ToString("")
	agencyName := t.Directer.GetContext().Get(directer.WfKeyResourceAgencyName).ToString("")
	resourceDomainId := t.Directer.GetContext().Get(directer.WfKeyResourceDomainId).ToString("")
	vpcName := t.Directer.GetContext().Get(directer.WfKeyVpcName).ToString("")
	var vpcId = t.Directer.GetContext().Get(directer.WfKeyVpcId).ToString("")

	fleetId := t.Directer.GetContext().Get(directer.WfKeyFleetId).ToString("")
	if vpcId == "" {
		if vpcName != "" {
			vpc, err := getVpc(regionId, resourceProjectId, agencyName, resourceDomainId, vpcName)
			if err != nil {
				return nil, err
			}
			if vpc != nil {
				vpcId = vpc.Id
			}
		} else {
			// 当fleet在创建弹性伸缩组时失败，不会不会拿到VPC id与VPC name, 此时VPC已被创建，
			// 需要尝试根据fleetID作为VPC name获取vpc，若获取的到则需被删除
			vpc, err := getVpc(regionId, resourceProjectId, agencyName, resourceDomainId, fleetId)
			if err != nil {
				return nil, err
			}
			if vpc != nil {
				vpcId = vpc.Id
				vpcName = fleetId
			}
		}
	}
	// 根据VPC name不存在，但VPCID存在时，尝试根据VPCID获取vpc name，避免指定VPC创建fleet时，VPC被误删
	if vpcName == "" && vpcId != "" {
		vpc, err := GetVpcById(regionId, resourceProjectId, agencyName, resourceDomainId, vpcId)
		if err != nil {
			return nil, err
		}
		if vpc != nil {
			vpcName = vpc.Name
		}
	}

	if vpcId == "" || vpcName != fleetId {
		// 没有创建vpc, 或使用的指定的VPC创建的资源，不删除VPC
		return nil, nil
	}

	// 删除vpc
	if err := deleteVpc(regionId, resourceProjectId, agencyName, resourceDomainId, vpcId); err != nil {
		return nil, err
	}

	// 等待vpc删除完成
	if err := waitVpcDeleted(regionId, resourceProjectId, agencyName, resourceDomainId, vpcId); err != nil {
		return nil, err
	}

	return nil, nil
}

var getVpc = func(regionId string, projectId string, agencyName string, resDomainId string,
	vpcName string) (*vpcmodel.Vpc, error) {
	vpcClient, err := client.GetAgencyVpcClient(regionId, projectId, agencyName, resDomainId)
	if err != nil {
		return nil, err
	}
	vpc, err := client.GetVpc(vpcClient, &vpcName)
	return vpc, err
}

var GetVpcById = func(regionId string, projectId string, agencyName string, resDomainId string,
	vpcId string) (*vpcmodel.Vpc, error) {
	vpcClient, err := client.GetAgencyVpcClient(regionId, projectId, agencyName, resDomainId)
	if err != nil {
		return nil, err
	}
	vpc, err := client.GetVpcById(vpcClient, &vpcId)
	return vpc, err
}

var deleteVpc = func(regionId string, projectId string, agencyName string, resDomainId string, vpcId string) error {
	vpcClient, err := client.GetAgencyVpcClient(regionId, projectId, agencyName, resDomainId)
	if err != nil {
		return err
	}
	return client.DeleteVpc(vpcClient, &vpcId)
}

var waitVpcDeleted = func(regionId string, projectId string, agencyName string, resDomainId string, vpcId string) error {
	vpcClient, err := client.GetAgencyVpcClient(regionId, projectId, agencyName, resDomainId)
	if err != nil {
		return err
	}
	return client.WaitVpcDeleted(vpcClient, &vpcId)
}

// NewDeleteVpcTask 新建vpc删除任务
func NewDeleteVpcTask(meta meta.TaskMeta, directer directer.Directer, step int) components.Task {
	t := &DeleteVpcTask{
		components.NewBaseTask(meta, directer, step),
	}

	return t
}
