// Copyright (c) Huawei Technologies Co., Ltd. 2022-2022. All rights reserved.

// 弹性伸缩组数据表定义
package dao

import "fleetmanager/db/dbm"

type ScalingGroup struct {
	Id                 string `orm:"column(id);size(64);pk" json:"id"`
	FleetId            string `orm:"column(fleet_id);size(64)" json:"fleet_id"`
	RegionId           string `orm:"column(region_id);size(64)" json:"region_id"`
	VpcId              string `orm:"column(vpc_id);size(64)" json:"vpc_id"`
	SubnetId           string `orm:"column(subnet_id);size(64)" json:"subnet_id"`
	SecurityGroupId    string `orm:"column(security_group_id);size(64)" json:"security_group_id"`
	ResourceProjectId  string `orm:"column(resource_project_id);size(64)" json:"resource_project_id"`
	ResourceDomainId   string `orm:"column(resource_domain_id);size(64)" json:"resource_domain_id"`
	ResourceAgencyName string `orm:"column(resource_agency_name)" json:"resource_agency_name"`
}

type scalingGroupStorage struct{}

var ss = scalingGroupStorage{}

// GetScalingGroupStorage 获取弹性伸缩存储
func GetScalingGroupStorage() *scalingGroupStorage {
	return &ss
}

// Insert 插入弹性伸缩组信息
func (s *scalingGroupStorage) InsertOrUpdate(g *ScalingGroup) error {
	_, err := dbm.Ormer.InsertOrUpdate(g)
	return err
}

// Update 更新弹性伸缩
func (s *scalingGroupStorage) Update(g *ScalingGroup, cols ...string) error {
	_, err := dbm.Ormer.Update(g, cols...)

	return err
}

// GetOne 获取弹性伸缩
func (s *scalingGroupStorage) GetOne(f Filters) (*ScalingGroup, error) {
	var g ScalingGroup
	err := f.Filter(ScalingGroupTable).One(&g)
	return &g, err
}

// List 查询弹性伸缩列表
func (s *scalingGroupStorage) List(f Filters, offset int, limit int) ([]ScalingGroup, error) {
	var gs []ScalingGroup
	_, err := f.Filter(ScalingGroupTable).Offset(offset).Limit(limit).All(&gs)
	return gs, err
}

// Count 查询弹性伸缩计数
func (s *scalingGroupStorage) Count(f Filters) (int64, error) {
	count, err := f.Filter(ScalingGroupTable).Count()
	return count, err
}
