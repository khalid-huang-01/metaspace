// Copyright (c) Huawei Technologies Co., Ltd. 2022-2022. All rights reserved.

// 实例任务
package task

import (
	"time"

	"codehub-g.huawei.com/videocloud/mediaprocesscenter/application-gateway/config"
	"codehub-g.huawei.com/videocloud/mediaprocesscenter/application-gateway/pkg/common"
	"codehub-g.huawei.com/videocloud/mediaprocesscenter/application-gateway/pkg/distributedlock"
	"codehub-g.huawei.com/videocloud/mediaprocesscenter/application-gateway/pkg/utils/log"
)

func InitInstanceTask() {
	if config.GlobalConfig.DeployModel == config.DeployModelSingleton {
		return
	}
	w := &instanceWorker{}
	c := distributedlock.NewDistributedLockController(config.GlobalConfig.InstanceName, common.LockInstanceCategory, w)
	c.SetLockLeaseTime(30 * time.Second)          // 每次续期30秒
	c.SetTryLockOrLeaseInterval(20 * time.Second) // 每20秒续期/重试一次
	c.Work()
}

type instanceWorker struct {
}

func (w *instanceWorker) HolderHook() {
	log.RunLogger.Infof("[instance worker] start instance worker")
}

func (w *instanceWorker) CompetitorHook() {
	log.RunLogger.Infof("[instance worker] stop instance worker")
}
