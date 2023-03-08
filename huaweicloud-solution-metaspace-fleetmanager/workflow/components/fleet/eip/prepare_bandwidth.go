// Copyright (c) Huawei Technologies Co., Ltd. 2022-2022. All rights reserved.

// 准备eip
package eip

import (
	"fleetmanager/client"
	"fleetmanager/workflow/components"
	"fleetmanager/workflow/directer"
	"fleetmanager/workflow/meta"
)

type PrepareBandwidthsTask struct {
	components.BaseTask
}

// Execute 执行带宽准备任务
func (t *PrepareBandwidthsTask) Execute(ctx *directer.ExecuteContext) (output interface{}, err error) {
	defer func() { t.ExecNext(output, err) }()
	resourceProjectId := t.Directer.GetContext().Get(directer.WfKeyResourceProjectId).ToString("")
	regionId := t.Directer.GetContext().Get(directer.WfKeyRegion).ToString("")
	bandwidthName := t.Directer.GetContext().Get(directer.WfKeyBandwidthName).ToString("")
	bandwidthSize := t.Directer.GetContext().Get(directer.WfKeyBandwidth).ToInt(0)
	bandwidthType := t.Directer.GetContext().Get(directer.WfKeyBandwidthType).ToString("")
	agencyName := t.Directer.GetContext().Get(directer.WfKeyResourceAgencyName).ToString("")
	resourceDomainId := t.Directer.GetContext().Get(directer.WfKeyResourceDomainId).ToString("")
	var bandwidthId = t.Directer.GetContext().Get(directer.WfKeyBandwidthId).ToString("")

	if bandwidthId == "" {
		// 1. 创建bandwidth
		bandwidthId, err = createBandwidth(regionId, resourceProjectId, agencyName, resourceDomainId, bandwidthName,
			int32(bandwidthSize), bandwidthType)
		if err != nil {
			// 获取镜像信息错误，待重试
			return nil, err
		}

		// 2. 更新vpcId记录
		if err = t.Directer.GetContext().Set(directer.WfKeyBandwidthId, bandwidthId); err != nil {
			// 任务设置异常，资源回滚再重试 TODO(wangjun)
			return nil, err
		}
	}

	return nil, nil
}

var createBandwidth = func(regionId string, projectId string, agencyName string, resDomainId string,
	bandwidthName string, bandwidthSize int32, bandwidthType string) (string, error) {
	eipClient, err := client.GetAgencyEipClient(regionId, projectId, agencyName, resDomainId)
	if err != nil {
		return "", err
	}
	return client.CreateBandwidth(eipClient, &bandwidthName, &bandwidthSize, &bandwidthType)
}

// NewPrepareBandwidthsTask 初始化带宽准备任务
func NewPrepareBandwidthsTask(meta meta.TaskMeta, directer directer.Directer, step int) components.Task {
	t := &PrepareBandwidthsTask{
		components.NewBaseTask(meta, directer, step),
	}

	return t
}
