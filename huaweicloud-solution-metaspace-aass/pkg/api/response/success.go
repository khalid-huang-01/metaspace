// Copyright (c) Huawei Technologies Co., Ltd. 2022-2022. All rights reserved.

// 响应成功
package response

import (
	"github.com/beego/beego/v2/server/web/context"

	"scase.io/application-auto-scaling-service/pkg/utils/logger"
)

// Success return success response
func Success(ctx *context.Context, status int, body interface{}) {
	ctx.Output.SetStatus(status)
	ctx.Output.Header(HttpContentType, MimeApplicationJSON)
	ctx.Output.Header(HttpContentTypeOptions, HttpOptionsNoSniff)
	requestId := ""
	if i := ctx.Input.GetData(logger.RequestId); i != nil {
		if id, ok := i.(string); ok {
			requestId = id
		}
	}
	ctx.Output.Header(HttpRequestId, requestId)

	if body != nil {
		if err := ctx.Output.JSON(body, true, false); err != nil {
			logger.R.Error("serve json error %v", err)
		}
	}
}
