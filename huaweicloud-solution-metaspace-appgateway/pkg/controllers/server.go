// Copyright (c) Huawei Technologies Co., Ltd. 2022-2022. All rights reserved.

// 服务端应用初始化
package controllers

import (
	"crypto/tls"
	"net/http"

	"github.com/beego/beego/v2/server/web"
	"github.com/beego/beego/v2/server/web/context"

	"codehub-g.huawei.com/videocloud/mediaprocesscenter/application-gateway/config"
	"codehub-g.huawei.com/videocloud/mediaprocesscenter/application-gateway/pkg/utils/errors"
	"codehub-g.huawei.com/videocloud/mediaprocesscenter/application-gateway/pkg/utils/hhmac"
	"codehub-g.huawei.com/videocloud/mediaprocesscenter/application-gateway/pkg/utils/log"
)

func Init() {
	initConfig()
	web.InsertFilter("/*", web.BeforeStatic, BeforeFilter)
}

func initConfig() {
	// init beego web
	// set this so that we can get request body
	// https配置
	web.BConfig.CopyRequestBody = true
	// 禁用http，只用https
	web.BConfig.Listen.EnableHTTP = false
	web.BConfig.Listen.EnableHTTPS = true
	web.BConfig.Listen.HTTPSAddr = ""
	// 安全要求：监听本地 "127.0.0.1" 的地址
	web.BConfig.Listen.HTTPSAddr = config.HttpsAddr
	web.BConfig.Listen.HTTPSPort = config.HttpsPort
	// ssl 证书路径
	web.BConfig.Listen.HTTPSCertFile = config.HttpsCertFilePath
	// ssl 证书 keyfile 的路径
	web.BConfig.Listen.HTTPSKeyFile = config.HttpsKeyFilePath
	// 不校验客户端证书合法性，且客户端可不传证书
	web.BConfig.Listen.ClientAuth = int(tls.NoClientCert)
}

// BeforeFilter 中间件
func BeforeFilter(ctx *context.Context) {
	// hmac 身份认证
	if config.ServerHmacConf.AuthEnable {
		if err := hhmac.ValidateHmac(ctx.Request); err != nil {
			log.RunLogger.Errorf("Validate hmac err: %+v", err)
			Response(ctx, http.StatusForbidden, errors.NewAuthenticationError())
			return
		}
	}
}
