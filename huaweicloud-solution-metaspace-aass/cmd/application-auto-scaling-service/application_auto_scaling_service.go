// Copyright (c) Huawei Technologies Co., Ltd. 2022-2022. All rights reserved.

// aass启动入口
package main

import (
	"scase.io/application-auto-scaling-service/cmd/application-auto-scaling-service/app"
	"scase.io/application-auto-scaling-service/pkg/utils"
)

func main() {
	stopCh := utils.SetupSignalHandler()
	app.Init()
	app.Run(stopCh)
}
