// Copyright (c) Huawei Technologies Co., Ltd. 2022-2022. All rights reserved.

// 实例配置数据表
package db

import (
	"github.com/beego/beego/v2/client/orm"
	"github.com/pkg/errors"
)

const (
	tableNameInstanceConfiguration = "instance_configuration"
)

type InstanceConfiguration struct {
	ScalingGroup *ScalingGroup `orm:"reverse(one)"`
	Id           string        `orm:"column(id);size(128);pk"`
	// 运行时配置
	RuntimeConfiguration                    string `orm:"column(runtime_configuration);type(text)"`
	ServerSessionProtectionPolicy           string `orm:"column(server_session_protection_policy);sie(64)"`
	ServerSessionProtectionTimeLimitMinutes int32  `orm:"column(server_session_protection_time_limit_minutes);type(int);default(5)"`
	MaxServerSession                        int32  `orm:"column(max_server_session);type(int);default(1)"`
	TimeModel
}

// AddInstanceConfiguration add InstanceConfiguration
func AddInstanceConfiguration(config *InstanceConfiguration) error {
	if config == nil {
		return errors.New("func AddInstanceConfiguration has invalid args")
	}

	config.IsDeleted = notDeletedFlag
	_, err := ormer.Insert(config)
	if err != nil {
		return errors.Wrapf(err, "orm insert instance configuration[%s] err", config.Id)
	}
	return nil
}

// DeleteInstanceConfiguration delete InstanceConfiguration
func DeleteInstanceConfiguration(configId string) error {
	_, err := ormer.Delete(&InstanceConfiguration{Id: configId})
	if err != nil {
		return errors.Wrapf(err, "delete instance configuration[%s] err", configId)
	}
	return nil
}

// GetInstanceConfigurationById get InstanceConfiguration
func GetInstanceConfigurationById(configId string) (*InstanceConfiguration, error) {
	o := orm.NewOrm()
	g := &InstanceConfiguration{Id: configId}
	if err := o.Read(g); err != nil {
		return nil, errors.Wrapf(err, "read instance configuration[%s] err", configId)
	}
	return g, nil
}

// UpdateInstanceConfiguration update InstanceConfiguration
func UpdateInstanceConfiguration(config *InstanceConfiguration) error {
	var err error

	o := orm.NewOrm()
	if _, err = o.Update(config); err != nil {
		return errors.Wrapf(err, "update instance configuration[%s] err", config.Id)
	}

	return nil
}
