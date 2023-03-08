// Copyright (c) Huawei Technologies Co., Ltd. 2022-2022. All rights reserved.

// 数据库管理模块初始化
package dbm

import (
	"fleetmanager/setting"
	"fmt"

	"github.com/beego/beego/v2/client/orm"
	"github.com/go-redis/redis"
	_ "github.com/go-sql-driver/mysql"
)

// Ormer is Ormer object interface for all transaction processing and switching database
var Ormer orm.Ormer

var RedisClient *redis.Client
var RedisDB int = 10

func getDataSource() (string, error) {
	address := setting.MysqlAddress
	db := setting.MysqlDBName
	user := setting.MysqlUser
	charset := setting.MysqlCharset
	password := setting.MysqlPassword
	return fmt.Sprintf("%s:%s@tcp(%s)/%s?charset=%s", user, password, address, db, charset), nil
}

// Should be called only once
func Init() error {
	if err := orm.RegisterDriver("mysql", orm.DRMySQL); err != nil {
		return err
	}

	ds, err := getDataSource()
	if err != nil {
		return err
	}
	if err = orm.RegisterDataBase("default", "mysql", ds); err != nil {
		return err
	}

	if err = orm.RunSyncdb("default", false, true); err != nil {
		return err
	}

	// create orm
	Ormer = orm.NewOrm()

	RedisClient = redis.NewClient(&redis.Options{
		Addr:     setting.RedisAddress,
		Password: setting.RedisPassword,
		DB:       RedisDB,
	})

	return nil
}
