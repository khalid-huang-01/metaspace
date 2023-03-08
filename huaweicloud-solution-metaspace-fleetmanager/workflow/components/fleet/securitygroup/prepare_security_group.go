// Copyright (c) Huawei Technologies Co., Ltd. 2022-2022. All rights reserved.

// 准备安全组
package securitygroup

import (
	"fleetmanager/client"
	"fleetmanager/workflow/components"
	"fleetmanager/workflow/directer"
	"fleetmanager/workflow/meta"
	vpcmodel "github.com/huaweicloud/huaweicloud-sdk-go-v3/services/vpc/v2/model"
)

const (
	FiltedGroupRule = "0.0.0.0/0"
)

var (
	clientCreateSecurityGroup     = client.CreateSecurityGroup
	clientGetAgencyVpcClient      = client.GetAgencyVpcClient
	clientGetSecurityGroupRules   = client.GetSecurityGroupRules
	clientDeleteSecurityGroupRule = client.DeleteSecurityGroupRule
)

type PrepareSecurityGroupTask struct {
	components.BaseTask
}

// Execute 执行准备安全组任务
func (t *PrepareSecurityGroupTask) Execute(ctx *directer.ExecuteContext) (output interface{}, err error) {
	defer func() { t.ExecNext(output, err) }()
	resourceProjectId := t.Directer.GetContext().Get(directer.WfKeyResourceProjectId).ToString("")
	regionId := t.Directer.GetContext().Get(directer.WfKeyRegion).ToString("")
	securityGroupName := t.Directer.GetContext().Get(directer.WfKeySecurityGroupName).ToString("")
	resourceDomainId := t.Directer.GetContext().Get(directer.WfKeyResourceDomainId).ToString("")
	var securityGroupId = t.Directer.GetContext().Get(directer.WfKeySecurityGroupId).ToString("")
	agencyName := t.Directer.GetContext().Get(directer.WfKeyResourceAgencyName).ToString("")
	EnterpriseProject := t.Directer.GetContext().Get(directer.WfKeyEnterpriseProjectId).ToString("")

	if securityGroupId == "" {
		// 查询securityGroupId
		securityGroupId, err = getSecurityGroup(regionId, resourceProjectId, agencyName, resourceDomainId,
			securityGroupName)
		if err != nil {
			return nil, err
		}
	}

	if securityGroupId == "" {
		// 1. 创建security group
		securityGroupId, err = createSecurityGroup(regionId, resourceProjectId, agencyName, resourceDomainId,
			securityGroupName, EnterpriseProject)
		if err != nil {
			// 创建安全组失败，待重试
			return nil, err
		}
	}

	// 2. 查询多余的规则 (华为云创建安全组v2接口有Bug, 会开通多余的0.0.0.0/0:22/3389访问规则, 需要开发者自己删除)
	var toRemovedRuleIds *[]string
	toRemovedRuleIds, err = findSecurityGroupRules(regionId, resourceProjectId, agencyName, resourceDomainId,
		securityGroupId, FiltedGroupRule)
	if err != nil {
		// 查询待删除的安全组规则失败, 待重试
		return nil, err
	}

	// 3. 删除多余的规则
	if err = deleteRules(regionId, resourceProjectId, agencyName, resourceDomainId, *toRemovedRuleIds); err != nil {
		return nil, err
	}

	// 4. 更新securityGroupId记录
	if err = t.Directer.GetContext().Set(directer.WfKeySecurityGroupId, securityGroupId); err != nil {
		// 任务设置异常，资源回滚再重试 TODO(wangjun)
		return nil, err
	}

	return nil, nil
}

var createSecurityGroup = func(regionId string, projectId string, agencyName string, resDomainId string,
	securityGroupName string, enterpriseProject string) (string, error) {
	vpcClient, err := clientGetAgencyVpcClient(regionId, projectId, agencyName, resDomainId)
	if err != nil {
		return "", err
	}
	return clientCreateSecurityGroup(vpcClient, &securityGroupName, &enterpriseProject)
}

var findSecurityGroupRules = func(regionId string, projectId string, agencyName string, resDomainId string,
	securityGroupId string, filtedGroupRule string) (*[]string, error) {
	var vpcClient, err = clientGetAgencyVpcClient(regionId, projectId, agencyName, resDomainId)
	if err != nil {
		return nil, err
	}

	var rules *[]vpcmodel.SecurityGroupRule
	rules, err = clientGetSecurityGroupRules(vpcClient, &securityGroupId)
	if err != nil {
		return nil, err
	}

	var ruleIds []string
	for _, value := range *rules {
		if value.RemoteIpPrefix == filtedGroupRule {
			ruleIds = append(ruleIds, value.Id)
		}
	}

	return &ruleIds, nil
}

var deleteRules = func(regionId string, projectId string, agencyName string, resDomainId string,
	ruleIds []string) error {
	var vpcClient, err = clientGetAgencyVpcClient(regionId, projectId, agencyName, resDomainId)
	if err != nil {
		return err
	}

	for _, ruleId := range ruleIds {
		if err = clientDeleteSecurityGroupRule(vpcClient, ruleId); err != nil {
			return err
		}
	}

	return nil
}

// NewPrepareSecurityGroupTask 新建准备安全组任务
func NewPrepareSecurityGroupTask(meta meta.TaskMeta, directer directer.Directer, step int) components.Task {
	t := &PrepareSecurityGroupTask{
		components.NewBaseTask(meta, directer, step),
	}

	return t
}
