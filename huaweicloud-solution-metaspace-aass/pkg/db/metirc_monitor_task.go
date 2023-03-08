// Copyright (c) Huawei Technologies Co., Ltd. 2022-2022. All rights reserved.

// 监控任务数据表定义
package db

import (
	"time"

	"github.com/beego/beego/v2/client/orm"
	"github.com/pkg/errors"

	"scase.io/application-auto-scaling-service/pkg/common"
)

const (
	tableNameMetricMonitorTask = "metric_monitor_task"
)

type MetricMonitorTask struct {
	Id              string `orm:"column(id);size(128);pk"`
	MetricName      string `orm:"column(metric_name);size(128)"`             // 指标名称
	TargetValue     int32  `orm:"column(target_value);type(int);default(0)"` // 目标阈值
	ScalingGroupID  string `orm:"column(scaling_group_id);size(128)"`
	ScalingPolicyID string `orm:"column(scaling_policy_id);size(128)"`
	WorkNodeId      string `orm:"column(work_node_id);size(128)"` // 任务执行节点
	TimeModel
}

// AddMetricMonitorTask ...
func AddMetricMonitorTask(task *MetricMonitorTask) error {
	if task == nil {
		return errors.New("func AddMetricMonitorTask has invalid args")
	}

	task.IsDeleted = notDeletedFlag
	task.WorkNodeId = common.LocalWorkNodeId
	_, err := ormer.Insert(task)
	if err != nil {
		return errors.Wrapf(err, "orm insert metric monitor task[%s] err", task.Id)
	}
	return nil
}

// DeleteMetricMonitorTask ...
func DeleteMetricMonitorTask(id string) error {
	_, err := ormer.QueryTable(tableNameMetricMonitorTask).
		Filter(fieldNameIsDeleted, notDeletedFlag).Filter(fieldNameId, id).
		Update(orm.Params{
			fieldNameIsDeleted: deletedFlag,
			fieldNameDeleteAt:  time.Now().UTC()})
	if err != nil {
		if errors.Is(err, orm.ErrNoRows) {
			return nil
		}
		return errors.Wrapf(err, "delete metric monitor task[%s] err", id)
	}
	return nil
}

// UpdateMetricMonitorTask ...
func UpdateMetricMonitorTask(id string, value int32) error {
	_, err := ormer.QueryTable(tableNameMetricMonitorTask).
		Filter(fieldNameIsDeleted, notDeletedFlag).Filter(fieldNameId, id).
		Update(orm.Params{
			fieldNameTargetValue: value,
			fieldNameUpdateAt:    time.Now().UTC()})
	if err != nil {
		if errors.Is(err, orm.ErrNoRows) {
			return nil
		}
		return errors.Wrapf(err, "update metric monitor task[%s] err", id)
	}
	return nil
}

// GetMetricMonitorTaskById ...
func GetMetricMonitorTaskById(id string) (*MetricMonitorTask, error) {
	return getMetricMonitorTaskByFilters(Filters{
		fieldNameIsDeleted: notDeletedFlag,
		fieldNameId:        id,
	})
}

// GetMetricMonitorTaskByScalingPolicy ...
func GetMetricMonitorTaskByScalingPolicy(policyId string) (*MetricMonitorTask, error) {
	return getMetricMonitorTaskByFilters(Filters{
		fieldNameIsDeleted:       notDeletedFlag,
		fieldNameScalingPolicyId: policyId,
	})
}

// GetMetricMonitorTasksByWorkNodeId ...
func GetMetricMonitorTasksByWorkNodeId(wnId string) ([]*MetricMonitorTask, error) {
	return getMetricMonitorTasksByFilters(Filters{
		fieldNameIsDeleted:  notDeletedFlag,
		fieldNameWorkNodeId: wnId,
	})
}

func getMetricMonitorTaskByFilters(f Filters) (*MetricMonitorTask, error) {
	var task MetricMonitorTask
	if err := f.Filter(tableNameMetricMonitorTask).RelatedSel().One(&task); err != nil {
		return nil, errors.Wrapf(err, "read metric menitor task by Filters[%s] err", f)
	}
	return &task, nil
}

func getMetricMonitorTasksByFilters(f Filters) ([]*MetricMonitorTask, error) {
	var tasks []*MetricMonitorTask
	_, err := f.Filter(tableNameMetricMonitorTask).All(&tasks)
	if err != nil {
		return nil, errors.Wrapf(err, "get all metirc monitor task from db err")
	}
	return tasks, nil
}

// TakeOverMetricMonitorTasks ...
func TakeOverMetricMonitorTasks(taskIds []string, takeOverId string) error {
	if len(taskIds) == 0 {
		return nil
	}
	_, err := ormer.QueryTable(tableNameMetricMonitorTask).
		Filter(fieldNameIdIn, taskIds).
		Update(orm.Params{
			fieldNameWorkNodeId: takeOverId,
		})
	return err
}
