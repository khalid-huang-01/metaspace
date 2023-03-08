// Copyright (c) Huawei Technologies Co., Ltd. 2022-2022. All rights reserved.

// 应用包查询结构体创建
package build

// List
// @Description:
type List struct {
	TotalCount int64   `json:"total_count"`
	Count      int     `json:"count"`
	Builds     []Build `json:"builds"`
}

type BucketList struct {
	Count      int      `json:"count"`
	BucketName []string `json:"bucket_name"`
}

type ImageList struct {
	Count     int     `json:"count"`
	ImageList []Image `json:"image_list"`
}

type Image struct {
	Id      string `json:"image_id"`
	Name    string `json:"image_name"`
	Version string `json:"image_version"`
}

type VpcList struct {
	Count   int   `json:"count"`
	VpcList []Vpc `json:"vpc_list"`
}

type Vpc struct {
	Id   string `json:"vpc_id"`
	Name string `json:"vpc_name"`
}

type SubnetList struct {
	Count   int      `json:"count"`
	VpcList []Subnet `json:"subnet_list"`
}

type Subnet struct {
	Id   string `json:"subnet_id"`
	Name string `json:"subnet_name"`
}

// UploadCredentials
// @Description:
type UploadCredentials struct {
	BucketName      string `json:"bucket_name"`
	BucketKey       string `json:"bucket_key"`
	RegionId        string `json:"region_id"`
	AccessKeyId     string `json:"access_key_id"`
	SecretAccessKey string `json:"secret_access_key"`
	SecurityToken   string `json:"security_token"`
}

type FleetMsg struct {
	FleetId      string `json:"fleet_id"`
	FleetName    string `json:"fleet_name"`
	FleetState   string `json:"fleet_state"`
	CreationTime string `json:"creation_time"`
}

type FleetList struct {
	Count int        `json:"count"`
	Fleet []FleetMsg `json:"fleet"`
}
