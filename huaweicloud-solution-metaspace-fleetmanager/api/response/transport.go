// Copyright (c) Huawei Technologies Co., Ltd. 2022-2022. All rights reserved.

// 响应体转换
package response

import (
	"fleetmanager/logger"
	"github.com/beego/beego/v2/server/web/context"
)

// TransPort: 响应体转换
func TransPort(ctx *context.Context, status int, body []byte) {
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

	if err := ctx.Output.Body(body); err != nil {
		logger.R.Error("serve transport error: %v", err)
	}
}
