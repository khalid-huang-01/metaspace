// Copyright (c) Huawei Technologies Co., Ltd. 2022-2022. All rights reserved.

// main
package main

import (
	"fleetmanager/boot"
	"fleetmanager/utils"
)

var (
	setupSignalHandler = utils.SetupSignalHandler
	bootInit           = boot.Init
	bootRun            = boot.Run
)

func main() {
	stopCh := setupSignalHandler()
	bootInit()
	bootRun(stopCh)
}
