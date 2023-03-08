// Copyright (c) Huawei Technologies Co., Ltd. 2022-2022. All rights reserved.

// http服务
package httpserver

import (
	context2 "context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"

	"github.com/beego/beego/v2/server/web"
	"github.com/beego/beego/v2/server/web/context"

	"codehub-g.huawei.com/videocloud/mediaprocesscenter/auxproxy/config"
	"codehub-g.huawei.com/videocloud/mediaprocesscenter/auxproxy/pkg/apis"
	"codehub-g.huawei.com/videocloud/mediaprocesscenter/auxproxy/pkg/common"
	"codehub-g.huawei.com/videocloud/mediaprocesscenter/auxproxy/pkg/processmanager"
	errors2 "codehub-g.huawei.com/videocloud/mediaprocesscenter/auxproxy/pkg/utils/errors"
	"codehub-g.huawei.com/videocloud/mediaprocesscenter/auxproxy/pkg/utils/hhmac"
	"codehub-g.huawei.com/videocloud/mediaprocesscenter/auxproxy/pkg/utils/log"
	"codehub-g.huawei.com/videocloud/mediaprocesscenter/auxproxy/pkg/utils/sdk/processservice"
)

type HttpSever struct {
	httpServer *web.HttpServer
	Addr       string
}

// Server http server
var Server *HttpSever

// InitHttpServer init http server
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

		Server = &HttpSever{
			Addr:       addr,
			httpServer: web.NewHttpSever(),
		}

		// 认证校验
		Server.httpServer.InsertFilter("/v1/server-sessions/start", web.BeforeStatic, BeforeFilter)

		// register router
		Server.httpServer.Post("/v1/server-sessions/start", StartServerSession)

		Server.httpServer.Post("/v1/cleanup", StartCleanUp)
		Server.httpServer.Get("/v1/cleanup-state", ShowCleanUpState)
	})
}

// 中间件
func BeforeFilter(ctx *context.Context) {
	// hmac 身份认证
	if config.ServerHmacConf.AuthEnable {
		if err := hhmac.ValidateHmac(ctx.Request); err != nil {
			log.RunLogger.Infof("Validate hmac err: %+v", err)
			Response(ctx, http.StatusForbidden, errors2.NewAuthenticationError())
			return
		}
	}
}

// Work let http server work
func (h *HttpSever) Work() {
	go func() {
		h.httpServer.Run(h.Addr)
	}()
}

// Response response
func Response(ctx *context.Context, statusCode int, body interface{}) {
	ctx.Output.SetStatus(statusCode)
	ctx.JSONResp(body)
}

// StartServerSession start server session
func StartServerSession(ctx *context.Context) {
	var ss apis.ServerSession
	err := json.Unmarshal(ctx.Input.RequestBody, &ss)
	if err != nil {
		log.RunLogger.Errorf("[http server] failed to unmarshal start server session for %v", err)
		errResp := errors2.NewStartServerSessionError(
			fmt.Sprintf("failed to unmarshal start server session for %v", err), http.StatusBadRequest)
		Response(ctx, errResp.HttpCode, errResp)
		return
	}

	process := processmanager.ProcessMgr.GetProcess(ss.PID)
	if process == nil {
		log.RunLogger.Errorf("[http server] process manager do not exist process id %d", ss.PID)
		errResp := errors2.NewStartServerSessionError(
			fmt.Sprintf(" process manager do not exist process id %d", ss.PID), http.StatusInternalServerError)
		Response(ctx, errResp.HttpCode, errResp)
		return
	}
	go startServerSession(process, &ss)
	ctx.Output.SetStatus(http.StatusOK)

}

func startServerSession(process *processmanager.Process, ss *apis.ServerSession) {
	if processmanager.ProcessMgr.IsServerSessionStarted(process, ss.ID) {
		log.RunLogger.Infof("[http server] process %d already stared the server session %s, start "+
			"server session process just return", ss.PID, ss.ID)
		return
	}

	ctx2 := context2.Background()
	joinable := false
	if ss.ClientSessionCreationPolicy == common.ClientSessionCreationPolicyAcceptAll {
		joinable = true
	}
	gameProperties := cover2GameProperties(ss.SessionProperties)
	req := &processservice.StartServerSessionRequest{
		ServerSession: &processservice.ServerSession{
			ServerSessionId:   ss.ID,
			FleetId:           ss.FleetID,
			Name:              ss.Name,
			MaxClients:        int32(ss.MaxClientSessionNum),
			Joinable:          joinable,
			SessionProperties: gameProperties,
			Port:              int32(ss.ClientPort),
			IpAddress:         ss.PublicIP,
			SessionData:       ss.SessionData,
		},
	}
	log.RunLogger.Infof("[http server] fetch process %d start call on start game "+
		"server session", process.Pid)
	_, err := process.Client.OnStartServerSession(ctx2, req)
	if err != nil {
		log.RunLogger.Errorf("[http server] call OnStartGameServerSession failed for %v", err)
	} else {
		// 记录server session 已经启动过
		processmanager.ProcessMgr.RecordServerSessionStarted(process, ss.ID)
	}

}

func cover2GameProperties(ssProperties []apis.KV) []*processservice.SessionProperty {
	if len(ssProperties) == 0 {
		return nil
	}
	rsl := make([]*processservice.SessionProperty, len(ssProperties))
	for i, kv := range ssProperties {
		rsl[i] = &processservice.SessionProperty{
			Key:   kv.Key,
			Value: kv.Value,
		}
	}
	return rsl
}
