package task

import (
	"codehub-g.huawei.com/videocloud/mediaprocesscenter/application-gateway/pkg/common"
	"codehub-g.huawei.com/videocloud/mediaprocesscenter/application-gateway/pkg/distributedlock"
	"codehub-g.huawei.com/videocloud/mediaprocesscenter/application-gateway/pkg/services"
	"codehub-g.huawei.com/videocloud/mediaprocesscenter/application-gateway/pkg/utils/log"
	"time"
)

func InitServerSessionDispatch() {
	w := &serverSessionDispatchWorker{}
	m := distributedlock.NewDistributedLockController(common.LockServerSessionDispatch, common.LockBizCategory, w)
	m.SetLockLeaseTime(15 * time.Second)          // 每次续期15s
	m.SetTryLockOrLeaseInterval(10 * time.Second) // 每个10秒续期/尝试获取锁一次
	m.Work()
}

type serverSessionDispatchWorker struct {
	dispatcher *services.ServerSessionDispatcher
	stopCh     chan struct{}
}

func (w *serverSessionDispatchWorker) HolderHook() {
	log.RunLogger.Infof("[ss dispatch worker] start server session dispatch worker")
	w.stopCh = make(chan struct{}, 0)
	w.dispatcher = &services.ServerSessionDispatcher{}
	w.dispatcher.Work(w.stopCh)
}

func (w *serverSessionDispatchWorker) CompetitorHook() {
	log.RunLogger.Infof("[ss dispatch worker] stop server session dispatch worker")
	close(w.stopCh)
}
