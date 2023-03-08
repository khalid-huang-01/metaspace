// Copyright (c) Huawei Technologies Co., Ltd. 2022-2022. All rights reserved.

// 异步执行应用包创建
package build

import (
	"fleetmanager/db/dao"
	"fleetmanager/workflow/components"
	"fleetmanager/workflow/directer"
	"fleetmanager/workflow/meta"
)

type SyncBuildImageTask struct {
	components.BaseTask
}

// Execute 执行镜像信息同步
func (t *SyncBuildImageTask) Execute(ctx *directer.ExecuteContext) (output interface{}, err error) {
	defer func() { t.ExecNext(output, err) }()

	if buildImageId := t.Directer.GetContext().Get(directer.WfKeyBuildImageId).ToString(""); buildImageId != "" {
		// 已经同步过, 则直接返回
		return nil, nil
	}

	projectId := t.Directer.GetContext().Get(directer.WfKeyProjectId).ToString("")
	regionId := t.Directer.GetContext().Get(directer.WfKeyRegion).ToString("")
	buildId := t.Directer.GetContext().Get(directer.WfKeyBuildId).ToString("")

	// 1. 查询build的镜像信息
	imageId, err := dao.GetBuildImage(buildId, regionId, projectId)
	if err != nil {
		// 获取镜像信息错误，待重试
		return nil, err
	}

	// 2. 如果build镜像已经存在任务，则等待直到任务结束，返回 1
	// TODO(wangjun):

	// 3. 如果build镜像不存在任务，则启动镜像同步任务，返回 1
	// TODO(wangjun):

	// 4. Sync结束
	if err = t.Directer.GetContext().Set(directer.WfKeyBuildImageId, imageId); err != nil {
		// 任务设置异常，待重试
		return nil, err
	}

	return nil, nil
}

// NewSyncBuildImageTask 新建镜像同步任务
func NewSyncBuildImageTask(meta meta.TaskMeta, directer directer.Directer, step int) components.Task {
	t := &SyncBuildImageTask{
		components.NewBaseTask(meta, directer, step),
	}

	return t
}
