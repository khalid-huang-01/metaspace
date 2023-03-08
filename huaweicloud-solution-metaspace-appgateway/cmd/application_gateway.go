// Copyright (c) Huawei Technologies Co., Ltd. 2022-2022. All rights reserved.
package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/beego/beego/v2/server/web"

	"codehub-g.huawei.com/videocloud/mediaprocesscenter/application-gateway/config"
	"codehub-g.huawei.com/videocloud/mediaprocesscenter/application-gateway/pkg/controllers"
	"codehub-g.huawei.com/videocloud/mediaprocesscenter/application-gateway/pkg/filters"
	"codehub-g.huawei.com/videocloud/mediaprocesscenter/application-gateway/pkg/metrics"
	"codehub-g.huawei.com/videocloud/mediaprocesscenter/application-gateway/pkg/models"
	"codehub-g.huawei.com/videocloud/mediaprocesscenter/application-gateway/pkg/routers"
	"codehub-g.huawei.com/videocloud/mediaprocesscenter/application-gateway/pkg/task"
	"codehub-g.huawei.com/videocloud/mediaprocesscenter/application-gateway/pkg/utils/clean"
	"codehub-g.huawei.com/videocloud/mediaprocesscenter/application-gateway/pkg/utils/clients"
	"codehub-g.huawei.com/videocloud/mediaprocesscenter/application-gateway/pkg/utils/hhmac"
	"codehub-g.huawei.com/videocloud/mediaprocesscenter/application-gateway/pkg/utils/log"
	"codehub-g.huawei.com/videocloud/mediaprocesscenter/application-gateway/pkg/utils/validator"
)

func init() {
	flag.StringVar(&config.GlobalConfig.GatewayAddr, "gateway-addrress", "0.0.0.0:9999", "application gateway address")
	flag.StringVar(&config.GlobalConfig.AassAddr, "aass-address", "", "aass service address")
	flag.StringVar(&config.GlobalConfig.DbUserName, "database-user-name", "", "database user name")
	flag.StringVar(&config.GlobalConfig.DbPassword, "database-password", "", "database password")
	flag.StringVar(&config.GlobalConfig.DbAddr, "database-address", "", "database address")
	flag.StringVar(&config.GlobalConfig.DbName, "database-name", "", "database name")

	flag.StringVar(&config.GlobalConfig.InfluxUsername, "influx-username", "", "influxdb user name")
	flag.StringVar(&config.GlobalConfig.InfluxPassword, "influx-password", "", "influxdb password")
	flag.StringVar(&config.GlobalConfig.InfluxAddr, "influx-address", "", "influxdb address")
	flag.StringVar(&config.GlobalConfig.InfluxDBName, "influx-dbname", "", "influxdb name")

	flag.StringVar(&config.GlobalConfig.GCMKey, "gcm-key", "", "gcm key")
	flag.StringVar(&config.GlobalConfig.GCMNonce, "gcm-nonce", "", "gcm nonce")

	flag.StringVar(&config.GlobalConfig.CleanStrategy, "clean-strategy", "on", "clean strategy")
	flag.StringVar(&config.GlobalConfig.LogLevel, "log-level", config.LogLevelInfo, "log level,support "+
		"debug and info")
	flag.IntVar(&config.GlobalConfig.LogRotateSize, "log-rotate-size", config.DefaultLogRotateSize, "log rotate size")
	flag.IntVar(&config.GlobalConfig.LogBackupCount, "log-backup-count", config.DefaultLogBackupCount, "log backup count")
	flag.IntVar(&config.GlobalConfig.LogMaxAge, "log-max-age", config.DefaultLogMaxAge, "log max days")
	flag.StringVar(&config.GlobalConfig.DeployModel, "deploy-model", config.DeployModelMultiInstance,
		"deploy model，support singleton and multi-instances")

}

// ReturnErr return when err is not nil
func ReturnErr(err error) {
	if err != nil {
		fmt.Printf("app gateway init failed for %v\n", err)
		os.Exit(1)
	}
}

func main() {
	flag.Parse()

	ReturnErr(log.InitLog())
	ReturnErr(models.InitDB())
	ReturnErr(validator.Init())
	ReturnErr(config.Init())
	ReturnErr(hhmac.InitHMACKey())

	// 实例任务
	task.InitInstanceTask()
	// 监控任务
	task.InitMonitorTask()
	// 接管任务
	task.InitTakeoverTask()
	// init metrics
	metrics.Init()
	// 启动server session dispatcher
	task.InitServerSessionDispatch()

	clients.InitAASSClient() // default timeout seconds

	// init beego web
	// set this so that we can get request body
	web.BConfig.CopyRequestBody = true

	controllers.Init()
	filters.InitFilters()
	routers.InitRouters()

	clean.Init()
	web.Run(config.GlobalConfig.GatewayAddr)
}
