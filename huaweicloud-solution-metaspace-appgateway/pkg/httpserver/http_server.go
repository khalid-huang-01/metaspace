// Copyright (c) Huawei Technologies Co., Ltd. 2022-2022. All rights reserved.

// http请求响应构造
package httpserver

import (
	"crypto/tls"
	"fmt"
	"net/http"
	"sync"

	"github.com/beego/beego/v2/server/web"
	"github.com/beego/beego/v2/server/web/context"

	"codehub-g.huawei.com/videocloud/mediaprocesscenter/application-gateway/config"
	"codehub-g.huawei.com/videocloud/mediaprocesscenter/application-gateway/pkg/utils/hhmac"
	"codehub-g.huawei.com/videocloud/mediaprocesscenter/application-gateway/pkg/utils/log"
)

type HttpServer struct {
	httpServer *web.HttpServer
	Addr       string
}

var Server *HttpServer

func InitHttpServer(addr string) {
	var once sync.Once
	once.Do(func() {
		// init beego web
		// set this so that we can get request body
		// https配置
		web.BConfig.CopyRequestBody = true
		// 禁用http，只用https
		web.BConfig.Listen.EnableHTTP = false
		web.BConfig.Listen.EnableHTTPS = true
		web.BConfig.Listen.HTTPSAddr = config.HttpsAddr

		web.BConfig.Listen.HTTPSPort = config.HttpsPort
		// ssl 证书路径
		web.BConfig.Listen.HTTPSCertFile = config.HttpsCertFilePath
		// ssl 证书 keyfile 的路径
		web.BConfig.Listen.HTTPSKeyFile = config.HttpsKeyFilePath
		// 不校验客户端证书合法性，且客户端可不传证书
		web.BConfig.Listen.ClientAuth = int(tls.NoClientCert)

		Server = &HttpServer{
			Addr:       addr,
			httpServer: web.NewHttpSever(),
		}

		// 认证校验
		Server.httpServer.InsertFilter("/*", web.BeforeStatic, BeforeFilter)
	})
}

// 中间件
func BeforeFilter(ctx *context.Context) {
	// hmac 身份认证
	if config.ServerHmacConf.AuthEnable {
		if err := hhmac.ValidateHmac(ctx.Request); err != nil {
			log.RunLogger.Errorf("Validate hmac err: %+v", err)
			Response(ctx, http.StatusForbidden, fmt.Errorf("authentication error"))
			return
		}
	}
}

// Work let http server work
func (h *HttpServer) Work() {
	go func() {
		h.httpServer.Run(h.Addr)
	}()
}

// Response response
func Response(ctx *context.Context, statusCode int, body interface{}) {
	ctx.Output.SetStatus(statusCode)
	ctx.JSONResp(body)
}
