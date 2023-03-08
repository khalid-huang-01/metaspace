package db

import (
	"fmt"
	"time"
)

const tableNameLtsConfig = "lts_config"

type LtsConfig struct {
	Id               string    `orm:"column(id);size(64);pk"`
	FleetId          string    `orm:"column(fleet_id);size(64);"`
	ASGroupId        string    `orm:"column(as_group_id);size(64)"`
	ProjectId        string    `orm:"column(project_id);size(64)"`
	LogGroupId       string    `orm:"column(log_group_id);size(64)"`
	LogGroupName     string    `orm:"column(log_group_name);size(64)"`
	LogStreamId      string    `orm:"column(log_stream_id);size(64)"`
	LogStreamName    string    `orm:"column(log_stream_name);size(64)"`
	HostGroupID      string    `orm:"column(host_group_id);size(64)"`
	HostGroupName    string    `orm:"column(host_group_name);size(64)"`
	LogConfigPath    string    `orm:"column(log_config_path);size(64)"`
	AccessConfigName string    `orm:"column(access_config_name);size(64)"`
	AccessConfigId   string    `orm:"column(access_config_id);size(64)"`
	CreateTime       time.Time `orm:"column(create_time);auto_now_add;type(datetime)"`
	ObsTransferPath  string    `orm:"column(obs_transfer_path);size(256)"`
	IsDeleted        bool      `orm:"column(is_deleted)"`
	Description      string    `orm:"column(description);size(128)"`
	EnterpriseProjectId	string	`orm:"column(enterprise_project_id);size(128)"`
}

type ltsConfigTable struct{}

var ltstable = ltsConfigTable{}

func LtsConfigTable() *ltsConfigTable {
	return &ltstable
}

func (s *ltsConfigTable) Insert(ltsConfig *LtsConfig) error {
	_, err := ormer.Insert(ltsConfig)
	return err
}

func (s *ltsConfigTable) Get(f Filters) (LtsConfig, error) {
	var ltsconfig LtsConfig
	err := f.Filter(tableNameLtsConfig).One(&ltsconfig)
	return ltsconfig, err
}

func (s *ltsConfigTable) Update(lts *LtsConfig, cols ...string) error {
	_, err := ormer.Update(lts, cols...)
	return err
}

func (s *ltsConfigTable) Delete(lts *LtsConfig) error {
	_, err := ormer.Delete(lts)
	return err
}

func (s *ltsConfigTable) List() ([]LtsConfig, error) {
	var list []LtsConfig
	qs := ormer.QueryTable(tableNameLtsConfig)
	_, err := qs.All(&list, "id", "fleet_id", "as_group_id", "project_id", "host_group_id")
	if err != nil {
		return nil, err
	}
	return list, nil
}

func (s *ltsConfigTable) CheckExist(FleetId, LogStreamName, HostGroupName, AccessConfigName string) (bool, error) {
	checkFleetId := Filters{"FleetId": FleetId}.Filter(tableNameLtsConfig).Exist()
	if checkFleetId {
		return false, fmt.Errorf("FleetId exist in DB")
	}
	checkLogStreamName := Filters{"FleetId": LogStreamName}.Filter(tableNameLtsConfig).Exist()
	if checkLogStreamName {
		return false, fmt.Errorf("LogStreamName exist in DB")
	}
	checkHostGroupName := Filters{"HostGroupName": HostGroupName}.Filter(tableNameLtsConfig).Exist()
	if checkHostGroupName {
		return false, fmt.Errorf("HostGroupName exist in DB")
	}
	checkAccessConfigName := Filters{"AccessConfigName": AccessConfigName}.Filter(tableNameLtsConfig).Exist()
	if checkAccessConfigName {
		return false, fmt.Errorf("AccessConfigName exist in DB")
	}
	return true, nil
}

func (s *ltsConfigTable) ListbyProject(offset int, limit int, projectId string) ([]LtsConfig, error) {
	var list []LtsConfig
	qs := ormer.QueryTable(tableNameLtsConfig)
	_, err := qs.Filter("ProjectId", projectId).Offset(offset).Limit(limit).All(&list)
	if err != nil {
		return nil, err
	}
	return list, nil
}

func (s *ltsConfigTable) Total(projectId string) (int, error) {
	qs := ormer.QueryTable(tableNameLtsConfig)
	total, err := qs.Filter("ProjectId", projectId).Count()
	if err != nil {
		return 0, err
	}
	return int(total), nil
}
