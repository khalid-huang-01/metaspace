// Copyright (c) Huawei Technologies Co., Ltd. 2022-2022. All rights reserved.

// 过滤入口
package entrance

import (
	"fmt"
	"net/http"
	"time"

	"github.com/beego/beego/v2/server/web/context"
	"github.com/google/uuid"

	apierr "scase.io/application-auto-scaling-service/pkg/api/errors"
	"scase.io/application-auto-scaling-service/pkg/api/response"
	"scase.io/application-auto-scaling-service/pkg/setting"
	"scase.io/application-auto-scaling-service/pkg/utils"
	"scase.io/application-auto-scaling-service/pkg/utils/hhmac"
	"scase.io/application-auto-scaling-service/pkg/utils/logger"
)

func generateRequestId() string {
	u, _ := uuid.NewUUID()
	uid := u.String()
	timestamp := time.Now().Unix()
	hostname := utils.GetLocalIP()
	return fmt.Sprintf("%s-%d-%s", uid, timestamp, hostname)
}

// todo: token鉴权
// Filter filter request
func Filter(ctx *context.Context) {
	requestId := generateRequestId()
	traceLogger := logger.R.WithField(logger.RequestId, requestId)

	ctx.Input.SetData(logger.StartTime, time.Now())
	ctx.Input.SetData(logger.RequestId, requestId)
	ctx.Input.SetData(logger.TraceLogger, traceLogger)

	// hmac 身份认证
	if setting.ServerHmacConf.AuthEnable {
		if err := hhmac.ValidateHmac(ctx.Request); err != nil {
			// 注意：这里 err 需要做脱敏处理
			traceLogger.Error("Validate hmac err: %+v", err)
			response.Error(ctx, http.StatusForbidden, apierr.NewErrorResp(apierr.AuthenticationError))
			return
		}
	}
}
