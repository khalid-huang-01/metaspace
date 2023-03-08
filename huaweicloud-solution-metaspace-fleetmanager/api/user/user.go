// Copyright (c) Huawei Technologies Co., Ltd. 2022-2022. All rights reserved.

// 用户管理
package user

//
//  ResDomainAgency
//  @Description: ResDomainAgency
//
type ResDomainAgency struct {
	ResProjectId        string
	ResDomainId         string
	AgencyName          string
	OriginUserDomain    string
}


//
//  UserSecurityInfo
//  @Description:
//
type UserSecurityInfo struct {
	AK        string
	SK         string
	Token          string
}