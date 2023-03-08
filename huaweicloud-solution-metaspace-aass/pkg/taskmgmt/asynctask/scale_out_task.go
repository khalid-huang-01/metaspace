// Copyright (c) Huawei Technologies Co., Ltd. 2022-2022. All rights reserved.

// 扩容任务
package asynctask

import (
	"scase.io/application-auto-scaling-service/pkg/cloudresource"
	"scase.io/application-auto-scaling-service/pkg/cloudresource/cloudhelper"
	"scase.io/application-auto-scaling-service/pkg/db"
	"scase.io/application-auto-scaling-service/pkg/utils/logger"
)

const (
	TaskTypeScaleOut = db.TaskTypeScaleOutScalingGroup
)

// ScaleOutTask 扩容任务
type ScaleOutTask struct {
	GroupId           string
	targetInstanceNum int32
	BaseTask
}

// NewScaleOutTask ...
func NewScaleOutTask(groupId string, targetInstanceNum int32) *ScaleOutTask {
	return &ScaleOutTask{
		GroupId:           groupId,
		targetInstanceNum: targetInstanceNum,
	}
}

// GetKey get id of the resource corresponding to the task
func (t *ScaleOutTask) GetKey() string {
	return t.GroupId
}

// GetType get task type
func (t *ScaleOutTask) GetType() string {
	return TaskTypeScaleOut
}

// Run run task
func (t *ScaleOutTask) Run(log *logger.FMLogger) error {
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

	// 2. 扩容as伸缩组，修改as伸缩组的 desireNum
	err = cloudhelper.ScaleOutAsScalingGroupToTarget(log, asGroupId, projectId, t.targetInstanceNum)
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

	// 3. db记录伸缩组缩容结束
	return db.TxRecordGroupScaleOutComplete(t.GroupId)
}
