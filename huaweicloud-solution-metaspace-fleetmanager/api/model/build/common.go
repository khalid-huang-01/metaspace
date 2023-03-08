// Copyright (c) Huawei Technologies Co., Ltd. 2022-2022. All rights reserved.

// 应用包通用结构体定义
package build

// Build
// @Description:
type Build struct {
	BuildId         string `json:"build_id"`
	Name            string `json:"name"`
	Description     string `json:"description"`
	State           string `json:"state"`
	Size            int64  `json:"size"`
	CreationTime    string `json:"creation_time"`
	Version         string `json:"version"`
	OperatingSystem string `json:"operating_system"`
}

type FullBuild struct {
	Id                string `json:"id"`
	ProjectId         string `json:"project_id"`
	Name              string `json:"name"`
	Description       string `json:"description"`
	State             string `json:"state"`
	CreationTime      string `json:"creation_time"`
	UpdateTime        string `json:"update_time"`
	ImageId           string `json:"image_id"`
	ImageRegion       string `json:"image_region"`
	StorageBucketName string `json:"storage_bucket_name"`
	StorageKey        string `json:"storage_key"`
	StorageRegion     string `json:"storage_region"`
	OperatingSystem   string `json:"operating_system"`
	Version           string `json:"version"`
	Size              int64  `json:"size"`
}
