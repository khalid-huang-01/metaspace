// Copyright (c) Huawei Technologies Co., Ltd. 2022-2022. All rights reserved.

// 服务初始化
package api

import (
	"crypto/tls"
	"fmt"
	"net/http"
	"runtime"

	"github.com/beego/beego/v2/server/web"
	"github.com/beego/beego/v2/server/web/context"

	"scase.io/application-auto-scaling-service/pkg/api/errors"
	"scase.io/application-auto-scaling-service/pkg/api/filter/entrance"
	"scase.io/application-auto-scaling-service/pkg/api/filter/export"
	"scase.io/application-auto-scaling-service/pkg/api/response"
	"scase.io/application-auto-scaling-service/pkg/api/validator"
	"scase.io/application-auto-scaling-service/pkg/setting"
	"scase.io/application-auto-scaling-service/pkg/utils/logger"
)

func initConfig() {
	web.BConfig.RunMode = "prod"
	web.BConfig.CopyRequestBody = true
	web.BConfig.WebConfig.AutoRender = false

	// 禁用http，只用https
	web.BConfig.Listen.EnableHTTP = false
	web.BConfig.Listen.EnableHTTPS = true
	// 安全要求：监听本地 "127.0.0.1" 的地址
	web.BConfig.Listen.HTTPSAddr = setting.HttpsListenAddr
	web.BConfig.Listen.HTTPSPort = setting.GetWebHttpPort()
	// ssl 证书路径
	web.BConfig.Listen.HTTPSCertFile = setting.HttpsCertFile
	// ssl 证书 keyfile 的路径
	web.BConfig.Listen.HTTPSKeyFile = setting.HttpsKeyFile
	// 不校验客户端证书合法性，且客户端可不传证书
	web.BConfig.Listen.ClientAuth = int(tls.NoClientCert)

	// panic 恢复
	web.BConfig.RecoverPanic = true
	web.BConfig.RecoverFunc = RecoverPanic
}

// RecoverPanic 记录处理api请求时发生的panic错误
func RecoverPanic(ctx *context.Context, config *web.Config) {
	if err := recover(); err != nil {
		var stack string
		for i := 1; ; i++ {
			_, file, line, ok := runtime.Caller(i)
			if !ok {
				break
			}
			stack += fmt.Sprintf("\n%s:%d", file, line)
		}
		log := logger.GetTraceLogger(ctx)
		log.Error("Recover panic err: %+v;\n Stack info: %s", err, stack)

		response.Error(ctx, http.StatusInternalServerError,
			errors.NewErrorRespWithHttpCode(errors.ServerInternalError, http.StatusInternalServerError))
	}
}

// Init init api server...
func Init() error {
	initConfig()

	if err := validator.Init(); err != nil {
		return err
	}

	web.InsertFilter("/*", web.BeforeStatic, entrance.Filter)
	web.InsertFilter("/*", web.FinishRouter, export.Filter, web.WithReturnOnOutput(false))
	initRouters()

	return nil
}

func Run() {
	web.Run()
}
