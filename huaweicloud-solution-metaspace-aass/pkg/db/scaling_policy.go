// Copyright (c) Huawei Technologies Co., Ltd. 2022-2022. All rights reserved.

// 弹性伸缩策略数据表定义
package db

import (
	"time"

	"github.com/beego/beego/v2/client/orm"
	"github.com/pkg/errors"

	"scase.io/application-auto-scaling-service/pkg/common"
)

const (
	tableNameScalingPolicy = "scaling_policy"
)

type ScalingPolicy struct {
	ScalingGroup *ScalingGroup `orm:"rel(fk)"`
	Id           string        `orm:"column(id);size(64);pk"`
	Name         string        `orm:"column(name);size(64)"`
	PolicyType   string        `orm:"column(policy_type);size(64)"`
	PolicyConfig string        `orm:"column(policy_config);size(512)"`
	ProjectId    string        `orm:"column(project_id);size(64)"`
	TimeModel
}

func (p *ScalingPolicy) replace(newPolicy *ScalingPolicy) {
	if newPolicy.Id != "" {
		p.Id = newPolicy.Id
	}
	if newPolicy.Name != "" {
		p.Name = newPolicy.Name
	}
	if newPolicy.PolicyType != "" {
		p.PolicyType = newPolicy.PolicyType
	}
	if newPolicy.PolicyConfig != "" {
		p.PolicyConfig = newPolicy.PolicyConfig
	}
}

// AddScalingPolicy add ScalingPolicy
func AddScalingPolicy(policy *ScalingPolicy) error {
	if policy == nil || policy.Id == "" {
		return errors.Errorf("func AddScalingPolicy[%+v] has invalid args", policy)
	}

	policy.IsDeleted = notDeletedFlag
	_, err := ormer.Insert(policy)
	if err != nil {
		return errors.Wrapf(err, "add scaling policy[%s] err", policy.Id)
	}
	return nil
}

// DeleteScalingPolicy delete ScalingPolicy
func DeleteScalingPolicy(projectId, policyId string) error {
	_, err := ormer.QueryTable(tableNameScalingPolicy).Filter(fieldNameIsDeleted, notDeletedFlag).
		Filter(fieldNameProjectId, projectId).Filter(fieldNameId, policyId).
		Update(orm.Params{fieldNameIsDeleted: deletedFlag, fieldNameDeleteAt: time.Now().UTC()})
	if err != nil {
		if errors.Is(err, orm.ErrNoRows) {
			return nil
		}
		return errors.Wrapf(err, "delete scaling policy[%s] err", policyId)
	}
	return nil
}

// UpdateScalingPolicyById update ScalingPolicy
func UpdateScalingPolicyById(policyId string, new *ScalingPolicy) error {
	var err error

	curPolicy := &ScalingPolicy{Id: policyId}
	if err = ormer.Read(curPolicy); err != nil {
		return errors.Wrapf(err, "read scaling policy[%s] err", policyId)
	}

	curPolicy.replace(new)
	if _, err = ormer.Update(curPolicy); err != nil {
		return errors.Wrapf(err, "update scaling policy[%s] err", policyId)
	}

	return nil
}

// GetScalingPolicyById get ScalingPolicy detail
func GetScalingPolicyById(projectId, policyId string) (*ScalingPolicy, error) {
	var policy ScalingPolicy
	err := ormer.QueryTable(tableNameScalingPolicy).Filter(fieldNameIsDeleted, notDeletedFlag).
		Filter(fieldNameProjectId, projectId).Filter(fieldNameId, policyId).RelatedSel().One(&policy)
	if err != nil {
		return nil, errors.Wrapf(err, "read scaling policy[%s] err", policyId)
	}
	return &policy, nil
}

// IsScalingPolicyExist check whether the ScalingPolicy exists
func IsScalingPolicyExist(projectId, policyId string) bool {
	return ormer.QueryTable(tableNameScalingPolicy).Filter(fieldNameIsDeleted, notDeletedFlag).
		Filter(fieldNameProjectId, projectId).Filter(fieldNameId, policyId).Exist()
}

// IsTargetBasedPolicyExistInScalingGroup check whether the target based policy exists in the ScalingGroup
func IsTargetBasedPolicyExistInScalingGroup(projectId, groupId string) bool {
	return ormer.QueryTable(tableNameScalingPolicy).Filter(fieldNameIsDeleted, notDeletedFlag).
		Filter(fieldNameProjectId, projectId).Filter(fieldNamePolicyType, common.PolicyTypeTargetBased).
		Filter(fieldNameScalingGroupId, groupId).Exist()
}

// ListScalingPolicyByType filter policy type list ScalingPolicy
func ListScalingPolicyByType(policyType string) ([]*ScalingPolicy, error) {
	var list []*ScalingPolicy
	_, err := ormer.QueryTable(tableNameScalingPolicy).Filter(fieldNamePolicyType, policyType).
		Filter(fieldNameIsDeleted, notDeletedFlag).All(&list)
	if err != nil {
		return nil, errors.Wrapf(err, "list type[%s] scaling policy err", policyType)
	}
	return list, nil
}
