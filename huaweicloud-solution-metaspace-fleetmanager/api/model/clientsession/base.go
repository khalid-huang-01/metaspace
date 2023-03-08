// Copyright (c) Huawei Technologies Co., Ltd. 2022-2022. All rights reserved.

// 客户端会话基础结构体定义
package clientsession

type Session struct {
	ClientId   string `json:"client_id" validate:"required,min=1,max=128"`
	ClientData string `json:"client_data" validate:"min=0,max=128"`
}
