// Copyright (c) Huawei Technologies Co., Ltd. 2022-2022. All rights reserved.

// 缩容任务
package asynctask

import (
	"github.com/pkg/errors"

	"scase.io/application-auto-scaling-service/pkg/cloudresource"
	"scase.io/application-auto-scaling-service/pkg/cloudresource/cloudhelper"
	"scase.io/application-auto-scaling-service/pkg/common"
	"scase.io/application-auto-scaling-service/pkg/db"
	"scase.io/application-auto-scaling-service/pkg/taskmgmt"
	"scase.io/application-auto-scaling-service/pkg/utils/logger"
)

const (
	TaskTypeScaleIn = db.TaskTypeScaleInScalingGroup
)

// ScaleInTask 缩容任务
type ScaleInTask struct {
	GroupId            string
	ScaleInInstanceIds []string
	BaseTask
}

// NewScaleInTask ...
func NewScaleInTask(groupId string, scaleInInstanceIds []string) *ScaleInTask {
	return &ScaleInTask{
		GroupId:            groupId,
		ScaleInInstanceIds: scaleInInstanceIds,
	}
}

// GetKey get id of the resource corresponding to the task
func (t *ScaleInTask) GetKey() string {
	return t.GroupId
}

// GetType get task type
func (t *ScaleInTask) GetType() string {
	return TaskTypeScaleIn
}

// Run run task
func (t *ScaleInTask) Run(log *logger.FMLogger) error {
	// 1. 查询该伸缩组对应的底层资源信息，即as伸缩组信息
	group, err := db.GetScalingGroupById("", t.GroupId)
	if err != nil {
		return err
	}
	vmGroup, err := db.GetVmScalingGroupById(group.ResourceId)
	if err != nil {
		return err
	}
	asGroupId := vmGroup.AsGroupId
	projectId := group.ProjectId

	// 2. 缩容as伸缩组，从as伸缩组中移除实例，并将实例关机
	err = cloudhelper.ScaleInAsScalingGroupByInstances(log, asGroupId, projectId, t.ScaleInInstanceIds)
	if err != nil {
		return err
	}
	// 等待as伸缩组稳定
	resCtrl, err := cloudresource.GetResourceController(projectId)
	if err != nil {
		return err
	}
	err = resCtrl.WaitAsGroupStable(log, asGroupId)
	if err != nil {
		return err
	}

	// 3. 关闭所有虚机
	if err = resCtrl.BatchStopServers(log, t.ScaleInInstanceIds); err != nil {
		return err
	}

	// 4.DB记录vm删除任务、启动vm删除异步任务
	for _, instanceId := range t.ScaleInInstanceIds {
		err = db.InsertDeleteVmAsyncTask(instanceId, asGroupId, projectId)
		if err != nil {
			// 若该任务正在运行，不处理
			if errors.Is(err, common.ErrDeleteTaskAlreadyExists) {
				continue
			}
			return err
		}
		// 临时方案，目前还是数据库记录了待删除vm，后面优化后删除
		if err = db.AddDeletingVm(&db.DeletingVm{
			Id:        instanceId,
			AsGroupId: asGroupId,
			ProjectId: projectId,
		}); err != nil {
			return err
		}
		taskmgmt.GetTaskMgmt().AddTask(NewDelVmTask(instanceId, projectId))
	}

	// 5. db记录伸缩组缩容结束
	return db.TxRecordGroupScaleInComplete(t.GroupId)
}
