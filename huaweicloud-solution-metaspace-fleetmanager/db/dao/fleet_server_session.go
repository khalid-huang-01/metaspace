// Copyright (c) Huawei Technologies Co., Ltd. 2022-2022. All rights reserved.

// fleet服务端会话数据表定义
package dao

import (
	"fleetmanager/db/dbm"
	"time"
)

// FleetServerSession TODO:fleetId作为外键关联fleet表
type FleetServerSession struct {
	Id              string    `orm:"column(id);size(128);pk"`
	FleetId         string    `orm:"column(fleet_id);size(128)"`
	ServerSessionId string    `orm:"column(server_session_id);size(128)"`
	Region          string    `orm:"column(region);size(64)" json:"region"`
	CreationTime    time.Time `orm:"column(creation_time);type(datetime);auto_now_add" json:"creation_time"`
}

type fleetSessionStorage struct{}

var fss = fleetSessionStorage{}

// GetFleetServerSessionStorage 获取Fleet服务器会话存储
func GetFleetServerSessionStorage() *fleetSessionStorage {
	return &fss
}

// Insert 插入Fleet服务器会话信息
func (s *fleetSessionStorage) Insert(fleetSession *FleetServerSession) error {
	_, err := dbm.Ormer.Insert(fleetSession)
	return err
}

func (s *fleetSessionStorage) TableIndex() [][]string {
	return [][]string{
		{"server_session_id"},
	}
}

// GetOne 获取Fleet服务器会话信息
func (s *fleetSessionStorage) GetOne(f Filters) (*FleetServerSession, error) {
	var fleetSession FleetServerSession
	err := f.Filter(FleetServerSessionTable).One(&fleetSession)
	return &fleetSession, err
}
