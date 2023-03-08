// Copyright (c) Huawei Technologies Co., Ltd. 2022-2022. All rights reserved.

// cidr数据表定义
package dao

import (
	"fleetmanager/db/dbm"
	"time"
)

type FleetVpcCidr struct {
	Id         string    `orm:"column(id);size(64);pk" json:"id"`
	VpcCidr    string    `orm:"column(vpc_cidr);size(128)" json:"vpc_cidr"`
	FleetId    string    `orm:"column(fleet_id);size(64)" json:"fleet_id"`
	Namespace  string    `orm:"column(namespace);size(128)" json:"namespace"`
	CreateTime time.Time `orm:"column(creation_time);type(datetime);auto_now_add" json:"creation_time"`
}

// TableUnique 设置联合主键
func (fvc *FleetVpcCidr) TableUnique() [][]string {
	return [][]string{
		{"VpcCidr", "Namespace"},
	}
}

// InsertFleetVpcCidr 插入Fleet vpc cidr
func InsertFleetVpcCidr(fvc *FleetVpcCidr) error {
	_, err := dbm.Ormer.Insert(fvc)
	return err
}

// DeleteFleetVpcCidr 删除Fleet vpc cidr
func DeleteFleetVpcCidr(fvc *FleetVpcCidr) error {
	_, err := dbm.Ormer.Delete(fvc)
	return err
}

// GetAllFleetVpcCidr 获取所有的Fleet vpc cidr
func GetAllFleetVpcCidr() ([]FleetVpcCidr, error) {
	var fvc []FleetVpcCidr
	_, err := dbm.Ormer.QueryTable(FleetVpcCidrTable).All(&fvc)
	return fvc, err
}

// GetFleetVpcCidrByFleetId 获取fleet id关联的vpc cidr
func GetFleetVpcCidrByFleetId(fleetId string) (*FleetVpcCidr, error) {
	var fvc FleetVpcCidr
	err := dbm.Ormer.QueryTable(FleetVpcCidrTable).Filter("FleetId", fleetId).One(&fvc)
	return &fvc, err
}
