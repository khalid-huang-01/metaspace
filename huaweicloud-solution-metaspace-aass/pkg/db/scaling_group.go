// Copyright (c) Huawei Technologies Co., Ltd. 2022-2022. All rights reserved.

// 弹性伸缩数据表定义
package db

import (
	"time"

	"github.com/beego/beego/v2/client/orm"
	"github.com/pkg/errors"
)

const (
	tableNameScalingGroup = "scaling_group"

	// ScalingGroup 状态变化
	// creating → stable/scaling → deleting → deleted
	//     ↓                           ↑
	//       —— —— —— →  error —— —— —→
	ScalingGroupStateCreating = "creating" // creating状态，外部不可见，仅用于记录正在创建的资源
	ScalingGroupStateStable   = "stable"
	ScalingGroupStateScaling  = "scaling"
	ScalingGroupStateError    = "error" // error状态，外部不可见，仅用于标记需要清理的资源
	ScalingGroupStateDeleting = "deleting"
	ScalingGroupStateDeleted  = "deleted" // deleted状态，外部不可见，仅用于记录已删除的资源
)

var groupInvisibleStates = []string{ScalingGroupStateCreating, ScalingGroupStateError, ScalingGroupStateDeleted}

type ScalingGroup struct {
	// InstanceConfiguration：实例配置（运行时相关）
	InstanceConfiguration *InstanceConfiguration `orm:"rel(one)"`
	ScalingPolicies       []*ScalingPolicy       `orm:"reverse(many)"`
	Id                    string                 `orm:"column(id);size(128);pk"`
	Name                  string                 `orm:"column(name);size(128)"`
	// InstanceType：实例类型：vm/pod
	InstanceType string `orm:"column(instance_type);size(64)"`
	// ResourceId：资源Id，当实例类型为vm时对应为VmScalingGroup.Id
	ResourceId           string `orm:"column(resource_id);size(128)"`
	MinInstanceNumber    int32  `orm:"column(min_instance_number);type(int);default(0)"`
	MaxInstanceNumber    int32  `orm:"column(max_instance_number);type(int);default(200)"`
	DesireInstanceNumber int32  `orm:"column(desire_instance_number);type(int);default(1)"`
	// CoolDownTime：冷却时长，单位min
	CoolDownTime int64 `orm:"column(cool_down_time);type(int);default(5)"`
	// AutoScalingTimestamp: 自动伸缩的时间戳，单位：ns
	AutoScalingTimestamp int64  `orm:"column(auto_scaling_timestamp);type(int);default(0)"`
	VpcId                string `orm:"column(vpc_id);size(64)"`
	SubnetId             string `orm:"column(subnet_id);size(64)"`
	// VmTemplate：vm实例配置模板
	VmTemplate string `orm:"column(vm_template);size(2048)"`
	// PodTemplate：pod实例配置模板
	PodTemplate string `orm:"column(pod_template);size(2048)"`
	// FleetId：所属进程队列Id
	FleetId           string `orm:"column(fleet_id);size(128)"`
	EnableAutoScaling bool   `orm:"column(enable_auto_scaling);type(bool);size(128)"`
	State             string `orm:"column(state);size(64);default(creating)"`
	IsInvisible       string `orm:"column(is_invisible);default(1)"` // 当伸缩组处于stable状态之后，外部可见
	ProjectId         string `orm:"column(project_id);size(64)"`
	EnterpriseProjectId string `orm:"colum(enterprise_project_id);size(64)"`
	TimeModel
	InstanceTags 		string `orm:"colume(instance_tags);size(1024)"`
}
// AddScalingGroup add ScalingGroup
func AddScalingGroup(scalingGroup *ScalingGroup) error {
	if scalingGroup == nil {
		return errors.New("func AddScalingGroup has invalid args")
	}

	scalingGroup.IsDeleted = notDeletedFlag
	scalingGroup.State = ScalingGroupStateCreating
	scalingGroup.IsInvisible = invisibleFlag
	_, err := ormer.Insert(scalingGroup)
	if err != nil {
		return errors.Wrapf(err, "orm insert scaling group[%s] err", scalingGroup.Id)
	}
	return nil
}

// DeleteScalingGroup delete ScalingGroup
func DeleteScalingGroup(txOrm orm.TxOrmer, groupId string) error {
	// 1. 查询ScalingGroup，若不存在，不返回err
	var group ScalingGroup
	qs := txOrm.QueryTable(tableNameScalingGroup).Filter(fieldNameId, groupId).
		Filter(fieldNameIsDeleted, notDeletedFlag).RelatedSel()
	err := qs.One(&group)
	if err != nil {
		if errors.Is(err, orm.ErrNoRows) {
			return nil
		}
		return errors.Wrapf(err, "read scaling group[%s] err", groupId)
	}

	// 2. 删除ScalingGroup
	delAt := time.Now().UTC()
	_, err = qs.Update(orm.Params{
		fieldNameState:       ScalingGroupStateDeleted,
		fieldNameIsDeleted:   deletedFlag,
		fieldNameIsInvisible: invisibleFlag,
		fieldNameDeleteAt:    delAt})
	if err != nil {
		return errors.Wrapf(err, "delete scaling group[%s] err", groupId)
	}

	// 3. 删除对应的InstanceConfiguration
	if group.InstanceConfiguration == nil {
		return nil
	}
	_, err = txOrm.QueryTable(tableNameInstanceConfiguration).
		Filter(fieldNameId, group.InstanceConfiguration.Id).
		Update(orm.Params{
			fieldNameIsDeleted: deletedFlag,
			fieldNameDeleteAt:  delAt})
	if err != nil {
		return errors.Wrapf(err, "delete scaling group[%s] instance config err", groupId)
	}

	// 4. 删除对应的as伸缩组信息
	return DeleteVmScalingGroup(txOrm, group.ResourceId)
}

// UpdateScalingGroup update ScalingGroup
func UpdateScalingGroup(new *ScalingGroup, cols ...string) error {
	var err error

	if _, err = ormer.Update(new, cols...); err != nil {
		return errors.Wrapf(err, "update scaling group[%s] err", new.Id)
	}

	return nil
}

// UpdateScalingGroupInstanceTags
func UpdateScalingGroupInstanceTags(groupId string, instanceTags string) error {
	return UpdateScalingGroup(&ScalingGroup{
		Id:    groupId,
		InstanceTags: instanceTags,
	}, fleldNameInstanceTags)
}
 

// UpdateScalingGroupState update ScalingGroup state
func UpdateScalingGroupState(groupId string, state string) error {
	return UpdateScalingGroup(&ScalingGroup{
		Id:    groupId,
		State: state,
	}, fieldNameState)
}

// UpdateScalingGroupVisibleState ...
func UpdateScalingGroupVisibleState(groupId string, state string) error {
	return UpdateScalingGroup(&ScalingGroup{
		Id:          groupId,
		State:       state,
		IsInvisible: visibleFlag,
	}, fieldNameState, fieldNameIsInvisible)
}

func getScalingGroupById(groupId string, f Filters) (*ScalingGroup, error) {
	if len(groupId) == 0 {
		return nil, errors.New("func getScalingGroupById has invalid args")
	}
	f[fieldNameId] = groupId

	var group ScalingGroup
	err := f.Filter(tableNameScalingGroup).RelatedSel().One(&group)
	if err != nil {
		return nil, errors.Wrapf(err, "read scaling group[%s] err", groupId)
	}

	// 关联未删除的 Scaling Policies
	if _, err = ormer.LoadRelated(&group, "ScalingPolicies"); err != nil {
		return nil, errors.Wrapf(err, "list group[%s] scaling policies err", groupId)
	}
	policiesNotDeleted := make([]*ScalingPolicy, 0, len(group.ScalingPolicies))
	for _, policy := range group.ScalingPolicies {
		if policy.IsDeleted == deletedFlag {
			continue
		}
		policiesNotDeleted = append(policiesNotDeleted, policy)
	}
	group.ScalingPolicies = policiesNotDeleted

	return &group, nil
}

// GetScalingGroupById get ScalingGroup that exclude all externally invisible states
func GetScalingGroupById(projectId, groupId string) (*ScalingGroup, error) {
	f := Filters{
		fieldNameId:          groupId,
		fieldNameIsInvisible: visibleFlag,
	}
	if len(projectId) != 0 {
		f[fieldNameProjectId] = projectId
	}
	return getScalingGroupById(groupId, f)
}

// GetNotDeletedGroupById ...
func GetNotDeletedGroupById(projectId, groupId string) (*ScalingGroup, error) {
	f := Filters{
		fieldNameId:        groupId,
		fieldNameIsDeleted: notDeletedFlag,
	}
	if len(projectId) != 0 {
		f[fieldNameProjectId] = projectId
	}
	return getScalingGroupById(groupId, f)
}

// IsScalingGroupExist check whether the ScalingGroup exists
func IsScalingGroupExist(projectId, groupId string) bool {
	return ormer.QueryTable(tableNameScalingGroup).Filter(fieldNameIsInvisible, visibleFlag).
		Filter(fieldNameProjectId, projectId).Filter(fieldNameId, groupId).Exist()
}

// IsScalingGroupExistByName check whether the ScalingGroup exists by scaling group name
func IsScalingGroupExistByName(projectId, name string) bool {
	return ormer.QueryTable(tableNameScalingGroup).Filter(fieldNameIsInvisible, visibleFlag).
		Filter(fieldNameProjectId, projectId).Filter(fieldNameName, name).Exist()
}

// ListScalingGroupByFilter list ScalingGroup by filter name
func ListScalingGroupByFilter(projectId, name string, limit, offset int) ([]*ScalingGroup, error) {
	var list []*ScalingGroup
	q := ormer.QueryTable(tableNameScalingGroup).Filter(fieldNameIsInvisible, visibleFlag).
		Filter(fieldNameProjectId, projectId)
	if len(name) != 0 {
		q = q.Filter(fieldNameName, name)
	}
	_, err := q.Limit(limit, offset).All(&list)
	return list, err
}

// UpdateAutoScalingTimestamp update auto scaling timestamp of scaling group
func UpdateAutoScalingTimestamp(groupId string) error {
	return UpdateScalingGroup(&ScalingGroup{
		Id:                   groupId,
		AutoScalingTimestamp: time.Now().UnixNano(),
	}, fieldNameScaleTime)
}

// IsScalingGroupExistWithoutProject ...
func IsScalingGroupExistWithoutProject(groupId string) bool {
	return ormer.QueryTable(tableNameScalingGroup).Filter(fieldNameIsInvisible, visibleFlag).
		Filter(fieldNameId, groupId).Exist()
}
