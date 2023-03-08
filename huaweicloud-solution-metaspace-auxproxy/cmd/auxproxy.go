// Copyright (c) Huawei Technologies Co., Ltd. 2022-2022. All rights reserved.

// auxproxy启动入口
package main

import (
	"flag"
	"fmt"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"

	"codehub-g.huawei.com/videocloud/mediaprocesscenter/auxproxy/config"
	"codehub-g.huawei.com/videocloud/mediaprocesscenter/auxproxy/pkg/buildmanager"
	"codehub-g.huawei.com/videocloud/mediaprocesscenter/auxproxy/pkg/configmanager"
	"codehub-g.huawei.com/videocloud/mediaprocesscenter/auxproxy/pkg/grpcserver"
	"codehub-g.huawei.com/videocloud/mediaprocesscenter/auxproxy/pkg/httpserver"
	"codehub-g.huawei.com/videocloud/mediaprocesscenter/auxproxy/pkg/processmanager"
	"codehub-g.huawei.com/videocloud/mediaprocesscenter/auxproxy/pkg/restorer"
	"codehub-g.huawei.com/videocloud/mediaprocesscenter/auxproxy/pkg/utils/clients"
	"codehub-g.huawei.com/videocloud/mediaprocesscenter/auxproxy/pkg/utils/hhmac"
	"codehub-g.huawei.com/videocloud/mediaprocesscenter/auxproxy/pkg/utils/log"
)

func init() {
	flag.StringVar(&config.Opts.CloudPlatformAddr, "cloud-platform-address",
		"", "cloud platform address for user data and meta data")
	flag.StringVar(&config.Opts.AuxProxyAddr, "auxproxy-address", "", "auxproxy address")
	flag.StringVar(&config.Opts.GrpcAddr, "grpc-address", "",
		"grpc address, this should listen at localhost, or we can not get right peer ip")
	flag.StringVar(&config.Opts.GatewayAddr, "gateway-address", "", "application gateway address")
	flag.StringVar(&config.Opts.CmdFleetId, "cmd-fleet-id", "", "fleet id from cmd")
	flag.StringVar(&config.Opts.ScalingGroupId, "scaling-group-id", "", "scaling group id from cmd")

	flag.StringVar(&config.Opts.LogLevel, "log-level", config.LogLevelInfo, "log level, support "+
		"debug and info")
	flag.BoolVar(&config.Opts.EnableBuild, "enable-build-download", false, "build download enable")
	flag.BoolVar(&config.Opts.EnableTest, "enable-test-mode", false, "test enable")
	flag.StringVar(&config.Opts.GCMKey, "gcm-key", "", "gcm decode or encode key")
	flag.StringVar(&config.Opts.GCMNonce, "gcm-nonce", "", "gcm decode or encode nonce")
}

// ReturnErr return when err is not nil
func ReturnErr(err error) {
	if err != nil {
		fmt.Printf("auxproxy init failed for %v\n", err)
		os.Exit(1)
	}
}

func main() {
	flag.Parse()

	ReturnErr(log.InitLog())

	if config.Opts.EnableTest {
		log.RunLogger.Infof("Start auxproxy success")
		return
	}
	ReturnErr(config.Init())
	ReturnErr(hhmac.InitHMACKey())
	// 如果命令行启用了应用下载功能，下载完成后直接结束进程
	if config.Opts.EnableBuild {
		ReturnErr(buildmanager.Init())
		return
	}

	// 1. init process manager
	processmanager.InitProcessManager()

	// 2. init config manager
	_, p, err := net.SplitHostPort(config.Opts.AuxProxyAddr)
	if err != nil {
		log.RunLogger.Errorf("[main] invalid auxproxy addr %s", config.Opts.AuxProxyAddr)
		return
	}
	log.RunLogger.Infof("auxproxyaddr %s, port %s", config.Opts.AuxProxyAddr, p)
	err = configmanager.InitConfigManager(60 * time.Second) // default config fetch interval
	if err != nil {
		log.RunLogger.Errorf("[main] failed to init config manager for %v", err)
		return
	}

	// 3. init gateway client
	clients.InitGatewayClient(configmanager.ConfMgr.GatewayAddr)

	// 4. let config manager to get runtime config
	configmanager.ConfMgr.Work()

	// wait for config manager to get config
	<-configmanager.ConfMgr.ConfigGetChan

	// 5. init and start grpc server
	grpcserver.InitGrpcServer(configmanager.ConfMgr.Config.FleetID, configmanager.ConfMgr.Config.InstanceID,
		config.Opts.GrpcAddr)
	grpcserver.GServer.Work()

	httpserver.InitHttpServer(config.Opts.AuxProxyAddr)
	httpserver.Server.Work()

	// 6. let process manager to start process
	restorer.Work()
	processmanager.ProcessMgr.Work()

	c := make(chan os.Signal)
	signal.Notify(c, syscall.SIGINT, syscall.SIGTERM)
	<-c
}
