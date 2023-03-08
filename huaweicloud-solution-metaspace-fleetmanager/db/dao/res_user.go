// Copyright (c) Huawei Technologies Co., Ltd. 2022-2022. All rights reserved.

// 资源账号数据表
package dao

import (
	"encoding/json"
	"fleetmanager/api/model/user"
	"fmt"
	"time"

	"github.com/beego/beego/v2/client/orm"
)

type ResUser struct {
	Id              string    `orm:"column(id);size(64);pk" json:"id"`
	OriginDomainId  string    `orm:"column(origin_domain_id);size(64)" json:"origin_domain_id"`
	ResDomainId     string    `orm:"column(res_domain_id);size(64)" json:"res_domain_id"`
	ResUserName     string    `orm:"column(res_user_name);" json:"res_user_name"`
	ResUserId       string    `orm:"column(res_user_id);size(64)" json:"res_user_id"`
	ResUserPassword string    `orm:"column(res_user_pw);" json:"res_user_pw"`
	CreationTime    time.Time `orm:"column(creation_time);type(datetime);auto_now_add" json:"creation_time"`
	Region          string    `orm:"column(region);size(64)" json:"region"`
	UpdateTime      time.Time `orm:"column(update_time);type(datetime);auto_now" json:"update_time"`
}

func BuildResUser(resConf *user.UserResourceConfig) *ResUser {
	resUser := ResUser{
		Id: resConf.Id,
		OriginDomainId: resConf.OriginDomainId,
		ResDomainId: resConf.ResDomainId,
		ResUserName: resConf.ResUserName,
		ResUserId: resConf.ResUserId,
		ResUserPassword: resConf.ResUserPassword,
		Region: resConf.Region,
	}
	return &resUser
}

// GetResUserByProjectId 获取资源租户用户信息
func GetResUserByProjectId(originProjectId string, region string) (*ResUser, error) {
	project, err := GetResProjectByProjectId(originProjectId, region)
	if err != nil {
		return nil, err
	}

	var u ResUser
	err = Filters{"OriginDomainId": project.OriginDomainId, "Region": region}.Filter(ResUserTable).One(&u)
	return &u, err
}

func GetResUserById(id string) (*ResUser, error) {
	var a ResUser
	err := Filters{"Id": id}.Filter(ResUserTable).One(&a)
	return &a, err
}

func InsertResUserByJson(to orm.TxOrmer, id string, resConf *user.UserResourceConfig) error {
	r := *BuildResUser(resConf)
	checkExist := ResUser{}
	_ = Filters{"OriginDomainId": r.OriginDomainId, "Region": r.Region}.Filter(ResUserTable).One(&checkExist)
	if checkExist != (ResUser{}) {
		return fmt.Errorf("originDomainId %s exist in %s", r.OriginDomainId, ResUserTable)
	}
	_, err := to.Insert(&r)
	if err != nil {
		_ = to.Rollback()
		return err
	}
	return nil
}

func UpdateResUserByJson(to orm.TxOrmer, request []byte, indexId string) error {
	newConfig := ResUser{}
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

func DeleteResUserById(to orm.TxOrmer, indexId string) error {
	resuser := ResUser{}
	err := Filters{"Id": indexId}.Filter(ResUserTable).One(&resuser)
	if err != nil {
		return err
	}
	_, err = to.Delete(&resuser)
	if err != nil {
		_ = to.Rollback()
		return fmt.Errorf("delete user err:%s", err.Error())
	}
	return nil
}
func DeleteResUser(originProjectId string, region string) error {
	project, err := GetResProjectByProjectId(originProjectId, region)
	if err != nil {
		return err
	}
	_, err = Filters{"OriginDomainId": project.OriginDomainId, "Region": region}.Filter(ResUserTable).Delete()
	return err
}
