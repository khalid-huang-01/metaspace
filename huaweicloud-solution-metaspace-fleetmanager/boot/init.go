// Copyright (c) Huawei Technologies Co., Ltd. 2022-2022. All rights reserved.

// fleetmanager初始化运行
package boot

import (
	"fleetmanager/api"
	"fleetmanager/client"
	"fleetmanager/db"
	"fleetmanager/logger"
	"fleetmanager/setting"
	"fleetmanager/worknode"
	"fmt"
	"os"
)

func exitErr(err error) {
	if err != nil {
		// TODO(wangjun): 打印堆栈
		fmt.Fprintf(os.Stderr, "%s\n", err.Error())
		os.Exit(1)
	}
}

// Init fleet manager启动初始化函数
func Init() {
	exitErr(logger.Init())
	exitErr(setting.Init())
	exitErr(db.Init())
	exitErr(client.Init())
	exitErr(worknode.Init())
	exitErr(api.Init())
}

// Run fleet manager运行函数
func Run(stopCh <-chan struct{}){
	// 启动节点心跳任务
	worknode.StartWorkNodeHeartBeatPeriodTask(stopCh)

	// 启动Dead Node检测任务
	worknode.StartDeadWorkNodeUpdatePeriodTask(stopCh)

	// 启动工作流接管异步任务
	worknode.StartWorkNodeTakeOverPeriodTask(stopCh)

	// 启动API启动任务
	api.Run()
}
