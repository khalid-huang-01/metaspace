// Copyright (c) Huawei Technologies Co., Ltd. 2022-2022. All rights reserved.

// 进程列表响应
package model

import (
	"github.com/huaweicloud/huaweicloud-sdk-go-v3/core/utils"
	"strings"
)

// Response Object
type ListProcessResponse struct {
	Count        int       `json:"count"`
	AppProcesses []Process `json:"app_processes"`
}

type Process struct {
	AppProcessId string `json:"app_process_id"`
	State        string `json:"state"`
}

// ListProcessResponse 打印ListProcessResponse信息
func (o ListProcessResponse) String() string {
	data, err := utils.Marshal(o)
	if err != nil {
		return "ListProcessResponse struct{}"
	}

	return strings.Join([]string{"ListProcessResponse", string(data)}, " ")
}
