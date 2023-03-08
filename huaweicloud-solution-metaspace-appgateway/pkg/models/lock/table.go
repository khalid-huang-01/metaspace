// Copyright (c) Huawei Technologies Co., Ltd. 2022-2022. All rights reserved.

// 锁表
package lock

import (
	"time"

	"github.com/beego/beego/v2/client/orm"
)

const (
	TableNameLock      = "DISTRIBUTED_LOCK"
	FieldNameExpiredAt = "EXPIRED_AT"
	FieldNameName      = "NAME"
	FieldNameCategory  = "CATEGORY"
)

type Lock struct {
	IDInc     int32     `orm:" pk; auto; column(ID_INC); default(0);"`
	CreatedAt time.Time `orm:" column(CREATED_AT); type(datetime);auto_now_add"`
	UpdatedAt time.Time `orm:" column(UPDATED_AT); type(datetime);auto_now"`
	ExpiredAt time.Time `orm:" column(EXPIRED_AT); type(datetime)"`
	Name      string    `orm:" column(NAME); size(255); null"`
	Holder    string    `orm:" column(HOLDER); size(255); null"`
	Category  string    `orm:" column(CATEGORY); size(255); null"`
}

func init() {
	orm.RegisterModel(new(Lock))
}

func (l *Lock) TableName() string {
	return TableNameLock
}

func (l *Lock) TableUnique() [][]string {
	return [][]string{
		{FieldNameName},
	}
}
