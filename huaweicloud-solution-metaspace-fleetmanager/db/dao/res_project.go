// Copyright (c) Huawei Technologies Co., Ltd. 2022-2022. All rights reserved.

// 资源账号项目数据表定义
package dao

import (
	"encoding/json"
	"fleetmanager/api/model/user"
	"fmt"
	"time"

	"github.com/beego/beego/v2/client/orm"
)

type ResProject struct {
	Id              string    `orm:"column(id);size(64);pk" json:"id"`
	OriginDomainId  string    `orm:"column(origin_domain_id);size(64)" json:"origin_domain_id"`
	ResDomainId     string    `orm:"column(res_domain_id);size(64)" json:"res_domain_id"`
	OriginProjectId string    `orm:"column(origin_project_id);size(64)" json:"origin_project_id"`
	ResProjectId    string    `orm:"column(res_project_id);size(64)" json:"res_project_id"`
	Region          string    `orm:"column(region);size(64)" json:"region"`
	CreationTime    time.Time `orm:"column(creation_time);type(datetime);auto_now_add" json:"creation_time"`
}

func BuildResProject(resConf *user.UserResourceConfig) *ResProject {
	resProject := ResProject{
		Id:              resConf.Id,
		OriginDomainId:  resConf.OriginDomainId,
		ResDomainId:     resConf.ResDomainId,
		OriginProjectId: resConf.OriginProjectId,
		ResProjectId:    resConf.ResProjectId,
		Region:          resConf.Region,
	}
	return &resProject
}

// GetResProjectByProjectId 获取资源租户项目信息
func GetResProjectByProjectId(originProjectId string, region string) (*ResProject, error) {
	var p ResProject
	err := Filters{"OriginProjectId": originProjectId, "Region": region}.Filter(ResProjectTable).One(&p)
	return &p, err
}

func GetResProjectById(id string) (*ResProject, error) {
	var a ResProject
	err := Filters{"Id": id}.Filter(ResProjectTable).One(&a)
	return &a, err
}

func InsertResProjectByJson(to orm.TxOrmer, id string, resConf *user.UserResourceConfig) error {
	r := *BuildResProject(resConf)
	checkExist := ResProject{}
	_ = Filters{"OriginProjectId": r.OriginProjectId, "OriginDomainId": r.OriginDomainId}.
		Filter(ResProjectTable).One(&checkExist)
	if checkExist != (ResProject{}) {
		return fmt.Errorf("projectId %s exist in %s", r.OriginProjectId, ResProjectTable)
	}
	_, err := to.Insert(&r)
	if err != nil {
		_ = to.Rollback()
		return err
	}
	return nil
}

func UpdateResProjectByJson(to orm.TxOrmer, request []byte, indexId string) error {
	newConfig := ResProject{}
	if err := json.Unmarshal(request, &newConfig); err != nil {
		return err
	}
	_, err := to.Update(&newConfig, "OriginDomainId", "OriginProjectId")
	if err != nil {
		_ = to.Rollback()
		return err
	}
	return nil
}

func DeleteResProjectById(to orm.TxOrmer, indexId string) error {
	resproject := ResProject{}
	err := Filters{"Id": indexId}.Filter(ResProjectTable).One(&resproject)
	if err != nil {
		return err
	}
	_, err = to.Delete(&resproject)
	if err != nil {
		_ = to.Rollback()
		return fmt.Errorf("delete project err:%s", err.Error())
	}
	return nil
}

func DeleteResProject(originProjectId string, region string) error {
	_, err := Filters{"OriginProjectId": originProjectId, "Region": region}.Filter(ResProjectTable).Delete()
	return err
}
