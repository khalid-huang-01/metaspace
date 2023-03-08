// Copyright (c) Huawei Technologies Co., Ltd. 2022-2022. All rights reserved.

// 鉴权校验前过滤非法信息
package authz

import (
	"fleetmanager/api/common/log"
	"fleetmanager/api/errors"
	"fleetmanager/api/params"
	"fleetmanager/api/response"
	"fleetmanager/logger"
	"github.com/beego/beego/v2/server/web/context"
	"net/http"
)

// Filter: authz过滤器
func Filter(ctx *context.Context) {
	tLogger := log.GetTraceLogger(ctx).WithField(logger.Stage, "token_validate")
	projectId := ctx.Input.Param(params.ProjectId)
	token := ctx.Input.Header(params.HeaderParameterToken)

	// token check
	userInfo, err, errorCode := AuthenticateToken([]byte(token))
	if err != nil {
		response.Error(ctx, http.StatusBadRequest, errors.NewError(errorCode))
		tLogger.WithField(logger.Error, errorCode).Error(err.Error())
		return
	}

	// make sure user correct
	if projectId != "" && projectId != userInfo.GetProjectID() {
		response.ProjectMismatchError(ctx)
		tLogger.WithField(logger.Error, errorCode).Error(err.Error())
	}
}
