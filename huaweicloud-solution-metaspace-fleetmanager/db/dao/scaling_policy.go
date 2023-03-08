// Copyright (c) Huawei Technologies Co., Ltd. 2022-2022. All rights reserved.

// 弹性伸缩策略数据表定义
package dao

import (
	"fleetmanager/db/dbm"
)

const (
	TargetBasedPolicy = "TARGET_BASED"
	RuleBasedPolicy   = "RULE_BASED"
)

const (
	PolicyStateActive   = "ACTIVE"
	PolicyStateUpdating = "UPDATING"
	PolicyStateError    = "ERROR"
)

type ScalingPolicy struct {
	Id                       string `orm:"column(id);size(64);pk"`
	Name                     string `orm:"column(name);type(text);null"`
	FleetId                  string `orm:"column(fleet_id);size(64)"`
	PolicyType               string `orm:"column(policy_type);size(16)"`
	ScalingTarget            string `orm:"column(scaling_target);size(16)"`
	State                    string `orm:"column(state);size(32)"`
	TargetBasedConfiguration string `orm:"column(target_based_configuration);type(text);null"`
	RuleBasedConfiguration   string `orm:"column(rule_based_configuration);type(text);null"`
	ResourceProjectId        string `orm:"column(resource_project_id);size(64)" json:"resource_project_id"`
}

type scalingPolicyStorage struct{}

var policyS = scalingPolicyStorage{}

// GetScalingPolicyStorage 获取策略存储
func GetScalingPolicyStorage() *scalingPolicyStorage {
	return &policyS
}

// Insert 插入弹性伸缩策略
func (s *scalingPolicyStorage) Insert(p *ScalingPolicy) error {
	_, err := dbm.Ormer.Insert(p)
	return err
}

// Update 更新弹性伸缩策略
func (s *scalingPolicyStorage) Update(p *ScalingPolicy, cols ...string) error {
	_, err := dbm.Ormer.Update(p, cols...)
	return err
}

// Get 获取弹性伸缩
func (s *scalingPolicyStorage) Get(f Filters) (*ScalingPolicy, error) {
	var p ScalingPolicy
	return &p, f.Filter(ScalingPolicyTable).One(&p)
}

// List 查询弹性伸缩列表
func (s *scalingPolicyStorage) List(f Filters, offset int, limit int) ([]ScalingPolicy, error) {
	var policies []ScalingPolicy
	_, err := f.Filter(ScalingPolicyTable).Offset(offset).Limit(limit).All(&policies)
	return policies, err
}

// Count 获取弹性伸缩策略计数
func (s *scalingPolicyStorage) Count(f Filters) (int64, error) {
	return f.Filter(ScalingPolicyTable).Count()
}

// Delete 删除弹性伸缩计数
func (s *scalingPolicyStorage) Delete(p *ScalingPolicy) error {
	_, err := dbm.Ormer.Delete(p)

	return err
}
