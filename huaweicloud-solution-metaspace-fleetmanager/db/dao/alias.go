// Copyright (c) Huawei Technologies Co., Ltd. 2022-2022. All rights reserved.

// alias数据表定义
package dao

import (
	"fleetmanager/db/dbm"
	"time"
)

// alias
const (
	AliasTypeActive     = "ACTIVE"
	AliasTypeDeactive   = "DEACTIVE"
	AliasTypeTerminated = "TERMINATED"
)
type Alias struct {
	Id               string    `orm:"column(id);size(64);pk" json:"id"`
	ProjectId        string    `orm:"column(project_id);size(64)" json:"project_id"`
	Name             string    `orm:"column(name);size(1024)" json:"name"`
	Description      string    `orm:"column(description);size(1024)" json:"description"`
	CreationTime     time.Time `orm:"column(creation_time);type(datetime);auto_now_add" json:"creation_time"`
	UpdateTime       time.Time `orm:"column(update_time);type(datetime);auto_now" json:"update_time"`
	AssociatedFleets string    `orm:"column(associated_fleets);size(2048)" json:"associated_fleets"`
	Type             string    `orm:"column(type);size(16)" json:"type"`
	Message          string    `orm:"column(message);size(1024)" json:"message"`
}
type aliasStorage struct{}

var as = aliasStorage{}

// GetAliasStorage 获取alias存储对象
func GetAliasStorage() *aliasStorage {
	return &as
}

func (s *aliasStorage) TableIndex() [][]string {
	return [][]string{{"Id"}}
}

// Insert 插入Alias
func (s *aliasStorage) Insert(f *Alias) error {
	_, err := dbm.Ormer.Insert(f)
	return err
}

// Update 更新Alias
func (s *aliasStorage) Update(f *Alias, cols ...string) error {
	_, err := dbm.Ormer.Update(f, cols...)
	return err
}

// Get 获取Fleet详情
func (s *aliasStorage) Get(f Filters) (*Alias, error) {
	var alias Alias
	err := f.Filter(AliasTable).One(&alias)
	if err != nil {
		return nil, err
	}
	return &alias, err
}

// Get 获取Alias列表
func (s *aliasStorage) List(f Filters, offset int, limit int) ([]Alias, error) {
	var aliases []Alias
	_, err := dbm.Ormer.QueryTable(&Alias{}).SetCond(f.Condition()).Offset(offset).Limit(limit).All(&aliases)
	return aliases, err
}

// Count 获取Alias个数
func (s *aliasStorage) Count(f Filters) (int64, error) {
	count, err := dbm.Ormer.QueryTable(&Alias{}).SetCond(f.Condition()).Count()
	return count, err
}

// Delete: 删除 Alieas Db
func (s *aliasStorage) Delete(aliasId string, projectId string) error {
	_, err := dbm.Ormer.QueryTable(AliasTable).Filter("ProjectId", projectId).Filter("Id", aliasId).Delete()
	if err != nil {
		return err
	}
	return nil
}
