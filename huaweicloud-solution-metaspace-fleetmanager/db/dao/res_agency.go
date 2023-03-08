// Copyright (c) Huawei Technologies Co., Ltd. 2022-2022. All rights reserved.

// 资源账号委托数据表定义
package dao

import (
	"encoding/json"
	"fleetmanager/api/model/user"
	"fmt"
	"time"

	"github.com/beego/beego/v2/client/orm"
)

type ResAgency struct {
	Id             string    `orm:"column(id);size(64);pk" json:"id"`
	OriginDomainId string    `orm:"column(origin_domain_id);size(64)" json:"origin_domain_id"`
	ResDomainId    string    `orm:"column(res_domain_id);size(64)" json:"res_domain_id"`
	ResUserId      string    `orm:"column(res_user_id);size(64)" json:"res_user_id"`
	AgencyName     string    `orm:"column(agency_name);" json:"agency_name"`
	IamAgencyName  string    `orm:"column(iam_agency_name);" json:"iam_agency_name"`
	CreationTime   time.Time `orm:"column(creation_time);type(datetime);auto_now_add" json:"creation_time"`
	Region         string    `orm:"column(region);size(64)" json:"region"`
}

func BuildResAgency(resConf *user.UserResourceConfig) *ResAgency {
	resAgency := ResAgency{
		Id:             resConf.Id,
		OriginDomainId: resConf.OriginDomainId,
		ResDomainId:    resConf.ResDomainId,
		ResUserId:      resConf.ResUserId,
		AgencyName:     resConf.AgencyName,
		IamAgencyName:  resConf.IamAgencyName,
		Region:         resConf.Region,
	}
	return &resAgency
}

// GetResAgencyByDomainId 获取资源租户委托信息
func GetResAgencyByDomainId(originDomainId string, region string) (*ResAgency, error) {
	var a ResAgency
	err := Filters{"OriginDomainId": originDomainId, "Region": region}.Filter(ResAgencyTable).One(&a)
	return &a, err
}

func GetResAgencyById(id string) (*ResAgency, error) {
	var a ResAgency
	err := Filters{"Id": id}.Filter(ResAgencyTable).One(&a)
	return &a, err
}

func InsertResAgencyByJson(to orm.TxOrmer, id string, resConf *user.UserResourceConfig) error {
	r := *BuildResAgency(resConf)
	checkExist := ResAgency{}
	_ = Filters{"OriginDomainId": r.OriginDomainId, "Region": r.Region}.Filter(ResAgencyTable).One(&checkExist)
	if checkExist != (ResAgency{}) {
		return fmt.Errorf("originDomainId %s exist in %s", r.OriginDomainId, ResAgencyTable)
	}
	_, err := to.Insert(&r)
	if err != nil {
		_ = to.Rollback()
		return err
	}
	return nil
}

// 根据origin domain id 和region查询变更字段
func UpdateResAgencyByJson(to orm.TxOrmer, request []byte, indexId string) error {
	newConfig := ResAgency{}
	if err := json.Unmarshal(request, &newConfig); err != nil {
		return err
	}
	_, err := to.Update(&newConfig, "OriginDomainId")
	if err != nil {
		_ = to.Rollback()
		return err
	}
	return nil
}

func DeleteResAgencyById(to orm.TxOrmer, indexId string) error {
	resagency := ResAgency{}
	err := Filters{"Id": indexId}.Filter(ResAgencyTable).One(&resagency)
	if err != nil {
		return err
	}
	_, err = to.Delete(&resagency)
	if err != nil {
		_ = to.Rollback()
		return fmt.Errorf("delete agency err:%s", err.Error())
	}
	return nil
}

func DeleteResAgency(originDomainId string, region string) error {
	_, err := Filters{"OriginDomainId": originDomainId, "Region": region}.Filter(ResAgencyTable).Delete()
	return err
}
