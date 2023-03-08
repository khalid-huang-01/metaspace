// Copyright (c) Huawei Technologies Co., Ltd. 2022-2022. All rights reserved.

// 数据表通用操作
package dao

import (
	"fleetmanager/db/dbm"
	"github.com/beego/beego/v2/client/orm"
)

type Filters map[string]interface{}

// Filter filter过滤器 NOTE(nkpatx):这个Filter是个坑, 非原子性操作
func (f Filters) Filter(table string) orm.QuerySeter {
	qs := dbm.Ormer.QueryTable(table).Limit(-1)
	for k, v := range f {
		qs = qs.Filter(k, v)
	}
	return qs
}

func (f Filters) Condition() *orm.Condition {
	cond := orm.NewCondition()
	for k, v := range f {
		cond = cond.And(k, v)
	}
	return cond
}