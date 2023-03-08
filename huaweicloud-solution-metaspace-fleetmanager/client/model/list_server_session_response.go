// Copyright (c) Huawei Technologies Co., Ltd. 2022-2022. All rights reserved.

// 服务端会话列表响应
package model

import (
	"github.com/huaweicloud/huaweicloud-sdk-go-v3/core/utils"
	"strings"
)

// Response Object
type ListServerSessionResponse struct {
	Count          int             `json:"count"`
	ServerSessions []ServerSession `json:"server_sessions"`
}

type ServerSession struct {
	ServerSessionId string `json:"server_session_id"`
	State           string `json:"state"`
}

// ListServerSessionResponse 打印ListServerSessionResponse信息
func (o ListServerSessionResponse) String() string {
	data, err := utils.Marshal(o)
	if err != nil {
		return "ListServerSessionResponse struct{}"
	}

	return strings.Join([]string{"ListServerSessionResponse", string(data)}, " ")
}
