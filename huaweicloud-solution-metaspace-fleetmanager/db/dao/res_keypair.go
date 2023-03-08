// Copyright (c) Huawei Technologies Co., Ltd. 2022-2022. All rights reserved.

// 资源账号密匙对数据表
package dao

import (
	"encoding/json"
	"fleetmanager/api/model/user"
	"fmt"
	"time"

	"github.com/beego/beego/v2/client/orm"
)

type ResKeypair struct {
	Id             string    `orm:"column(id);size(64);pk" json:"id"`
	OriginDomainId string    `orm:"column(origin_domain_id);size(64)" json:"origin_domain_id"`
	ResDomainId    string    `orm:"column(res_domain_id);size(64)" json:"res_domain_id"`
	KeypairName    string    `orm:"column(keypair_name);" json:"keypair_name"`
	KeypairData    string    `orm:"column(keypair_data);" json:"keypair_data"`
	CreationTime   time.Time `orm:"column(creation_time);type(datetime);auto_now_add" json:"creation_time"`
	Region         string    `orm:"column(region);size(64)" json:"region"`
	UpdateTime     time.Time `orm:"column(update_time);type(datetime);auto_now" json:"update_time"`
}

func BuildResKeypair(resConf *user.UserResourceConfig) *ResKeypair {
	resKeypair := ResKeypair{
		Id:             resConf.Id,
		OriginDomainId: resConf.OriginDomainId,
		ResDomainId:    resConf.ResDomainId,
		KeypairName:    resConf.KeypairName,
		KeypairData:    resConf.KeypairData,
		Region:         resConf.Region,
	}
	return &resKeypair
}

// GetResKeypairByDomainId 获取资源租户委托keypair
func GetResKeypairByDomainId(originDomainId string, region string) (*ResKeypair, error) {
	var k ResKeypair
	err := Filters{"OriginDomainId": originDomainId, "Region": region}.Filter(ResKeypairTable).One(&k)
	return &k, err
}

func GetResKeypairById(id string) (*ResKeypair, error) {
	var a ResKeypair
	err := Filters{"Id": id}.Filter(ResKeypairTable).One(&a)
	return &a, err
}

func InsertResKeypairByJson(to orm.TxOrmer, id string, resConf *user.UserResourceConfig) error {
	r := *BuildResKeypair(resConf)
	checkExist := ResKeypair{}
	_ = Filters{"OriginDomainId": r.OriginDomainId, "Region": r.Region}.Filter(ResKeypairTable).One(&checkExist)
	if checkExist != (ResKeypair{}) {
		return fmt.Errorf("originDomainId %s exist in %s", r.OriginDomainId, ResKeypairTable)
	}
	_, err := to.Insert(&r)
	if err != nil {
		_ = to.Rollback()
		return err
	}
	return nil
}

func UpdateResKeypairByJson(to orm.TxOrmer, request []byte, indexId string) error {
	newConfig := ResKeypair{}
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

func DeleteResKeypairById(to orm.TxOrmer, indexId string) error {
	reskeypair := ResKeypair{}
	err := Filters{"Id": indexId}.Filter(ResKeypairTable).One(&reskeypair)
	if err != nil {
		return err
	}
	_, err = to.Delete(&reskeypair)
	if err != nil {
		_ = to.Rollback()
		return fmt.Errorf("delete keypair err:%s", err.Error())
	}
	return nil
}

func DeleteResKeypair(originDomainId string, region string) error {
	//这种删除方式会删除所有检索结果
	_, err := Filters{"OriginDomainId": originDomainId, "Region": region}.Filter(ResKeypairTable).Delete()
	return err
}
