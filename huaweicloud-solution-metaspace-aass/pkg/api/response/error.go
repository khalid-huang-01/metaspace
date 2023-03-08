// Copyright (c) Huawei Technologies Co., Ltd. 2022-2022. All rights reserved.

// 响应错误
package response

import (
	"fmt"

	"github.com/beego/beego/v2/server/web/context"

	"scase.io/application-auto-scaling-service/pkg/api/errors"
	"scase.io/application-auto-scaling-service/pkg/utils/logger"
)

// Error return error response
func Error(ctx *context.Context, status int, body interface{}) {
	// 错误响应必须带错误信息，不允许只带一个错误响应码
	if body == nil {
		logger.R.Error("con not response error without body")
		return
	}
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

	if _, ok := body.(*errors.ErrorResp); !ok {
		body = &errors.ErrorResp{
			ErrCode: errors.UnKnown,
			ErrMsg:  fmt.Sprintf("%v", body),
		}
	}

	if err := ctx.Output.JSON(body, true, false); err != nil {
		logger.R.Error("serve json error %v", err)
	}
}
