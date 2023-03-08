// Copyright (c) Huawei Technologies Co., Ltd. 2022-2022. All rights reserved.

// 删除弹性伸缩组任务
package asynctask

import (
	"time"

	"github.com/beego/beego/v2/client/orm"
	"github.com/pkg/errors"

	"scase.io/application-auto-scaling-service/pkg/cloudresource"
	"scase.io/application-auto-scaling-service/pkg/common"
	"scase.io/application-auto-scaling-service/pkg/db"
	"scase.io/application-auto-scaling-service/pkg/metricmonitor/interfaces"
	"scase.io/application-auto-scaling-service/pkg/utils/logger"
)

const (
	TaskTypeDelScalingGroup = db.TaskTypeDeleteScalingGroup

	waitVmDeletedTimes            = 60
	waitASGroupDeletedTimes       = 30
	waitVmDeletedInternal         = time.Second * 10
	waitGroupCanBeDeletedInternal = time.Second * 10
)

// DelScalingGroupTask 删除伸缩组异步任务
type DelScalingGroupTask struct {
	// 监控模块，删除伸缩组对应监控任务时使用
	monitor interfaces.MonitorInf
	groupId string
	BaseTask
}

// NewDelScalingGroupTask get new task to delete scaling group
func NewDelScalingGroupTask(groupId string, monitor interfaces.MonitorInf) *DelScalingGroupTask {
	return &DelScalingGroupTask{
		monitor: monitor,
		groupId: groupId,
	}
}

// GetKey get id of the resource corresponding to the task
func (t *DelScalingGroupTask) GetKey() string {
	return t.groupId
}

// GetType get task type
func (t *DelScalingGroupTask) GetType() string {
	return TaskTypeDelScalingGroup
}

// Run run task
func (t *DelScalingGroupTask) Run(log *logger.FMLogger) error {
	// 1. 从数据库读取group，若没有，说明资源已删除完毕
	group, err := db.GetNotDeletedGroupById("", t.groupId)
	if err != nil {
		if errors.Is(err, orm.ErrNoRows) {
			log.Info("Scaling group[%s] has been deleted, do nothing", t.groupId)
			return nil
		}
		return err
	}

	// 2. 删除伸缩策略 和 对应的监控任务
	if len(group.ScalingPolicies) != 0 {
		if err = t.delPolicyAndMonitorTask(log, group.ScalingPolicies); err != nil {
			return err
		}
	}

	// 3. 修改伸缩组的状态为“deleting”（相当于对该伸缩组对应的底层as伸缩组的操作进行加锁）
	// 目前来说，若update不成功，说明此时伸缩组可能处于扩缩活动中，重复update操作
	for {
		err = db.ChangeScalingGroupState2Deleting(log, t.groupId)
		if err != nil {
			// 当前伸缩组状态不支持删除，重试
			if errors.Is(err, common.ErrScalingGroupCannotBeDeleted) {
				time.Sleep(waitGroupCanBeDeletedInternal)
				continue
			}
			return err
		}
		// 修改伸缩组的状态“deleting”成功，继续执行之后的操作
		break
	}

	vmGroup, err := db.GetVmScalingGroupById(group.ResourceId)
	if err != nil {
		return err
	}
	// 4. 删除as伸缩组相关云资源，涉及AS、ECS
	if err = t.delAsGroupCloudRes(log, vmGroup.Id, group.ProjectId); err != nil {
		return err
	}
	// 5. 等待之前隶属于该as伸缩组的vm实例删除完毕
	if err = t.waitVmDeleted(log, vmGroup.AsGroupId); err != nil {
		return err
	}

	// 6. 删除数据库相关信息
	return db.TxRecordGroupDeletingComplete(log, t.groupId)
}

// 删除伸缩策略 和 对应的监控任务
func (t *DelScalingGroupTask) delPolicyAndMonitorTask(log *logger.FMLogger, policies []*db.ScalingPolicy) error {
	if len(policies) == 0 {
		return nil
	}

	ids := make([]string, 0, len(policies))
	for _, policy := range policies {
		if policy.PolicyType == common.PolicyTypeTargetBased {
			taskId := t.monitor.TaskIdForPolicy(policy.Id)
			if err := t.monitor.DeleteTask(taskId); err != nil {
				return err
			}
		}
		if err := db.DeleteScalingPolicy(policy.ProjectId, policy.Id); err != nil {
			return err
		}
		ids = append(ids, policy.Id)
	}
	log.Info("The scaling policies[%v] of group[%s] has been deleted", ids, t.groupId)
	return nil
}

// waitVmDeleted 等待之前隶属于该as伸缩组的vm实例删除完毕
// Deprecated: 临时方案，后续废弃
func (t *DelScalingGroupTask) waitVmDeleted(log *logger.FMLogger, asGroupId string) error {
	var (
		vmIds []string
		err   error
	)
	times := 0
	for {
		// 获取as伸缩组正在删除的实例vmId
		vmIds, err = db.GetAsGroupDeletingVmIds(asGroupId)
		if err != nil {
			return err
		}
		if len(vmIds) == 0 {
			return nil
		}
		log.Info("As group[%s] has vms %v waiting to be deleted", asGroupId, vmIds)
		time.Sleep(waitVmDeletedInternal)
		times++
		if times == waitVmDeletedTimes {
			return errors.Errorf("vm for as group[%s] has not been deleted for 10 min", asGroupId)
		}
	}
}

// delAsGroupCloudRes 删除 as伸缩组 相关云资源，涉及AS、ECS
func (t *DelScalingGroupTask) delAsGroupCloudRes(log *logger.FMLogger, vmGroupId, projectId string) error {
	vmGroup, err := db.GetVmScalingGroupById(vmGroupId)
	if err != nil {
		if errors.Is(err, orm.ErrNoRows) {
			log.Info("AsGroup[%s] cloud res has been deleted and no further operation is required")
			return nil
		}
		return err
	}

	resCtrl, err := cloudresource.GetResourceController(projectId)
	if err != nil {
		return err
	}
	// 删除as伸缩组
	if vmGroup.AsGroupId != "" {
		if err = resCtrl.DelAsGroup(log, vmGroup.AsGroupId); err != nil {
			return err
		}
	}
	// 等待as伸缩组删除
	time.Sleep(waitASGroupDeletedTimes * time.Second)
	// 删除as伸缩组配置
	if err = resCtrl.DeleteAsScalingConfig(log, vmGroup.ScalingConfigId); err != nil {
		return err
	}
	return nil
}
