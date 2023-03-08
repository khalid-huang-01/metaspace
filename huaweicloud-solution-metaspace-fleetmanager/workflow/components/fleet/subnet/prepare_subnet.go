// Copyright (c) Huawei Technologies Co., Ltd. 2022-2022. All rights reserved.

// 准备子网
package subnet

import (
	"fleetmanager/api/cidrmanager"
	"fleetmanager/api/service/constants"
	"fleetmanager/client"
	"fleetmanager/setting"
	"fleetmanager/workflow/components"
	"fleetmanager/workflow/directer"
	"fleetmanager/workflow/meta"
	"fmt"
	"strings"
)

var (
	clientCreateSubnet       = client.CreateSubnet
	clientGetAgencyVpcClient = client.GetAgencyVpcClient
)

type PrepareSubnetTask struct {
	components.BaseTask
}

// Execute 执行子网准备任务
func (t *PrepareSubnetTask) Execute(ctx *directer.ExecuteContext) (output interface{}, err error) {
	defer func() { t.ExecNext(output, err) }()
	resourceProjectId := t.Directer.GetContext().Get(directer.WfKeyResourceProjectId).ToString("")
	regionId := t.Directer.GetContext().Get(directer.WfKeyRegion).ToString("")
	subnetName := t.Directer.GetContext().Get(directer.WfKeySubnetName).ToString("")
	subnetCidr := t.Directer.GetContext().Get(directer.WfKeySubnetCidr).ToString("")
	gatewayIp := t.Directer.GetContext().Get(directer.WfKeySubnetGatewayIp).ToString("")
	vpcId := t.Directer.GetContext().Get(directer.WfKeyVpcId).ToString("")
	subnetId := t.Directer.GetContext().Get(directer.WfKeySubnetId).ToString("")
	agencyName := t.Directer.GetContext().Get(directer.WfKeyResourceAgencyName).ToString("")
	resourceDomainId := t.Directer.GetContext().Get(directer.WfKeyResourceDomainId).ToString("")

	if subnetId == "" {
		// 查询subnetId
		subnetId, _, err = getSubnet(regionId, resourceProjectId, agencyName, resourceDomainId, subnetName, vpcId)
		if err != nil {
			return nil, err
		}
	}

	if subnetId == "" {
		// 1. 创建subnet
		subnetId, err = createSubnet(regionId, resourceProjectId, agencyName, resourceDomainId, subnetName, subnetCidr,
			gatewayIp, vpcId, t.Directer.GetContext().Get(directer.WfDnsConfig).ToString(setting.DefaultDnsConfig))
		if err != nil {
			// 创建subnet错误，待重试
			return nil, err
		}
	}

	// 2. 更新subnetId记录
	if err = t.Directer.GetContext().Set(directer.WfKeySubnetId, subnetId); err != nil {
		// 任务设置异常，资源回滚再重试 TODO(wangjun)
		return nil, err
	}

	// 3. 等待subnet创建完成
	if err := waitSubnetReady(regionId, resourceProjectId, agencyName, resourceDomainId, subnetId); err != nil {
		return nil, err
	}

	return nil, nil
}

var createSubnet = func(regionId string, projectId string, agencyName string, resDomainId string,
	subnetName string, subnetCidr string, gatewayIp string, vpcId string, dnsConfig string) (string, error) {
	vpcClient, err := clientGetAgencyVpcClient(regionId, projectId, agencyName, resDomainId)
	if err != nil {
		return "", err
	}
	primaryDns, secondaryDns, dnsList := getSubnetDnsConfig(dnsConfig)
	return clientCreateSubnet(vpcClient, &vpcId, &gatewayIp, &subnetCidr, &subnetName, primaryDns, secondaryDns,
		dnsList)
}

var checkSubnet = func(regionId string, projectId string, agencyName string, resDomainId string,
	vpcId string, subnetCidr string, gatewayIp string, retryTimes int, fleetId string) (string, string, error) {
	vpcClient, err := client.GetAgencyVpcClient(regionId, projectId, agencyName, resDomainId)
	if err != nil {
		return "", "", err
	}
	var NewSubnetCidr = subnetCidr
	var NewGatewayIp = gatewayIp
	for a := 0; a < retryTimes; a++ {
		isOverlap, err := client.CheckSubnet(vpcClient, &NewSubnetCidr, vpcId)
		if err != nil {
			return "", "", err
		}
		if !isOverlap {
			return NewSubnetCidr, NewGatewayIp, err
		} else {
			vpcCidr, err := cidrmanager.CreateVpcCidr(constants.DefaultNameSpace, fleetId)
			if err != nil {
				return "", "", err
			}
			NewSubnetCidr, NewGatewayIp, err = cidrmanager.CreateSubnetCidr(vpcCidr)
			if err != nil {
				return "", "", err
			}
		}
	}
	return "", "", fmt.Errorf("can't generate available subnet after %d times for vpc %s", retryTimes, vpcId)
}

var getSubnetDnsConfig = func(dnsConfig string) (primaryDns *string, secondaryDns *string, dnsList *[]string) {
	primaryDns, secondaryDns, dnsList = nil, nil, nil
	configSlice := strings.Split(dnsConfig, ",")
	switch len(configSlice) {
	case 1:
		if len(configSlice[0]) != 0 {
			primaryDns = &configSlice[0]
			dnsList = &configSlice
		}
		return
	default:
		primaryDns = &configSlice[0]
		secondaryDns = &configSlice[1]
		dnsList = &[]string{
			*primaryDns,
			*secondaryDns,
		}
	}

	return
}

var waitSubnetReady = func(regionId string, projectId string, agencyName string, resDomainId string,
	subnetId string) error {
	vpcClient, err := client.GetAgencyVpcClient(regionId, projectId, agencyName, resDomainId)
	if err != nil {
		return err
	}
	return client.WaitSubnetReady(vpcClient, &subnetId)
}

// NewPrepareSubnetTask 新建准备子网任务
func NewPrepareSubnetTask(meta meta.TaskMeta, directer directer.Directer, step int) components.Task {
	t := &PrepareSubnetTask{
		components.NewBaseTask(meta, directer, step),
	}

	return t
}
