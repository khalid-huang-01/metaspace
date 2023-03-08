// Copyright (c) Huawei Technologies Co., Ltd. 2022-2022. All rights reserved.

// 应用初始化
package app

import (
	"context"
	"fmt"
	"os"

	"scase.io/application-auto-scaling-service/pkg/api"
	"scase.io/application-auto-scaling-service/pkg/cloudresource"
	"scase.io/application-auto-scaling-service/pkg/db"
	"scase.io/application-auto-scaling-service/pkg/metricmonitor"
	"scase.io/application-auto-scaling-service/pkg/service"
	"scase.io/application-auto-scaling-service/pkg/setting"
	"scase.io/application-auto-scaling-service/pkg/taskmgmt"
	"scase.io/application-auto-scaling-service/pkg/utils/hhmac"
	"scase.io/application-auto-scaling-service/pkg/utils/logger"
	"scase.io/application-auto-scaling-service/pkg/worknode"
)

func exitErr(err error) {
	if err != nil {
		fmt.Printf("Application auto scaling service init failed for: %+v\n", err)
		os.Exit(1)
	}
}

// Init init application auto scaling service
func Init() {
	exitErr(logger.Init())
	exitErr(setting.Init())
	exitErr(hhmac.InitHMACKey())
	exitErr(db.Init())
	exitErr(worknode.Init())
	exitErr(cloudresource.InitOpSvcIamClient())
	exitErr(metricmonitor.Init())
	taskmgmt.RunAsyncTaskMgmt(context.Background())
	exitErr(api.Init())
	service.InitTask()
}

func Run(stopCh <-chan struct{}) {
	// 启动节点心跳任务
	worknode.StartWorkNodeHeartBeatPeriodTask(stopCh)

	// 启动Dead Node检测任务
	worknode.StartDeadWorkNodeUpdatePeriodTask(stopCh)

	// 启动工作流接管异步任务
	worknode.StartWorkNodeTakeOverPeriodTask(stopCh)

	// 启动API启动任务
	api.Run()
}
