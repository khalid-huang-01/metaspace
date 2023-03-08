// Copyright (c) Huawei Technologies Co., Ltd. 2022-2022. All rights reserved.

// 数据库初始化
package db

import (
	"fmt"

	"github.com/beego/beego/v2/client/orm"
	_ "github.com/go-sql-driver/mysql"
	"github.com/pkg/errors"

	"scase.io/application-auto-scaling-service/pkg/setting"
)

const (
	databaseDriveName = "mysql"
)

// ormer is ormer object interface for all transaction processing and switching database
var ormer orm.Ormer

// Init init mysql
func Init() error {
	var (
		err error
		ds  string
	)

	// register db
	if err = orm.RegisterDriver("mysql", orm.DRMySQL); err != nil {
		return errors.Wrap(err, "register db driver failed")
	}
	ds = getDataSource()
	if err = orm.RegisterDataBase("default", databaseDriveName, ds); err != nil {
		return errors.Wrap(err, "register db failed")
	}

	// register model
	orm.RegisterModel(
		new(WorkNode),
		new(InstanceConfiguration),
		new(ScalingGroup),
		new(ScalingPolicy),
		new(VmScalingGroup),
		new(AgencyInfo),
		new(DeletingVm),
		new(AsyncTask),
		new(MetricMonitorTask),
		new(LtsConfig),
		new(LogTransfer),
	)

	// create orm
	ormer = orm.NewOrm()

	// create table
	if err = orm.RunSyncdb("default", false, true); err != nil {
		return err
	}

	{ // === debug
		orm.Debug = true
	}

	return nil
}

func getDataSource() string {
	return fmt.Sprintf("%s:%s@tcp(%s)/%s?charset=%s",
		setting.MysqlUser, setting.MysqlPassword, setting.MysqlAddress, setting.MysqlDbName, setting.MysqlCharset)
}
