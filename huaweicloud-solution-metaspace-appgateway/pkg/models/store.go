// Copyright (c) Huawei Technologies Co., Ltd. 2022-2022. All rights reserved.

// 数据库初始化
package models

import (
	"fmt"
	"time"

	"github.com/beego/beego/v2/client/orm"
	_ "github.com/go-sql-driver/mysql"

	"codehub-g.huawei.com/videocloud/mediaprocesscenter/application-gateway/config"
	"codehub-g.huawei.com/videocloud/mediaprocesscenter/application-gateway/pkg/common"
	"codehub-g.huawei.com/videocloud/mediaprocesscenter/application-gateway/pkg/utils/log"
	"codehub-g.huawei.com/videocloud/mediaprocesscenter/application-gateway/pkg/utils/security"
)

var MySqlOrm orm.Ormer

// RegisterDB register db
func RegisterDB(username, address, dbName string, password []byte) error {
	driverName := common.DbDriverName
	dataSource := fmt.Sprintf("%s:%s@tcp(%s)/%s?charset=utf8&loc=Local", username, password, address, dbName)

	err := orm.RegisterDriver(driverName, orm.DRMySQL)
	if err != nil {
		log.RunLogger.Errorf("[store] failed to register db driver for %v", err)
		return err
	}

	err = orm.RegisterDataBase("default", driverName, dataSource)
	if err != nil {
		log.RunLogger.Errorf("[store] failed to register data base for %v", err)
		return err
	}

	// set UTC
	orm.DefaultTimeLoc = time.UTC

	return nil
}

// SyncAndUseDB sync and use db
func SyncAndUseDB() error {
	err := orm.RunSyncdb("default", false, true)
	if err != nil {
		log.RunLogger.Errorf("[store] create tables failed, error: %v", err)
		return err
	}

	MySqlOrm = orm.NewOrmUsingDB("default")

	return nil
}

// InitDB init db
func InitDB() error {
	plainPw, err := security.GCM_Decrypt(config.GlobalConfig.DbPassword, 
		config.GlobalConfig.GCMKey, config.GlobalConfig.GCMNonce)

	if err != nil {
		log.RunLogger.Errorf("[store] failed to decrypt password for %v", err)
		return err
	}

	err = RegisterDB(config.GlobalConfig.DbUserName, config.GlobalConfig.DbAddr,
		config.GlobalConfig.DbName, []byte(plainPw))
	if err != nil {
		log.RunLogger.Errorf("[store] failed to register db for %v", err)
		return err
	}

	err = SyncAndUseDB()
	if err != nil {
		log.RunLogger.Errorf("[store] failed to sync and use db for %v", err)
		return err
	}

	return nil
}
