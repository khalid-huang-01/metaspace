// Copyright (c) Huawei Technologies Co., Ltd. 2022-2022. All rights reserved.

// fleet数据表定义
package dao

import (
	"fleetmanager/db/dbm"
	"github.com/beego/beego/v2/client/orm"
	"time"
)

const (
	FleetStateCreating   = "CREATING"
	FleetStateActive     = "ACTIVE"
	FleetStateDeleting   = "DELETING"
	FleetStateError      = "ERROR"
	FleetStateTerminated = "TERMINATED"
)

// Fleet TODO: BuildId需要关联应用包表
type Fleet struct {
	Id                                      string    `orm:"column(id);size(64);pk" json:"id"`
	ProjectId                               string    `orm:"column(project_id);size(64)" json:"project_id"`
	Name                                    string    `orm:"column(name);type(text);null" json:"name"`
	Description                             string    `orm:"column(description);type(text);null" json:"description"`
	BuildId                                 string    `orm:"column(build_id);size(64)" json:"build_id"`
	Region                                  string    `orm:"column(region);size(64)" json:"region"`
	Bandwidth                               int       `orm:"column(bandwidth);type(int);default(100)" json:"bandwidth"`
	InstanceSpecification                   string    `orm:"column(instance_specification);size(128)" json:"instance_specification"`
	ServerSessionProtectionPolicy           string    `orm:"column(server_session_protection_policy);size(32)" json:"server_session_protection_policy"`
	ServerSessionProtectionTimeLimitMinutes int       `orm:"column(server_session_protection_time_limit_minutes);type(int);default(5)" json:"server_session_protection_time_limit_minutes"`
	EnableAutoScaling                       bool      `orm:"column(enable_auto_scaling);default(False)" json:"enable_auto_scaling"`
	ScalingIntervalMinutes                  int       `orm:"column(scaling_interval_minutes);type(int);default(10)" json:"scaling_interval_minutes"`
	InstanceType                            string    `orm:"column(instance_type);size(8)" json:"instance_type"`
	InstanceTags                            string    `orm:"column(instance_tags);size(1024)" json:"instance_tags"`
	Minimum                                 int       `orm:"column(minimum);type(int)" json:"minimum"`
	Maximum                                 int       `orm:"column(maximum);type(int)" json:"maximum"`
	Desired                                 int       `orm:"column(desired);type(int)" json:"desired"`
	OperatingSystem                         string    `orm:"column(operating_system);size(32)" json:"operating_system"`
	State                                   string    `orm:"column(state);size(32)" json:"state"`
	CreationTime                            time.Time `orm:"column(creation_time);type(datetime);auto_now_add" json:"creation_time"`
	UpdateTime                              time.Time `orm:"column(update_time);type(datetime);auto_now" json:"update_time"`
	Terminated                              bool      `orm:"column(terminated);default(False)" json:"terminated"`
	TerminationTime                         time.Time `orm:"column(termination_time);auto_now" json:"termination_time"`
	PolicyPeriodInMinutes                   int       `orm:"column(policy_period_in_minutes);type(int);default(1)" json:"policy_period_in_minutes"`
	NewSessionsPerCreator                   int       `orm:"column(new_sessions_per_creator);type(int);default(1)" json:"new_sessions_per_creator"`
	EnterpriseProjectId                     string    `orm:"column(enterprise_project_id);size(64)" json:"enterprise_project_id"`
}

type fleetStorage struct{}

var fs = fleetStorage{}

// GetFleetStorage 获取fleet存储对象
func GetFleetStorage() *fleetStorage {
	return &fs
}

// 添加fleet索引
func (s *fleetStorage) TableIndex() [][]string {
	return [][]string{
		{"id"},
	}
}

// Insert 插入Fleet
func (s *fleetStorage) Insert(f *Fleet) error {
	_, err := dbm.Ormer.Insert(f)

	return err
}

// Update 更新Fleet
func (s *fleetStorage) Update(f *Fleet, cols ...string) error {
	_, err := dbm.Ormer.Update(f, cols...)

	return err
}

// Get 获取Fleet详情
func (s *fleetStorage) Get(f Filters) (*Fleet, error) {
	var fleet Fleet

	err := f.Filter(FleetTable).One(&fleet)
	if err != nil {
		return nil, err
	}

	return &fleet, err
}

// Get 获取Fleet列表
func (s *fleetStorage) List(f Filters, offset int, limit int) ([]Fleet, error) {
	var fleets []Fleet
	_, err := f.Filter(FleetTable).Offset(offset).Limit(limit).All(&fleets)

	return fleets, err
}

// Count 获取Fleet个数
func (s *fleetStorage) Count(f Filters) (int64, error) {
	count, err := f.Filter(FleetTable).Count()

	return count, err
}

func QueryFleetByCondition(projectId string, offset int, limit int, id string,
	name string, state string) ([]Fleet, int64, error) {
	var list []Fleet

	qs := dbm.Ormer.QueryTable(FleetTable)
	condition := orm.NewCondition()
	condition = condition.And("project_id", projectId)
	if name != "" {
		condition = condition.And("name__contains", name)
	}
	if id != "" {
		condition = condition.And("id__contains", id)
	}
	if state != "" {
		condition = condition.And("state", state)
	}
	condition = condition.AndNot("state__in", FleetStateTerminated)

	total, err := qs.SetCond(condition).All(&list)
	if err != nil {
		return nil, 0, err
	}

	_, err = qs.SetCond(condition).Offset((offset) * limit).Limit(limit).OrderBy("-creation_time").All(&list)
	if err != nil {
		return nil, 0, err
	}
	return list, total, nil
}
