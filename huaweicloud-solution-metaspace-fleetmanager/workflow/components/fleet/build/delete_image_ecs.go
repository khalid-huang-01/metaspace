package build

import (
	"fleetmanager/client"
	"fleetmanager/workflow/components"
	"fleetmanager/workflow/directer"
	"fleetmanager/workflow/meta"
)

type DeleteImageEcsTask struct {
	components.BaseTask
}

// Execute 执行删除ECS资源操作
func (t *DeleteImageEcsTask) Execute(ctx *directer.ExecuteContext) (output interface{}, err error) {
	defer func() { t.ExecNext(output, err) }()

	ecsId := t.Directer.GetContext().Get(directer.WfKeyBuildECSId).ToString("")
	resourceProjectId := t.Directer.GetContext().Get(directer.WfKeyResourceProjectId).ToString("")
	regionId := t.Directer.GetContext().Get(directer.WfKeyBuildRegion).ToString("")
	agencyName := t.Directer.GetContext().Get(directer.WfKeyResourceAgencyName).ToString("")
	resourceDomainId := t.Directer.GetContext().Get(directer.WfKeyResourceDomainId).ToString("")

	if err := deleteEcs(regionId, resourceProjectId, agencyName, resourceDomainId, ecsId); err != nil {
		return nil, err
	}
	return nil, nil
}

// NewDeleteImageEcsTask 新建删除资源任务
func NewDeleteImageEcsTask(meta meta.TaskMeta, directer directer.Directer, step int) components.Task {
	t := &DeleteImageEcsTask{
		components.NewBaseTask(meta, directer, step),
	}

	return t
}

func deleteEcs(regionId string, resourceProjectId string, agencyName string, resourceDomainId string,
	ecsId string) error {
	ecsClient, err := client.GetAgencyEcsClient(regionId, resourceProjectId, agencyName, resourceDomainId)
	if err != nil {
		return err
	}
	jobId, err := client.DeleteEcs(ecsClient, ecsId)

	err = client.WaitEcsDeleted(ecsClient, jobId)
	if err != nil {
		return err
	}

	return nil
}
