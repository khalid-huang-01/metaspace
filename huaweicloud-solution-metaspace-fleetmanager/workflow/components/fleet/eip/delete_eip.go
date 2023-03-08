package eip

import (
	"fleetmanager/client"
	"fleetmanager/workflow/components"
	"fleetmanager/workflow/directer"
	"fleetmanager/workflow/meta"
)

type DeleteBandwidthsTask struct {
	components.BaseTask
}

// Execute 执行带宽删除任务
func (t *DeleteBandwidthsTask) Execute(ctx *directer.ExecuteContext) (output interface{}, err error) {
	defer func() { t.ExecNext(output, err) }()
	resourceProjectId := t.Directer.GetContext().Get(directer.WfKeyResourceProjectId).ToString("")
	regionId := t.Directer.GetContext().Get(directer.WfKeyBuildRegion).ToString("")
	agencyName := t.Directer.GetContext().Get(directer.WfKeyResourceAgencyName).ToString("")
	resourceDomainId := t.Directer.GetContext().Get(directer.WfKeyResourceDomainId).ToString("")
	bandwidthId := t.Directer.GetContext().Get(directer.WfKeyBandwidthId).ToString("")

	err = deleteBandwidth(regionId, resourceProjectId, agencyName, resourceDomainId, bandwidthId)
	if err != nil {
		return nil, err
	}

	return nil, nil
}

var deleteBandwidth = func(regionId string, projectId string, agencyName string, resDomainId string,
	eipId string) error {
	eipClient, err := client.GetAgencyEipClient(regionId, projectId, agencyName, resDomainId)
	if err != nil {
		return err
	}
	return client.DeleteEip(eipClient, eipId)
}

// NewDeleteBandwidthsTask 初始化带宽删除任务
func NewDeleteBandwidthsTask(meta meta.TaskMeta, directer directer.Directer, step int) components.Task {
	t := &DeleteBandwidthsTask{
		components.NewBaseTask(meta, directer, step),
	}

	return t
}
