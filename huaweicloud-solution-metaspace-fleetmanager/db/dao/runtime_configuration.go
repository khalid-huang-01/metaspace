// Copyright (c) Huawei Technologies Co., Ltd. 2022-2022. All rights reserved.

// 运行配置数据表定义
package dao

import "fleetmanager/db/dbm"

type RuntimeConfiguration struct {
	Id                                    string `orm:"column(id);size(64);pk" json:"id"`
	FleetId                               string `orm:"column(fleet_id);size(64)" json:"fleet_id"`
	ServerSessionActivationTimeoutSeconds int    `orm:"column(server_session_activation_timeout_seconds);type(int);default(120)" json:"server_session_activation_timeout_seconds"`
	MaxConcurrentServerSessionsPerProcess int    `orm:"column(max_concurrent_server_sessions_per_process);type(int);default(1)" json:"max_concurrent_server_sessions_per_process"`
	ProcessConfigurations                 string `orm:"column(process_configurations);type(text)" json:"process_configurations"`
}

type runtimeConfigurationStorage struct{}

var rs = runtimeConfigurationStorage{}

// GetRuntimeConfigurationStorage 获取运行配置存储
func GetRuntimeConfigurationStorage() *runtimeConfigurationStorage {
	return &rs
}

// Insert 插入运行配置
func (s *runtimeConfigurationStorage) Insert(r *RuntimeConfiguration) error {
	_, err := dbm.Ormer.Insert(r)
	return err
}

// Update 更新运行配置
func (s *runtimeConfigurationStorage) Update(r *RuntimeConfiguration, cols ...string) error {
	_, err := dbm.Ormer.Update(r, cols...)
	return err
}

// Get 获取运行配置
func (s *runtimeConfigurationStorage) Get(f Filters) (*RuntimeConfiguration, error) {
	var r RuntimeConfiguration
	err := f.Filter(RuntimeConfigurationTable).One(&r)
	if err != nil {
		return nil, err
	}
	return &r, err
}

// List 查询运行配置列表
func (s *runtimeConfigurationStorage) List(f Filters, offset int, limit int) ([]RuntimeConfiguration, error) {
	var rs []RuntimeConfiguration
	_, err := f.Filter(RuntimeConfigurationTable).Offset(offset).Limit(limit).All(&rs)
	return rs, err
}

// Count 获取运行配置计数
func (s *runtimeConfigurationStorage) Count(f Filters) (int64, error) {
	count, err := f.Filter(RuntimeConfigurationTable).Count()
	return count, err
}
