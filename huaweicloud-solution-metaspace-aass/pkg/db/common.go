// Copyright (c) Huawei Technologies Co., Ltd. 2022-2022. All rights reserved.

// 通用字段定义
package db

import (
	"time"

	"github.com/beego/beego/v2/client/orm"
)

const (
	fieldNameId              = "id"
	fieldNameName            = "name"
	fieldNameProjectId       = "project_id"
	fieldNameState           = "state"
	fieldNameIsDeleted       = "is_deleted"
	fieldNameDeleteAt        = "delete_at"
	fieldNameUpdateAt        = "update_at"
	fieldNameScalingGroupId  = "scaling_group_id"
	fieldNameScalingPolicyId = "scaling_policy_id"
	fieldNamePolicyType      = "policy_type"
	fieldNameScaleTime       = "auto_scaling_timestamp"
	fieldNameAsGroupId       = "as_group_id"
	fieldNameAsConfigId      = "scaling_config_id"
	fieldNameVmId            = "vm_id"
	fieldNameWorkNodeId      = "work_node_id"
	fieldNameTakeOverId      = "take_over_id"
	fieldNameTaskType        = "task_type"
	fieldNameTaskKey         = "task_key"
	fieldNameIsInvisible     = "is_invisible"
	fieldNameTargetValue     = "target_value"
	fleldNameInstanceTags	 = "instance_tags"

	fieldNameStateIn    = "state__in"
	fieldNameIdIn       = "id__in"
	fieldNameUpdateAtLt = "update_at__lt"

	notDeletedFlag = "0"
	deletedFlag    = "1"
	visibleFlag    = "0"
	invisibleFlag  = "1"
)

// TimeModel record the time of creating, updating, deleting
type TimeModel struct {
	CreateAt  time.Time `orm:"column(create_at);type(datetime);auto_now_add" json:"-,omitempty"`
	UpdateAt  time.Time `orm:"column(update_at);type(datetime);auto_now" json:"-"`
	DeleteAt  time.Time `orm:"null;column(delete_at);type(datetime)" json:"-"`
	IsDeleted string    `orm:"column(is_deleted);default(0)" json:"-"`
}

type Filters map[string]interface{}

// Filter filter过滤器 NOTE(nkpatx):这个Filter是个坑, 非原子性操作
func (f Filters) Filter(table string) orm.QuerySeter {
	qs := ormer.QueryTable(table)
	for k, v := range f {
		qs = qs.Filter(k, v)
	}
	return qs
}
