// Copyright (c) Huawei Technologies Co., Ltd. 2022-2022. All rights reserved.

// 用户委托
package dao

import (
	"time"
)

//
//  UserAgency
//  @Description:
//
type UserAgency struct {
	Id             string    `orm:"column(id);size(64);pk" json:"id"`
	OriginDomainId string    `orm:"column(origin_domain_id);size(64)" json:"origin_domain_id"`
	AgencyName     string    `orm:"column(agency_name);" json:"agency_name"`
	CreationTime   time.Time `orm:"column(creation_time);type(datetime);auto_now_add" json:"creation_time"`
	Region         string    `orm:"column(region);size(64)" json:"region"`
}


// @Title GetOriginUserAgencyByDomainId
// @Description
// @Author wangnannan 2022-05-07 10:20:44 ${time}
// @Param originDomainId
// @Param region
// @Return *UserAgency
// @Return error
func GetOriginUserAgencyByDomainId(originDomainId string, region string) (*UserAgency, error) {
	var a UserAgency
	err := Filters{"OriginDomainId": originDomainId, "Region": region}.Filter(UserAgencyTable).One(&a)
	return &a, err
}
