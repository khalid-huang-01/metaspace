// Copyright (c) Huawei Technologies Co., Ltd. 2022-2022. All rights reserved.

// 伸缩组辅助数据表定义
package db

import (
	"context"

	"github.com/beego/beego/v2/client/orm"
	"github.com/pkg/errors"

	"scase.io/application-auto-scaling-service/pkg/common"
	"scase.io/application-auto-scaling-service/pkg/utils/logger"
)

// TxRecordGroupScaleOutStart 记录伸缩组扩容开始（事务）
func TxRecordGroupScaleOutStart(groupId string, targetNum int32) error {
	return ormer.DoTx(func(ctx context.Context, txOrm orm.TxOrmer) error {
		// 1. 更新伸缩组状态 stable -> scaling
		num, err := changeScalingGroupState(txOrm, groupId,
			ScalingGroupStateStable, ScalingGroupStateScaling)
		if err != nil {
			return err
		}
		// 伸缩组非稳定状态（包含伸缩组不存在的情况）
		if num != 1 {
			return errors.Wrapf(common.ErrScalingGroupNotStable,
				"scaling group[%s] is unstable or do not exist", groupId)
		}

		// 2. 记录扩容任务
		group := &ScalingGroup{Id: groupId}
		err = txOrm.QueryTable(tableNameScalingGroup).
			Filter(fieldNameId, groupId).
			One(group, fieldNameProjectId)
		if err != nil {
			return errors.Wrapf(err, "get project id for group[%s] err", groupId)
		}
		return txInsertAsyncTask(txOrm, newAsyncTaskForScaleOut(groupId, group.ProjectId, targetNum))
	})
}

// TxRecordGroupScaleOutComplete 记录伸缩组扩容完成（事务）
func TxRecordGroupScaleOutComplete(groupId string) error {
	return ormer.DoTx(func(ctx context.Context, txOrm orm.TxOrmer) error {
		// 1. 更新伸缩组状态 scaling -> stable
		_, err := changeScalingGroupState(txOrm, groupId,
			ScalingGroupStateScaling, ScalingGroupStateStable)
		if err != nil {
			return err
		}

		// 2. 删除扩容任务
		return txDeleteAsyncTask(txOrm, TaskTypeScaleOutScalingGroup, groupId)
	})
}

// TxRecordGroupScaleInStart 记录伸缩组缩容开始（事务）
func TxRecordGroupScaleInStart(groupId string, scaleInInstanceIds []string) error {
	return ormer.DoTx(func(ctx context.Context, txOrm orm.TxOrmer) error {
		// 1. 更新伸缩组状态 stable -> scaling
		num, err := changeScalingGroupState(txOrm, groupId,
			ScalingGroupStateStable, ScalingGroupStateScaling)
		if err != nil {
			return err
		}
		// 伸缩组非稳定状态（包含伸缩组不存在的情况）
		if num != 1 {
			return errors.Wrapf(common.ErrScalingGroupNotStable,
				"scaling group[%s] is unstable or do not exist", groupId)
		}

		// 2. 记录扩容任务
		group := &ScalingGroup{Id: groupId}
		err = txOrm.QueryTable(tableNameScalingGroup).
			Filter(fieldNameId, groupId).
			One(group, fieldNameProjectId)
		if err != nil {
			return errors.Wrapf(err, "get project id for group[%s] err", groupId)
		}
		return txInsertAsyncTask(txOrm, newAsyncTaskForScaleIn(groupId, group.ProjectId, scaleInInstanceIds))
	})
}

// TxRecordGroupScaleInComplete 记录伸缩组缩容完成（事务）
func TxRecordGroupScaleInComplete(groupId string) error {
	return ormer.DoTx(func(ctx context.Context, txOrm orm.TxOrmer) error {
		// 1. 更新伸缩组状态 scaling -> stable
		_, err := changeScalingGroupState(txOrm, groupId,
			ScalingGroupStateScaling, ScalingGroupStateStable)
		if err != nil {
			return err
		}

		// 2. 删除扩容任务
		return txDeleteAsyncTask(txOrm, TaskTypeScaleInScalingGroup, groupId)
	})
}

// ChangeScalingGroupState2Deleting 将伸缩组状态更新为"deleting"，即伸缩组删除开始
// 若当前伸缩组不可删除，返回错误 ErrScalingGroupCannotBeDeleted
func ChangeScalingGroupState2Deleting(log *logger.FMLogger, groupId string) error {
	return ormer.DoTx(func(ctx context.Context, txOrm orm.TxOrmer) error {
		// 1. 校验当前group的状态
		group := &ScalingGroup{
			Id:        groupId,
			TimeModel: TimeModel{IsDeleted: notDeletedFlag},
		}
		err := txOrm.ReadForUpdate(group)
		if err != nil {
			return errors.Wrapf(err, "read group[%s] from db err", groupId)
		}
		// 若处于“deleting”，无需处理
		if group.State == ScalingGroupStateDeleting {
			log.Info("Group[%s] state is deleting, do nothing", groupId)
			return nil
		}
		// 目前仅有“stable”和“error”状态时，group才可被删除
		if group.State != ScalingGroupStateStable && group.State != ScalingGroupStateError {
			log.Info("Group[%s] cannot be deleted in state[%s]", groupId, group.State)
			return common.ErrScalingGroupCannotBeDeleted
		}

		// 2. 更新状态为“deleting”
		group.State = ScalingGroupStateDeleting
		_, err = txOrm.Update(group, "State")
		if err != nil {
			return errors.Wrapf(err, "update group[%s] state to deleting err", groupId)
		}
		return nil
	})
}

// TxRecordGroupDeletingComplete 软删除伸缩组相关信息，并记录伸缩组删除任务结束
func TxRecordGroupDeletingComplete(log *logger.FMLogger, groupId string) error {
	return ormer.DoTx(func(ctx context.Context, txOrm orm.TxOrmer) error {
		// 1. 软删除ScalingGroup及其附属对象
		err := DeleteScalingGroup(txOrm, groupId)
		if err != nil {
			return err
		}

		// 2. 删除异步任务
		return txDeleteAsyncTask(txOrm, TaskTypeDeleteScalingGroup, groupId)
	})
}

// changeScalingGroupState 将伸缩组状态由 oldState 改为 newState
// 若group不存在，不会返回error
func changeScalingGroupState(txOrm orm.TxOrmer, groupId, oldState, newState string) (int64, error) {
	num, err := txOrm.QueryTable(tableNameScalingGroup).
		Filter(fieldNameId, groupId).
		Filter(fieldNameIsDeleted, notDeletedFlag).
		Filter(fieldNameState, oldState).
		Update(orm.Params{
			fieldNameState: newState,
		})
	if err != nil {
		return 0, errors.Wrapf(err, "update scaling group[%s] state to [%s] err", groupId, newState)
	}
	return num, nil
}
