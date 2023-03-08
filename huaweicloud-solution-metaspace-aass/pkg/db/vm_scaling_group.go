// Copyright (c) Huawei Technologies Co., Ltd. 2022-2022. All rights reserved.

// VM伸缩组定义
package db

import (
	"time"

	"github.com/beego/beego/v2/client/orm"
	"github.com/pkg/errors"
)

const (
	tableNameVmScalingGroup = "vm_scaling_group"
)

type VmScalingGroup struct {
	// Id: 与ScalingGroup.ResourceId对应
	Id string `orm:"column(id);size(128);pk"`
	// ScalingConfigId：AS伸缩组配置Id
	ScalingConfigId string `orm:"column(scaling_config_id);size(64)"`
	// AsGroupId：AS伸缩组Id
	AsGroupId string `orm:"column(as_group_id);size(64)"`
	TimeModel
}

// AddVmScalingGroup add VmScalingGroup
func AddVmScalingGroup(group *VmScalingGroup) error {
	if group == nil {
		return errors.New("func AddVmScalingGroup param can not be nil")
	}

	group.IsDeleted = notDeletedFlag
	_, err := ormer.Insert(group)
	if err != nil {
		return errors.Wrapf(err, "orm insert vm scaling group[%s] err", group.Id)
	}
	return nil
}

// DeleteVmScalingGroup delete VmScalingGroup
func DeleteVmScalingGroup(txOrm orm.TxOrmer, vmGroupId string) error {
	var vmGroup VmScalingGroup
	err := txOrm.QueryTable(tableNameVmScalingGroup).Filter(fieldNameId, vmGroupId).
		Filter(fieldNameIsDeleted, notDeletedFlag).One(&vmGroup)
	if err != nil {
		if errors.Is(err, orm.ErrNoRows) {
			return nil
		}
		return errors.Wrapf(err, "read vm scaling group[%s] err", vmGroupId)
	}

	vmGroup.IsDeleted = deletedFlag
	vmGroup.DeleteAt = time.Now().UTC()
	_, err = txOrm.Update(&vmGroup)
	if err != nil {
		return errors.Wrapf(err, "delete vm scaling group[%s] err", vmGroupId)
	}
	return nil
}

// GetVmScalingGroupById get VmScalingGroup detail
func GetVmScalingGroupById(groupId string) (*VmScalingGroup, error) {
	group := &VmScalingGroup{Id: groupId}
	group.IsDeleted = notDeletedFlag
	if err := ormer.Read(group); err != nil {
		return nil, errors.Wrapf(err, "read vm scaling group[%s] err", groupId)
	}
	return group, nil
}

// UpdateScalingGroup update ScalingGroup
func UpdateVmScalingGroup(new *VmScalingGroup, cols ...string) error {
	var err error

	if _, err = ormer.Update(new, cols...); err != nil {
		return errors.Wrapf(err, "update vm scaling group[%s] err", new.Id)
	}

	return nil
}

// UpdateAsGroupIdOfVmScalingGroup ...
func UpdateAsGroupIdOfVmScalingGroup(id, asGroupId string) error {
	group := &VmScalingGroup{
		Id:        id,
		AsGroupId: asGroupId,
	}
	return UpdateVmScalingGroup(group, fieldNameAsGroupId)
}

// UpdateAsConfigIdOfVmScalingGroup ...
func UpdateAsConfigIdOfVmScalingGroup(id, asConfigId string) error {
	group := &VmScalingGroup{
		Id:              id,
		ScalingConfigId: asConfigId,
	}
	return UpdateVmScalingGroup(group, fieldNameAsConfigId)
}
