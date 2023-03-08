// Copyright (c) Huawei Technologies Co., Ltd. 2022-2022. All rights reserved.

// 数据库初始化
package db

import (
	"fleetmanager/db/dao"
	"fleetmanager/db/dbm"
)

// Init 数据库模块初始化函数
func Init() error {
	dao.Init()
	if err := dbm.Init(); err != nil {
		return err
	}
	// 初始化用户
	if err := dao.InitUserTable(); err != nil {
		return err
	}
	return nil
}
