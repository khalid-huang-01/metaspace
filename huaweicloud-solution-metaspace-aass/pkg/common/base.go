
// Copyright (c) Huawei Technologies Co., Ltd. 2022-2022. All rights reserved.

// 常量定义
package common

import (
	"github.com/beego/beego/v2/server/web/context"
	"strconv"
	"time"
	"scase.io/application-auto-scaling-service/pkg/utils/logger"
)

func GetStartNumber(ctx *context.Context, tLogger *logger.FMLogger) int {
	start_number, err := strconv.Atoi(ctx.Input.Query(ParamOffset))
	if err != nil {
		tLogger.Warn("[query param getter] offset is not valid, "+
			"use default offset %d", DefaultStartNumber)
		start_number = DefaultStartNumber
	}

	return start_number
}

func GetLimit(ctx *context.Context, tLogger *logger.FMLogger) int {
	limit, err := strconv.Atoi(ctx.Input.Query(ParamLimit))

	if err != nil || limit <= 0{
		tLogger.Warn("[query param getter] limit is not valid, "+
			"use default limit %d", DefaultLimit)
		limit = DefaultLimit
	}

	return limit
}

func GetTime(ctx *context.Context, tLogger *logger.FMLogger) (time.Time) {
	created_at, err := time.ParseInLocation(TimeLayout, ctx.Input.Query(ParamCreatedAt), time.Local)
	if err != nil {
		created_at = time.Now().AddDate(0, 0, -7)
		tLogger.Warn("[query param getter] created_at is not valid, "+
			"use before 7 days %s", created_at)
	}
	return created_at
}