// Copyright (c) Huawei Technologies Co., Ltd. 2022-2022. All rights reserved.

// 管理监控策略
package metricmonitor

import (
	"fmt"
	"sync"

	"github.com/beego/beego/v2/client/orm"
	"github.com/pkg/errors"
	"github.com/robfig/cron/v3"

	"scase.io/application-auto-scaling-service/pkg/db"
	"scase.io/application-auto-scaling-service/pkg/metricmonitor/influxdb"
	"scase.io/application-auto-scaling-service/pkg/setting"
	"scase.io/application-auto-scaling-service/pkg/utils/logger"
)

var (
	metricMonitorManager *metricMonitorMgmt
)

// Init ...
func Init() error {
	var err error
	metricMonitorManager, err = newMetricMonitorMgmt()
	if err != nil {
		return err
	}
	return nil
}

func newWithSeconds() *cron.Cron {
	secondParser := cron.NewParser(cron.Second | cron.Minute |
		cron.Hour | cron.Dom | cron.Month | cron.DowOptional | cron.Descriptor)
	return cron.New(cron.WithParser(secondParser), cron.WithChain())
}

func getCronDuration(duration string) string {
	return "@every " + duration
}

type metricMonitorMgmt struct {
	metricCtr *influxdb.Controller
	cronTab   *cron.Cron
	period    string
	taskMgmt  map[string]cron.EntryID
	lock      sync.Mutex
}

func newMetricMonitorMgmt() (*metricMonitorMgmt, error) {
	var err error
	m := metricMonitorMgmt{
		cronTab:  newWithSeconds(),
		period:   getCronDuration(setting.GetMonitorDuration()),
		taskMgmt: make(map[string]cron.EntryID),
	}
	m.metricCtr, err = influxdb.NewController()
	if err != nil {
		return nil, err
	}
	return &m, nil
}

// GetMgmt ...
func GetMgmt() *metricMonitorMgmt {
	if metricMonitorManager == nil {
		logger.R.Error("metric monitor is not initialized correctly")
		return nil
	}
	return metricMonitorManager
}

// NewTaskForPolicy ...
func (m *metricMonitorMgmt) NewTaskForPolicy(groupId, policyId, metric string,
	value int32) (*db.MetricMonitorTask, error) {
	task := &db.MetricMonitorTask{
		Id:              policyId,
		ScalingGroupID:  groupId,
		ScalingPolicyID: policyId,
		MetricName:      metric,
		TargetValue:     value,
	}
	err := db.AddMetricMonitorTask(task)
	if err != nil {
		return nil, err
	}
	return task, nil
}

// AddTask ...
func (m *metricMonitorMgmt) AddTask(task *db.MetricMonitorTask) error {
	m.lock.Lock()
	defer m.lock.Unlock()
	entryID, err := m.cronTab.AddFunc(m.period, func() {
		log := logger.R.WithField(logger.MetricMonTask, fmt.Sprintf("scaling_policy[%s]", task.ScalingPolicyID))
		mt, err := db.GetMetricMonitorTaskById(task.Id)
		if err != nil {
			if errors.Is(err, orm.ErrNoRows) {
				log.Warn("task is not found, delete the task from metricMonitorMgmt")
				m.deleteTask(task.Id)
				return
			}
			log.Error("read metric monitor task from db err: %+v", err)
			return
		}

		targetBasedMonitorTask(log, m.metricCtr, mt)
	})
	if err != nil {
		return err
	}

	m.taskMgmt[task.Id] = entryID
	m.cronTab.Start()
	return nil
}

// TaskIdForPolicy ...
func (m *metricMonitorMgmt) TaskIdForPolicy(policyId string) string {
	task, err := db.GetMetricMonitorTaskByScalingPolicy(policyId)
	if err != nil {
		return ""
	}
	return task.Id
}

func (m *metricMonitorMgmt) deleteTask(taskId string) {
	m.lock.Lock()
	entryId := m.taskMgmt[taskId]
	m.cronTab.Remove(entryId)
	delete(m.taskMgmt, taskId)
	m.lock.Unlock()
}

// DeleteTask ...
func (m *metricMonitorMgmt) DeleteTask(taskId string) error {
	if len(taskId) == 0 {
		return errors.New("task id cannot be empty")
	}
	return db.DeleteMetricMonitorTask(taskId)
}

// UpdateTask ...
func (m *metricMonitorMgmt) UpdateTask(taskId string, targetValue int32) error {
	if len(taskId) == 0 {
		return errors.New("task id cannot be empty")
	}
	return db.UpdateMetricMonitorTask(taskId, targetValue)
}
