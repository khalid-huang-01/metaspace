// Copyright (c) Huawei Technologies Co., Ltd. 2022-2022. All rights reserved.

// 应用包创建结构体定义
package build

// CreateRequest
// @Description:
type CreateRequest struct {
	Name              string          `json:"name" validate:"required,buildName"`
	Description       string          `json:"description,omitempty" validate:"omitempty,min=1,max=100"`
	OperatingSystem   string          `json:"operating_system,omitempty" validate:"omitempty,min=1,max=50"`
	StorageLocation   StorageLocation `json:"storage_location"`
	Version           string          `json:"version" validate:"required,min=1,max=50"`
	Region            string          `json:"region" validate:"required,min=1,max=50"`
	VpcId             string          `json:"vpc_id" validate:"min=1,max=64"`
	SubnetId          string          `json:"subnet_id" validate:"min=1,max=64"`
	EnterpriseProject string          `json:"enterprise_project" validate:"omitempty,min=1,max=64"`
}

type CreateByImageRequest struct {
	Name              string `json:"name" validate:"required,buildName"`
	Description       string `json:"description,omitempty" validate:"omitempty,min=1,max=100"`
	ImageId           string `json:"image_id" validate:"required"`
	Version           string `json:"version" validate:"required,min=1,max=50"`
	Region            string `json:"region" validate:"required,min=1,max=50"`
	VpcName           string `json:"vpc_name" validate:"omitempty,min=1,max=64"`
	SubnetName        string `json:"subnet_name" validate:"omitempty,min=1,max=64"`
	EnterpriseProject string `json:"enterprise_project" validate:"omitempty,min=1,max=64"`
}

type QueryRequest struct {
	Name         string `json:"name" validate:"min=1,max=50"`
	State        string `json:"state" validate:"min=1,max=50"`
	CreationTime string `json:"creation_time" validate:"min=1,max=200"`
}

// StorageLocation
// @Description:
type StorageLocation struct {
	BucketName string `json:"bucket_name" validate:"required,min=1,max=100"`
	BucketKey  string `json:"bucket_key" validate:"required,min=1,max=100"`
}

type CreateResponse struct {
	Build Build `json:"build"`
}

type UploadResponse struct {
	BucketName    string `json:"bucket_name" validate:"required,min=1,max=100"`
	BucketKey     string `json:"bucket_key" validate:"required,min=1,max=100"`
	StorageRegion string `json:"storage_region" validate:"required,min=1,max=100"`
}
