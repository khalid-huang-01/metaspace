// Copyright (c) Huawei Technologies Co., Ltd. 2022-2022. All rights reserved.

// 删除虚机数据表定义
package db

import (
	"github.com/pkg/errors"
)

const (
	tableNameDeletingVm = "deleting_vm"
)

type DeletingVm struct {
	Id        string `orm:"column(vm_id);size(128);pk"`
	AsGroupId string `orm:"column(as_group_id);size(128)"`
	ProjectId string `orm:"column(project_id);size(64)"`
}

// AddDeletingVm add DeletingVm
// Deprecated: 临时方案使用，后续废弃
func AddDeletingVm(vm *DeletingVm) error {
	if vm == nil {
		return errors.New("func AddDeletingVm param can not be nil")
	}

	_, err := ormer.Insert(vm)
	if err != nil {
		return errors.Wrapf(err, "orm insert deleting vm[%s] err", vm.Id)
	}
	return nil
}

// DeleteDeletingVm delete deleting vm info
// Deprecated
func DeleteDeletingVm(vmId string) error {
	_, err := ormer.Delete(&DeletingVm{Id: vmId})
	if err != nil {
		return errors.Wrapf(err, "delete deleting vm info[%s] err", vmId)
	}
	return nil
}

// GetAllDeletingVms ...
// Deprecated
func GetAllDeletingVms() ([]*DeletingVm, error) {
	var vms []*DeletingVm
	_, err := ormer.QueryTable(tableNameDeletingVm).All(&vms)
	if err != nil {
		return nil, errors.Wrap(err, "get deleting vms from db err")
	}
	return vms, nil
}

// GetAsGroupDeletingVmIds 获取as伸缩组所有正在删除的实例id
// Deprecated
func GetAsGroupDeletingVmIds(asGroupId string) ([]string, error) {
	var vms []*DeletingVm
	_, err := ormer.QueryTable(tableNameDeletingVm).
		Filter(fieldNameAsGroupId, asGroupId).
		All(&vms, fieldNameVmId)
	if err != nil {
		return nil, errors.Wrapf(err, "get deleting vms for as group[%s] err", asGroupId)
	}

	ids := make([]string, 0, len(vms))
	for _, vm := range vms {
		ids = append(ids, vm.Id)
	}
	return ids, nil
}
