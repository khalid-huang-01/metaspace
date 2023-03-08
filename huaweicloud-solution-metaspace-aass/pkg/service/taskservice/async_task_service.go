// Copyright (c) Huawei Technologies Co., Ltd. 2022-2022. All rights reserved.

// 异步任务
package taskservice

import (
	"github.com/pkg/errors"

	"scase.io/application-auto-scaling-service/pkg/cloudresource"
	"scase.io/application-auto-scaling-service/pkg/common"
	"scase.io/application-auto-scaling-service/pkg/db"
	"scase.io/application-auto-scaling-service/pkg/metricmonitor/interfaces"
	"scase.io/application-auto-scaling-service/pkg/taskmgmt"
	"scase.io/application-auto-scaling-service/pkg/taskmgmt/asynctask"
	"scase.io/application-auto-scaling-service/pkg/utils/logger"
)

// StartScaleOutGroupTask 启动扩容任务
// 若此时伸缩组非稳定，会返回错误 ErrScalingGroupNotStable
func StartScaleOutGroupTask(groupId string, targetNum int32) error {
	if targetNum <= 0 {
		return errors.Errorf("invalid param scaleOutNum[%d]", targetNum)
	}

	// 1. db记录伸缩组扩容任务开始
	// 若此时伸缩组非稳定，会返回错误 ErrScalingGroupNotStable
	err := db.TxRecordGroupScaleOutStart(groupId, targetNum)
	if err != nil {
		return err
	}

	// 2. 启动扩容异步任务
	taskmgmt.GetTaskMgmt().AddTask(asynctask.NewScaleOutTask(groupId, targetNum))
	return nil
}

// StartScaleInGroupTaskForRandomVms 启动缩容任务(随机选择缩容的vm)
// 若此时伸缩组非稳定，会返回错误 ErrScalingGroupNotStable
// Deprecated: 临时方法，需要根据vm的负载计算出最适合的vm进行缩容
func StartScaleInGroupTaskForRandomVms(groupId, projectId string, scaleInNum int32) error {
	// 校验缩容值
	if scaleInNum < 0 {
		return errors.Errorf("invalid param scaleInNum[%d]", scaleInNum)
	}
	if scaleInNum == 0 {
		return nil
	}

	// 获取asGroupId
	group, err := db.GetScalingGroupById(projectId, groupId)
	if err != nil {
		return err
	}
	if group.State != db.ScalingGroupStateStable && group.State != db.ScalingGroupStateError {
		return errors.Wrapf(common.ErrScalingGroupNotStable,
			"scaling group[%s] is in state[%s], cannot be scaled in", groupId, group.State)
	}
	vmGroup, err := db.GetVmScalingGroupById(group.ResourceId)
	if err != nil {
		return err
	}
	asGroupId := vmGroup.AsGroupId

	// 获取所有vm实例
	resCtrl, err := cloudresource.GetResourceController(projectId)
	if err != nil {
		return err
	}
	instanceIds, err := resCtrl.GetAsScalingInstanceIds(logger.R, asGroupId)
	if err != nil {
		return err
	}
	// 校验缩容值
	// 这里可能: 决策时，伸缩组正在缩容；缩容时（此时），缩容完毕vm数量变化，之前的决策失效
	if len(instanceIds) < int(scaleInNum) {
		return errors.Wrapf(common.ErrScalingDecisionExpired, "len of instanceIds[%d] < scaleInNum[%d], "+
			"scaling decision may have expired", len(instanceIds), scaleInNum)
	}

	// 选取需要缩容的vm
	deleteIds := instanceIds[:scaleInNum]
	return StartScaleInGroupTask(groupId, deleteIds)
}

// StartScaleInGroupTask 启动缩容任务
// 若此时伸缩组非稳定，会返回错误 ErrScalingGroupNotStable
func StartScaleInGroupTask(groupId string, instanceIds []string) error {
	if len(instanceIds) == 0 {
		return errors.Errorf("instanceIds must be provided to StartScaleInGroupTask")
	}

	// 1. db记录伸缩组缩容任务开始
	// 若此时伸缩组非稳定，会返回错误 ErrScalingGroupNotStable
	err := db.TxRecordGroupScaleInStart(groupId, instanceIds)
	if err != nil {
		return err
	}

	// 2. 启动缩容异步任务
	taskmgmt.GetTaskMgmt().AddTask(asynctask.NewScaleInTask(groupId, instanceIds))
	return nil
}

// StartDeleteScalingGroupTask 启动伸缩组删除任务
// 若该任务已存在（删除伸缩组api的可重入，会导致重复插入的情况），返回错误 ErrDelGroupTaskAlreadyExists
func StartDeleteScalingGroupTask(groupId string, monitor interfaces.MonitorInf) error {
	// 1. db记录伸缩组删除任务开始
	// 这里只是加入异步任务，而不立即将group状态修改为“deleting”是出于以下考虑：
	// 若伸缩组处于伸缩活动中，会等待此次扩缩执行完毕后再执行删除操作
	// 该等待较为耗时，所以等待的逻辑放在DelScalingGroupTask异步任务中（轮询db尝试将group的状态修改为“deleting”）
	if err := db.InsertDeleteScalingGroupTask(groupId); err != nil {
		return err
	}

	// 2. 启动伸缩组删除异步任务
	taskmgmt.GetTaskMgmt().AddTask(asynctask.NewDelScalingGroupTask(groupId, monitor))
	return nil
}
