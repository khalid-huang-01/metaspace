// Copyright (c) Huawei Technologies Co., Ltd. 2022-2022. All rights reserved.

// 鉴权校验中token数据体定义
package token

import (
	"fleetmanager/api/filter/authz/user"
	"time"
)

// Token security-center token stuct including Domain, Role, User, Project
type Token struct {
	TokenInfo TokenInfo `json:"token"`
}

// TokenInfo token info struct
type TokenInfo struct {
	HWContext user.HWContext `json:"hw_context,omitempty"`
	Domain    Domain         `json:"domain"`
	Roles     []user.Role    `json:"roles"`
	Project   Project        `json:"project"`
	User      User           `json:"user"`
	AssumedBy AssumedBy      `json:"assumed_by"`
	ExpiresAt time.Time      `json:"expires_at"`
	IssuedAt  time.Time      `json:"issued_at"`
}

type AssumedBy struct {
	User User `json:"user"`
}

// Domain domain info struct
type Domain struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	XDomainType string `json:"xdomain_type"`
	XDomainID   string `json:"xdomain_id"`
}

// User user info struct
type User struct {
	Domain Domain `json:"domain"`
	ID     string `json:"id"`
	Name   string `json:"name"`
}

// Project project info struct
type Project struct {
	Domain Domain `json:"domain"`
	ID     string `json:"id"`
	Name   string `json:"name"`
}


