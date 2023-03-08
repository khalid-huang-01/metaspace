// Copyright (c) Huawei Technologies Co., Ltd. 2022-2022. All rights reserved.

// 查询校验方法
package query

import (
	"fleetmanager/api/params"
	"fmt"
	"github.com/beego/beego/v2/server/web/context"
	"strconv"
)

const (
	MiniOffset    = 0
	MaxiOffset    = 10
	DefaultOffset = 0
	MiniLimit     = 1
	MaxiLimit     = 100
	DefaultLimit  = 100
	MaxServerSessionNum = 250
)

// CheckOffset: 校验offset字段是否正常
func CheckOffset(ctx *context.Context) (int, error) {
	offset, err := strconv.Atoi(ctx.Input.Query(params.QueryOffset))
	if err != nil {
		offset = DefaultOffset
	}

	if offset < MiniOffset || offset > MaxiOffset {
		return offset, fmt.Errorf("offset query must between %v and %v", MiniOffset, MaxiOffset)
	}

	return offset, nil
}

// CheckLimit: 校验Limit字段是否正常
func CheckLimit(ctx *context.Context) (int, error) {
	limit, err := strconv.Atoi(ctx.Input.Query(params.QueryLimit))
	if err != nil {
		limit = DefaultLimit
	}

	if limit < MiniLimit || limit > MaxiLimit {
		return limit, fmt.Errorf("limit query must between %v and %v", MiniLimit, MaxiLimit)
	}

	return limit, nil
}
