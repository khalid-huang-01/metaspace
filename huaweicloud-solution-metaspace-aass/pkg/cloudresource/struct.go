// Copyright (c) Huawei Technologies Co., Ltd. 2022-2022. All rights reserved.

// 弹性伸缩组结构体与证书配置
package cloudresource

type credentialInfo struct {
	Access        string
	Secret        string
	SecurityToken string
}

type CreatAsGroupParams struct {
	DesireNumber int32
	AsConfigId   string
	VpcId        string
	SubnetId     string
	FleetId      string
	EnterpriseProjectId string
	IamAgencyName 		string
}
