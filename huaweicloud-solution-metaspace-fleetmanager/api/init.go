// Copyright (c) Huawei Technologies Co., Ltd. 2022-2022. All rights reserved.

package api

import (
	"crypto/tls"
	"fleetmanager/api/cidrmanager"
	"fleetmanager/api/filter/authz"
	"fleetmanager/api/filter/entrance"
	"fleetmanager/api/filter/export"
	"fleetmanager/api/router"
	"fleetmanager/api/validator"
	"fleetmanager/setting"

	"github.com/beego/beego/v2/server/web"
	_ "github.com/beego/beego/v2/server/web/session/redis"
)

func initConfig() {
	web.BConfig.CopyRequestBody = true
	web.BConfig.RunMode = setting.DefaultAPIRunModel
	web.BConfig.WebConfig.AutoRender = false
	web.BConfig.CopyRequestBody = true
	// 开启会话以及相关配置
	web.BConfig.WebConfig.Session.SessionOn = true
	web.BConfig.WebConfig.Session.SessionName = "sessionid"
	web.BConfig.WebConfig.Session.SessionGCMaxLifetime = int64(setting.SessionLifeTime)
	web.BConfig.WebConfig.Session.SessionProvider = "redis"
	web.BConfig.WebConfig.Session.SessionProviderConfig = setting.RedisAddress +
		"," + setting.RedisMaxConn + "," + setting.RedisPassword

	// 如果开启https
	if setting.EnableHttps {
		initHttpsConfig()
	} else {
		web.BConfig.Listen.EnableHTTPS = false
	}

	if setting.EnableHttp {
		initHttpConfig()
	} else {
		web.BConfig.Listen.EnableHTTP = false
	}
}

func initHttpsConfig() {
	web.BConfig.Listen.EnableHTTPS = true
	web.BConfig.Listen.HTTPSAddr = setting.WebHttpsAddr
	web.BConfig.Listen.HTTPSPort = setting.WebHttpsPort
	web.BConfig.Listen.HTTPSCertFile = setting.HttpsCertFile
	web.BConfig.Listen.HTTPSKeyFile = setting.HttpsKeyFile
	web.BConfig.Listen.ClientAuth = int(tls.NoClientCert)
	web.BConfig.MaxUploadSize = setting.DefaultUploadSize
}

func initHttpConfig() {
	web.BConfig.Listen.EnableHTTP = true
	web.BConfig.Listen.HTTPPort = setting.WebHttpPort
}

// Init fleet manager初始化函数
func Init() error {
	initConfig()
	if err := validator.Init(); err != nil {
		return err
	}

	if err := cidrmanager.Init(); err != nil {
		return err
	}

	web.InsertFilter("/*", web.BeforeExec, entrance.Filter)

	// 校验Token
	if setting.EnableTokenCheck {
		web.InsertFilter("/*", web.BeforeExec, authz.Filter)
	}

	web.InsertFilter("/*", web.FinishRouter, export.Filter, web.WithReturnOnOutput(false))
	router.Init()
	return nil
}

// Run fleet manager api运行函数
func Run() {
	web.Run()
}
