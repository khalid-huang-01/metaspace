// Copyright (c) Huawei Technologies Co., Ltd. 2022-2022. All rights reserved.

// 入站规则定义
package dao

import "fleetmanager/db/dbm"

// InboundPermission TODO:fleetId作为外键关联fleet表
type InboundPermission struct {
	Id              string `orm:"column(id);size(64);pk" json:"id"`
	FleetId         string `orm:"column(fleet_id);size(64)" json:"fleet_id"`
	SecurityGroupId string `orm:"column(security_group_id);size(64)" json:"security_group_id"`
	Protocol        string `orm:"column(protocol);size(8)" json:"protocol"`
	IpRange         string `orm:"column(ip_range);size(32)" json:"ip_range"`
	FromPort        int32  `orm:"column(from_port);type(int);default(1025)" json:"from_port"`
	ToPort          int32  `orm:"column(to_port);type(int);default(1025)" json:"to_port"`
}

type permissionStorage struct{}

var ps = permissionStorage{}

// GetPermissionStorage 获取inbound permission存储
func GetPermissionStorage() *permissionStorage {
	return &ps
}

// InsertOrUpdate 插入或者更新inbound permission
func (s *permissionStorage) InsertOrUpdate(p *InboundPermission) error {
	_, err := dbm.Ormer.InsertOrUpdate(p)
	return err
}

// Update 更新inbound permission
func (s *permissionStorage) Update(p *InboundPermission, cols ...string) error {
	_, err := dbm.Ormer.Update(p, cols...)

	return err
}

// Get 获取inbound permission
func (s *permissionStorage) Get(f Filters) (*InboundPermission, error) {
	var p InboundPermission

	err := f.Filter(InboundPermissionTable).One(&p)
	if err != nil {
		return nil, err
	}

	return &p, err
}

// List 查询inbound permission列表
func (s *permissionStorage) List(f Filters, offset int, limit int) ([]InboundPermission, error) {
	var ps []InboundPermission
	_, err := f.Filter(InboundPermissionTable).Offset(offset).Limit(limit).All(&ps)

	return ps, err
}

// Count 获取inbound permission计数
func (s *permissionStorage) Count(f Filters) (int64, error) {
	count, err := f.Filter(InboundPermissionTable).Count()

	return count, err
}

// Delete 删除inbound permission
func (s *permissionStorage) Delete(p *InboundPermission) error {
	_, err := dbm.Ormer.Delete(p)

	return err
}
