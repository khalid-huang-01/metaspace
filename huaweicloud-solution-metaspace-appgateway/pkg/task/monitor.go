// Copyright (c) Huawei Technologies Co., Ltd. 2022-2022. All rights reserved.

// 监控任务
package task

import (
	"time"

	"codehub-g.huawei.com/videocloud/mediaprocesscenter/application-gateway/config"
	"codehub-g.huawei.com/videocloud/mediaprocesscenter/application-gateway/pkg/common"
	"codehub-g.huawei.com/videocloud/mediaprocesscenter/application-gateway/pkg/distributedlock"
	"codehub-g.huawei.com/videocloud/mediaprocesscenter/application-gateway/pkg/models"
	app_process "codehub-g.huawei.com/videocloud/mediaprocesscenter/application-gateway/pkg/models/appprocess"
	"codehub-g.huawei.com/videocloud/mediaprocesscenter/application-gateway/pkg/utils/log"
)

func InitMonitorTask() {
	w := &monitorWorker{}
	if config.GlobalConfig.DeployModel == config.DeployModelSingleton {
		go w.monitorAppProcesses()
		return
	}

	c := distributedlock.NewDistributedLockController(common.LockMonitor, common.LockBizCategory, w)
	c.Work()
}

type monitorWorker struct {
	stopCh chan struct{}
}

func (w *monitorWorker) HolderHook() {
	log.RunLogger.Infof("[monitor worker] start monitor worker")
	w.stopCh = make(chan struct{}, 0)
	go w.monitorAppProcesses()
}

func (w *monitorWorker) CompetitorHook() {
	log.RunLogger.Infof("[monitor worker] stop monitor worker")
	close(w.stopCh)
}

// 1. 启动一个任务，静默一分钟，然后对于三个周期都没有接收到状态上报的process，设置为Error
func (w *monitorWorker) monitorAppProcesses() {
	log.RunLogger.Infof("[monitor worker] start to restore app process state monitorWorker")
	time.Sleep(defaultSleepTime)

	appProcessDao := app_process.NewAppProcessDao(models.MySqlOrm)

	ticker := time.NewTicker(appProcessStateCheckInterval)

	for {
		select {
		case <-w.stopCh:
			log.RunLogger.Infof("[monitor worker] exit monitor worker")
			return
		case <-ticker.C:
			log.RunLogger.Infof("[monitor worker] start to check and update process state")
			err := appProcessDao.VerifyZombieProcess()
			if err != nil {
				log.RunLogger.Infof("[monitor worker] verify zombie process failed %v", err)
			}
		}
	}
}
