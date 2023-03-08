// Copyright (c) Huawei Technologies Co., Ltd. 2022-2022. All rights reserved.

// 准备eip
package eip

import (
	"fleetmanager/client"
	"fleetmanager/workflow/components"
	"fleetmanager/workflow/directer"
	"fleetmanager/workflow/meta"
)

type PrepareEipTask struct {
	components.BaseTask
}

// Execute 执行eip准备任务
func (t *PrepareEipTask) Execute(ctx *directer.ExecuteContext) (output interface{}, err error) {
	defer func() { t.ExecNext(output, err) }()
	resourceProjectId := t.Directer.GetContext().Get(directer.WfKeyResourceProjectId).ToString("")
	regionId := t.Directer.GetContext().Get(directer.WfKeyBuildRegion).ToString("")
	bandwidthName := t.Directer.GetContext().Get(directer.WfKeyBandwidthName).ToString("")
	bandwidthSize := t.Directer.GetContext().Get(directer.WfKeyBuildBandwidth).ToInt(0)
	bandwidthType := t.Directer.GetContext().Get(directer.WfKeyBandwidthType).ToString("")
	agencyName := t.Directer.GetContext().Get(directer.WfKeyResourceAgencyName).ToString("")
	resourceDomainId := t.Directer.GetContext().Get(directer.WfKeyResourceDomainId).ToString("")
	var bandwidthId = t.Directer.GetContext().Get(directer.WfKeyBandwidthId).ToString("")

	if bandwidthId == "" {
		// 1. 创建eip
		bandwidthId, err = createEip(regionId, resourceProjectId, agencyName, resourceDomainId, bandwidthName,
			int32(bandwidthSize), bandwidthType)
		if err != nil {
			// 获取镜像信息错误，待重试
			return nil, err
		}

		// 2. 更新eipId记录
		if err = t.Directer.GetContext().Set(directer.WfKeyBandwidthId, bandwidthId); err != nil {
			return nil, err
		}
	}

	return nil, nil
}

func (t *PrepareEipTask) Rollback(*directer.ExecuteContext) (output interface{}, err error) {
	defer func() { t.RollbackPrev(output, err) }()

	regionId := t.Directer.GetContext().Get(directer.WfKeyBuildRegion).ToString("")
	eipId := t.Directer.GetContext().Get(directer.WfKeyBandwidthId).ToString("")
	agencyName := t.Directer.GetContext().Get(directer.WfKeyResourceAgencyName).ToString("")
	resourceDomainId := t.Directer.GetContext().Get(directer.WfKeyResourceDomainId).ToString("")
	resourceProjectId := t.Directer.GetContext().Get(directer.WfKeyResourceProjectId).ToString("")
	eipClient, err := client.GetAgencyEipClient(regionId, resourceProjectId, agencyName, resourceDomainId)

	err = client.DeleteEip(eipClient, eipId)
	if err != nil {
		return nil, err
	}
	return nil, nil
}

var createEip = func(regionId string, projectId string, agencyName string, resDomainId string,
	bandwidthName string, bandwidthSize int32, bandwidthType string) (string, error) {
	eipClient, err := client.GetAgencyEipClient(regionId, projectId, agencyName, resDomainId)
	if err != nil {
		return "", err
	}
	return client.CreateEip(eipClient, &bandwidthName, &bandwidthSize, &bandwidthType)
}

// NewPrepareEipTask 初始化带宽准备任务
func NewPrepareEipTask(meta meta.TaskMeta, directer directer.Directer, step int) components.Task {
	t := &PrepareEipTask{
		components.NewBaseTask(meta, directer, step),
	}

	return t
}
