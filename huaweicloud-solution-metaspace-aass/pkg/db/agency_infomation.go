// Copyright (c) Huawei Technologies Co., Ltd. 2022-2022. All rights reserved.

// 委托数据表定义
package db

import (
	"github.com/pkg/errors"
)

type AgencyInfo struct {
	ProjectId  string `orm:"column(project_id);size(64);pk" json:"project_id"`
	AgencyName string `orm:"column(agency_name);size(64)" json:"agency_name"`
	DomainId   string `orm:"column(domain_id);size(64)" json:"domain_id"`
	TimeModel
}

// AddOrUpdateAgencyInfo add or update AgencyInfo
func AddOrUpdateAgencyInfo(info *AgencyInfo) error {
	if info == nil || info.ProjectId == "" {
		return errors.Errorf("func AddOrUpdateAgencyInfo[%+v] has invalid args", info)
	}

	info.IsDeleted = notDeletedFlag
	_, err := ormer.InsertOrUpdate(info)
	if err != nil {
		return errors.Wrapf(err, "add or update agency info[%s] err", info.ProjectId)
	}
	return nil
}

// GetAgencyInfo get AgencyInfo
func GetAgencyInfo(projectId string) (*AgencyInfo, error) {
	g := &AgencyInfo{ProjectId: projectId}
	if err := ormer.Read(g); err != nil {
		return nil, errors.Wrapf(err, "read agency info[project_id: %s] from db err", projectId)
	}
	return g, nil
}
