// Copyright (c) Huawei Technologies Co., Ltd. 2022-2022. All rights reserved.

// fleet事件数据表定义
package dao

import (
	"fleetmanager/db/dbm"
	"time"
)

// FleetEvent TODO:fleetId作为外键关联fleet表
type FleetEvent struct {
	Id              string    `orm:"column(id);size(64);pk"`
	FleetId         string    `orm:"column(fleet_id);size(64)"`
	EventCode       string    `orm:"column(event_code);size(128)"`
	EventTime       time.Time `orm:"column(event_time);type(datetime);auto_now_add"`
	Message         string    `orm:"column(message);size(1600)"`
	PreSignedLogUrl string    `orm:"column(pre_signed_log_url);size(128)"`
}

type fleetEventStorage struct{}

var es = fleetEventStorage{}

// GetFleetEventStorage 获取fleet event存储
func GetFleetEventStorage() *fleetEventStorage {
	return &es
}

// Insert 插入fleet event
func (s *fleetEventStorage) Insert(e *FleetEvent) error {
	_, err := dbm.Ormer.Insert(e)
	return err
}

// Update 更新fleet event
func (s *fleetEventStorage) Update(e *FleetEvent, cols ...string) error {
	_, err := dbm.Ormer.Update(e, cols...)
	return err
}

// Get 获取fleet event
func (s *fleetEventStorage) Get(f Filters) (*FleetEvent, error) {
	var event FleetEvent
	err := f.Filter(FleetEventTable).One(&event)
	if err != nil {
		return nil, err
	}

	return &event, err
}

// List 获取fleet event列表
func (s *fleetEventStorage) List(f Filters, offset int, limit int) ([]FleetEvent, error) {
	var events []FleetEvent
	_, err := f.Filter(FleetEventTable).Offset(offset).Limit(limit).All(&events)

	return events, err
}

// Count 获取fleet event计数
func (s *fleetEventStorage) Count(f Filters) (int64, error) {
	count, err := f.Filter(FleetEventTable).Count()

	return count, err
}
