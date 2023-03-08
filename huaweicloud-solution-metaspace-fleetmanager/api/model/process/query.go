// Copyright (c) Huawei Technologies Co., Ltd. 2022-2022. All rights reserved.

// 进程查询结构体构建
package process

// 接受转换后数据结构
type ListProcess struct {
	Count     int       `json:"count"`
	Processes []Process `json:"app_processes"`
}

// 接受appgw接口返回结果结构
type ListAppProcessesResponse struct {
	Count        int          `json:"count"`
	AppProcesses []AppProcess `json:"app_processes"`
}
