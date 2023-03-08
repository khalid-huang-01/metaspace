// Copyright (c) Huawei Technologies Co., Ltd. 2022-2022. All rights reserved.

// 资源账号用户名数据表
package dao

import (
	"encoding/json"
	"fleetmanager/api/model/user"
	"fmt"
	"time"

	"github.com/beego/beego/v2/client/orm"
)

type ResDomain struct {
	Id               string    `orm:"column(id);size(64);pk" json:"id"`
	OriginDomainName string    `orm:"column(origin_domain_name);" json:"origin_domain_name"`
	OriginDomainId   string    `orm:"column(origin_domain_id);size(64)" json:"origin_domain_id"`
	ResDomainId      string    `orm:"column(res_domain_id);size(64)" json:"res_domain_id"`
	ResDomainName    string    `orm:"column(res_domain_name);" json:"res_domain_name"`
	CreationTime     time.Time `orm:"column(creation_time);type(datetime);auto_now_add" json:"creation_time"`
	Region           string    `orm:"column(region);size(64)" json:"region"`
}

func BuildResDomain(resConf *user.UserResourceConfig) *ResDomain {
	resDomain := ResDomain{
		Id:               resConf.Id,
		OriginDomainName: resConf.OriginDomainName,
		OriginDomainId:   resConf.OriginDomainId,
		ResDomainId:      resConf.ResDomainId,
		ResDomainName:    resConf.ResDomainName,
		Region:           resConf.Region,
	}
	return &resDomain
}

func GetResDomainIdByDomainId(originDomainId string, region string) (*ResDomain, error) {
	var u ResDomain
	err := Filters{"OriginDomainId": originDomainId, "Region": region}.Filter(ResDomainTable).One(&u)
	return &u, err
}

func GetResDomainIdById(id string) (*ResDomain, error) {
	var a ResDomain
	err := Filters{"Id": id}.Filter(ResDomainTable).One(&a)
	return &a, err
}

func InsertResDomainByJson(to orm.TxOrmer, id string, resConf *user.UserResourceConfig) error {
	r := *BuildResDomain(resConf)
	checkExist := ResDomain{}
	_ = Filters{"OriginDomainId": r.OriginDomainId, "Region": r.Region}.Filter(ResDomainTable).One(&checkExist)
	if checkExist != (ResDomain{}) {
		return fmt.Errorf("originDomainId %s exist in %s", r.OriginDomainId, ResDomainTable)
	}
	_, err := to.Insert(&r)
	if err != nil {
		_ = to.Rollback()
		return err
	}
	return nil
}

// 根据origin domain id 和region查询变更字段
func UpdateResDomainByJson(to orm.TxOrmer, r []byte, indexId string) error {
	newConfig := ResDomain{}
	if err := json.Unmarshal(r, &newConfig); err != nil {
		return err
	}
	_, err := to.Update(&newConfig, "OriginDomainId", "OriginDomainName")
	if err != nil {
		_ = to.Rollback()
		return err
	}
	return nil
}

func DeleteResDomainById(to orm.TxOrmer, indexId string) error {
	resdomain := ResDomain{}
	err := Filters{"Id": indexId}.Filter(ResDomainTable).One(&resdomain)
	if err != nil {
		return err
	}
	_, err = to.Delete(&resdomain)
	if err != nil {
		_ = to.Rollback()
		return fmt.Errorf("delete domain err:%s", err.Error())
	}
	return nil
}

func DeleteResDomain(OriginDomainName string, region string) error {
	_, err := Filters{"originDomainName": OriginDomainName, "Region": region}.Filter(ResDomainTable).Delete()
	return err
}
