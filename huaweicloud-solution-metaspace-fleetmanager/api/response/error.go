// Copyright (c) Huawei Technologies Co., Ltd. 2022-2022. All rights reserved.

// 响应错误信息定义
package response

import (
	"fleetmanager/api/errors"
	"fleetmanager/logger"
	"fmt"
	"github.com/beego/beego/v2/server/web/context"
	"net/http"
)

// Error: 构造Error对象
func Error(ctx *context.Context, status int, body interface{}) {
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
		if _, ok := body.(*errors.CodedError); !ok {
			body = &errors.CodedError{
				ErrC: errors.ServerInternalError,
				ErrD: fmt.Sprintf("%v", body),
			}
		}

		if err := ctx.Output.JSON(body, true, false); err != nil {
			logger.R.Error("serve json error %v", err)
		}
	}
}

// ProjectMismatchError: 请求的资源与Project不匹配错误
func ProjectMismatchError(ctx *context.Context) {
	body := errors.NewError(errors.ProjectMismatchError)
	Error(ctx, http.StatusBadRequest, body)
}

// InputError: 请求体格式错误
func InputError(ctx *context.Context) {
	body := errors.NewErrorF(errors.InvalidParameterValue, " invalid request body.")
	Error(ctx, http.StatusBadRequest, body)
}

// ParamsError: 请求入参错误
func ParamsError(ctx *context.Context, err error) {
	body := errors.NewErrorF(errors.InvalidParameterValue, err.Error())
	Error(ctx, http.StatusBadRequest, body)
}

// InternalError: 内部服务错误
func InternalError(ctx *context.Context, err error) {
	body := errors.NewErrorF(errors.ServerInternalError, err.Error())
	Error(ctx, http.StatusInternalServerError, body)
}

// ServiceError: 基于错误码对外返回错误
func ServiceError(ctx *context.Context, err *errors.CodedError) {
	switch err.ErrC {
	case errors.ServerInternalError, errors.DBError:
		Error(ctx, http.StatusInternalServerError, err)
	case errors.PolicyNotFound:
		Error(ctx, http.StatusNotFound, err)
	case errors.FleetNotFound:
		Error(ctx, http.StatusNotFound, err)
	default:
		Error(ctx, http.StatusBadRequest, err)
	}
}
