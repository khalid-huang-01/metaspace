// Copyright (c) Huawei Technologies Co., Ltd. 2022-2022. All rights reserved.

// 异步任务
package db

import (
	"context"
	"time"

	"github.com/beego/beego/v2/client/orm"
	"github.com/pkg/errors"

	"scase.io/application-auto-scaling-service/pkg/common"
	"scase.io/application-auto-scaling-service/pkg/utils"
	"scase.io/application-auto-scaling-service/pkg/utils/logger"
)

const (
	tableNameAsyncTask = "async_task"

	TaskTypeScaleOutScalingGroup = "scale_out"
	TaskTypeScaleInScalingGroup  = "scale_in"
	TaskTypeDeleteVm             = "delete_vm"
	TaskTypeDeleteScalingGroup   = "delete_scaling_group"
)

type AsyncTask struct {
	// 自增主键，若不设置，执行db事务会报错
	Id int
	// 任务类型
	TaskType string `orm:"column(task_type);size(64)"`
	// 任务操作的资源对象的Id
	// 比如扩缩容任务的TaskKey为待扩缩的groupID；vm删除任务的TaskKey为待删除的vmId
	TaskKey string `orm:"column(task_key);size(128)"`
	// 任务配置
	TaskConf string `orm:"column(task_conf);type(text)"`
	// 任务执行节点
	WorkNodeId string `orm:"column(work_node_id);size(128)"`
	TimeModel
}

// TableIndex 设置索引
func (t *AsyncTask) TableIndex() [][]string {
	return [][]string{
		{"TaskType", "TaskKey", "IsDeleted"},
	}
}

// newAsyncTaskForScaleOut 扩容任务db对象构造方法
func newAsyncTaskForScaleOut(groupId, projectId string, targetNum int32) *AsyncTask {
	task := &AsyncTask{
		TaskType: TaskTypeScaleOutScalingGroup,
		TaskKey:  groupId,
		TaskConf: utils.ToJson(&ScaleOutTaskConf{
			groupId,
			targetNum,
		}),
		WorkNodeId: common.LocalWorkNodeId,
	}
	task.IsDeleted = notDeletedFlag
	return task
}

// newAsyncTaskForScaleIn 缩容任务db对象构造方法
func newAsyncTaskForScaleIn(groupId, projectId string, scaleInInstanceIds []string) *AsyncTask {
	task := &AsyncTask{
		TaskType: TaskTypeScaleInScalingGroup,
		TaskKey:  groupId,
		TaskConf: utils.ToJson(&ScaleInTaskConf{
			groupId,
			scaleInInstanceIds,
		}),
		WorkNodeId: common.LocalWorkNodeId,
	}
	task.IsDeleted = notDeletedFlag
	return task
}

// newAsyncTaskForDeleteVm vm删除任务db对象构造方法
func newAsyncTaskForDeleteVm(vmId string, asGroupId string, projectId string) *AsyncTask {
	task := &AsyncTask{
		TaskType: TaskTypeDeleteVm,
		TaskKey:  vmId,
		TaskConf: utils.ToJson(&DeleteVmTaskConf{
			vmId,
			asGroupId,
			projectId,
		}),
		WorkNodeId: common.LocalWorkNodeId,
	}
	task.IsDeleted = notDeletedFlag
	return task
}

// newAsyncTaskForDeleteVm 伸缩组删除任务db对象构造方法
func newAsyncTaskForDeleteScalingGroup(groupId string) *AsyncTask {
	task := &AsyncTask{
		TaskType: TaskTypeDeleteScalingGroup,
		TaskKey:  groupId,
		TaskConf: utils.ToJson(&DeleteScalingGroupTaskConf{
			groupId,
		}),
		WorkNodeId: common.LocalWorkNodeId,
	}
	task.IsDeleted = notDeletedFlag
	return task
}

// InsertDeleteScalingGroupTask ...
// 若该任务已存在（删除伸缩组api的可重入，会导致重复插入的情况），返回错误 ErrDeleteTaskAlreadyExists
func InsertDeleteScalingGroupTask(groupId string) error {
	return ormer.DoTx(func(ctx context.Context, txOrm orm.TxOrmer) error {
		// 1. 判断若task已存在，无需insert，返回错误 ErrAsyncTaskAlreadyExists
		queryTask := &AsyncTask{
			TaskType:  TaskTypeDeleteScalingGroup,
			TaskKey:   groupId,
			TimeModel: TimeModel{IsDeleted: notDeletedFlag},
		}
		err := txOrm.ReadForUpdate(queryTask, "TaskType", "TaskKey", "IsDeleted")
		if err != nil {
			// 忽略task不存在的错误
			if !errors.Is(err, orm.ErrNoRows) {
				return errors.Wrap(err, "db read task err")
			}
		}
		if queryTask.TaskConf != "" {
			return errors.Wrapf(common.ErrDeleteTaskAlreadyExists,
				"delete group task for group[%s] already exists", groupId)
		}

		// 2. 执行task插入操作
		task := newAsyncTaskForDeleteScalingGroup(groupId)
		task.IsDeleted = notDeletedFlag
		_, err = txOrm.Insert(task)
		if err != nil {
			return errors.Wrapf(err, "db add async task[%s:%s] err", task.TaskType, task.TaskKey)
		}
		return nil
	})
}

// InsertDeleteVmAsyncTask ...
func InsertDeleteVmAsyncTask(vmId string, asGroupId string, projectId string) error {
	// 若该任务已存在，不做操作，返回错误 ErrAsyncTaskAlreadyExists
	exist := ormer.QueryTable(tableNameAsyncTask).
		Filter(fieldNameTaskType, TaskTypeDeleteVm).
		Filter(fieldNameTaskKey, vmId).
		Filter(fieldNameIsDeleted, notDeletedFlag).Exist()
	if exist {
		return errors.Wrapf(common.ErrDeleteTaskAlreadyExists,
			"delete vm task for vm[%s] already exists", vmId)
	}

	task := newAsyncTaskForDeleteVm(vmId, asGroupId, projectId)
	task.IsDeleted = notDeletedFlag
	_, err := ormer.Insert(task)
	if err != nil {
		return errors.Wrapf(err, "db add async task[%s:%s] err", task.TaskType, task.TaskKey)
	}
	return nil
}

// DeleteAsyncTask ...
func DeleteAsyncTask(taskType, taskKey string) error {
	_, err := ormer.QueryTable(tableNameAsyncTask).
		Filter(fieldNameTaskType, taskType).
		Filter(fieldNameTaskKey, taskKey).
		Filter(fieldNameIsDeleted, notDeletedFlag).
		Update(orm.Params{
			fieldNameIsDeleted: deletedFlag,
			fieldNameDeleteAt:  time.Now().UTC()})
	if err != nil {
		if errors.Is(err, orm.ErrNoRows) {
			return nil
		}
		return errors.Wrapf(err, "db delete async task[%s:%s] err", taskType, taskKey)
	}
	return nil
}

// GetAllAsyncTasks ...
func GetAllAsyncTasks() ([]*AsyncTask, error) {
	var tasks []*AsyncTask
	_, err := ormer.QueryTable(tableNameAsyncTask).
		Filter(fieldNameIsDeleted, notDeletedFlag).
		All(&tasks)
	if err != nil {
		return nil, errors.Wrapf(err, "get all async task from db err")
	}
	return tasks, nil
}

// GetAsyncTasksByWorkNodeId ...
func GetAsyncTasksByWorkNodeId(wnId string) ([]*AsyncTask, error) {
	var tasks []*AsyncTask
	_, err := ormer.QueryTable(tableNameAsyncTask).
		Filter(fieldNameIsDeleted, notDeletedFlag).
		Filter(fieldNameWorkNodeId, wnId).
		All(&tasks)
	if err != nil {
		return nil, errors.Wrapf(err, "get all async task of work node[%s] from db err", wnId)
	}
	return tasks, nil
}

// TakeOverAsyncTasks ...
func TakeOverAsyncTasks(taskIds []int, takeOverId string) error {
	if len(taskIds) == 0 {
		return nil
	}
	_, err := ormer.QueryTable(tableNameAsyncTask).
		Filter(fieldNameIdIn, taskIds).
		Update(orm.Params{
			fieldNameWorkNodeId: takeOverId,
		})
	return err
}

// ScaleOutTaskConf 扩容任务配置
type ScaleOutTaskConf struct {
	ScalingGroupId       string `json:"scaling_group_id"`
	TargetInstanceNumber int32  `json:"target_instance_number"`
}

// ScaleInTaskConf 缩容任务配置
type ScaleInTaskConf struct {
	ScalingGroupId     string   `json:"scaling_group_id"`
	ScaleInInstanceIds []string `json:"scale_in_instance_ids"`
	// CurrentInstanceNumber int      `json:"current_instance_number"`
}

// DeleteVmTaskConf vm删除任务配置
type DeleteVmTaskConf struct {
	// 待删除的vmId
	DeleteVmId string `json:"delete_vm_id"`
	// 该vm之前隶属的as伸缩组的Id
	AsGroupId    string `json:"as_group_id"`
	ResProjectId string `json:"res_project_id"`
}

// DeleteScalingGroupTaskConf 伸缩组删除任务配置
type DeleteScalingGroupTaskConf struct {
	ScalingGroupId string `json:"scaling_group_id"`
}

// txInsertAsyncTask ...
func txInsertAsyncTask(txOrm orm.TxOrmer, task *AsyncTask) error {
	if task == nil {
		return errors.New("task provided to txInsertAsyncTask must not be nil")
	}

	task.IsDeleted = notDeletedFlag
	_, err := txOrm.Insert(task)
	if err != nil {
		return errors.Wrapf(err, "db add async task[%s:%s] err", task.TaskType, task.TaskKey)
	}
	return nil
}

// txDeleteAsyncTask ...
func txDeleteAsyncTask(txOrm orm.TxOrmer, taskType, taskKey string) error {
	num, err := txOrm.QueryTable(tableNameAsyncTask).
		Filter(fieldNameTaskType, taskType).
		Filter(fieldNameTaskKey, taskKey).
		Filter(fieldNameIsDeleted, notDeletedFlag).
		Update(orm.Params{
			fieldNameIsDeleted: deletedFlag,
			fieldNameDeleteAt:  time.Now().UTC()})
	if err != nil {
		if errors.Is(err, orm.ErrNoRows) {
			return nil
		}
		return errors.Wrapf(err, "db delete async task[%s:%s] err", taskType, taskKey)
	}
	if num != 1 {
		logger.R.Warn("There may be a program logic error, updated num[%d]", num)
	}
	return nil
}
