// Copyright (c) Huawei Technologies Co., Ltd. 2022-2022. All rights reserved.

// 响应体构建
package controllers

import (
	"github.com/beego/beego/v2/server/web/context"

	"codehub-g.huawei.com/videocloud/mediaprocesscenter/application-gateway/pkg/utils/log"
)

// Response response
func Response(ctx *context.Context, statusCode int, body interface{}) {
	ctx.Input.SetData(log.ResponseCode, statusCode)
	ctx.Output.SetStatus(statusCode)
	ctx.JSONResp(body)
}
