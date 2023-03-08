// Copyright (c) Huawei Technologies Co., Ltd. 2022-2022. All rights reserved

// 任务结束
package task

import (
	"sync"
	"time"

	"codehub-g.huawei.com/videocloud/mediaprocesscenter/application-gateway/config"
	"codehub-g.huawei.com/videocloud/mediaprocesscenter/application-gateway/pkg/apis"
	"codehub-g.huawei.com/videocloud/mediaprocesscenter/application-gateway/pkg/common"
	"codehub-g.huawei.com/videocloud/mediaprocesscenter/application-gateway/pkg/distributedlock"
	"codehub-g.huawei.com/videocloud/mediaprocesscenter/application-gateway/pkg/models"
	app_process "codehub-g.huawei.com/videocloud/mediaprocesscenter/application-gateway/pkg/models/appprocess"
	client_session "codehub-g.huawei.com/videocloud/mediaprocesscenter/application-gateway/pkg/models/clientsession"
	"codehub-g.huawei.com/videocloud/mediaprocesscenter/application-gateway/pkg/models/lock"
	server_session "codehub-g.huawei.com/videocloud/mediaprocesscenter/application-gateway/pkg/models/serversession"
	"codehub-g.huawei.com/videocloud/mediaprocesscenter/application-gateway/pkg/services"
	"codehub-g.huawei.com/videocloud/mediaprocesscenter/application-gateway/pkg/utils/log"
)

const (
	defaultSleepTime             = 60 * time.Second
	appProcessStateCheckInterval = 90 * time.Second
	deadInstanceChLength		 = 10
	waitGroupAddNum				 = 2
)

func InitTakeoverTask() {
	w := &takeoverWorker{
		lockDao: lock.NewLockDao(models.MySqlOrm),
	}
	if config.GlobalConfig.DeployModel == config.DeployModelSingleton {
		go w.work()
		return
	}
	c := distributedlock.NewDistributedLockController(common.LockTakeover, common.LockBizCategory, w)
	c.Work()
}

type takeoverWorker struct {
	lockDao        *lock.Dao
	stopCh         chan struct{}
	deadInstanceCh chan string
}

func (w *takeoverWorker) HolderHook() {
	log.RunLogger.Infof("[takeover worker] start takeover worker")
	w.stopCh = make(chan struct{}, 0)
	w.deadInstanceCh = make(chan string, deadInstanceChLength)
	go w.work()
}

func (w *takeoverWorker) CompetitorHook() {
	log.RunLogger.Infof("[takeover worker] stop takeover worker")
	close(w.stopCh)
}

// 周期性去检查每个节点的锁的存活状态是mater判断为死掉之后，全部接管，然后执行，完成之后把对应的锁给删除了
func (w *takeoverWorker) work() {
	// 启动一个后台任务不断处理deadInstanceCh
	go w.takeOverInstanceTaskLoop()
	// 不断接手deadInstanceCh的内容然后进行处理
	ticker := time.NewTicker(20 * time.Second)
	for {
		select {
		case <-w.stopCh:
			ticker.Stop()
			log.RunLogger.Infof("[takeover worker] exit monitor worker")
			return
		case <-ticker.C:
			// 查询是否有dead instance，就是看是否有过期的key
			instanceLocks, err := w.lockDao.GetExpiredLocks(common.LockInstanceCategory)
			if err != nil {
				log.RunLogger.Errorf("[takeover worker] failed to get dead instance list")
				continue
			}
			for _, instanceLock := range instanceLocks {
				w.deadInstanceCh <- instanceLock.Name
			}
		}
	}
}

func (w *takeoverWorker) takeOverInstanceTaskLoop() {
	for {
		select {
		case <-w.stopCh:
			return
		case deadInstance := <-w.deadInstanceCh:
			go w.takeOverInstanceTask(deadInstance)
		}
	}
}

// 代理任务的执行，执行完成之后，把锁给删除了就可以了
func (w *takeoverWorker) takeOverInstanceTask(deadInstance string) {
	log.RunLogger.Infof("[takeover worker] %s instance start to take over %s instance task",
		config.GlobalConfig.InstanceName, deadInstance)
	var wg sync.WaitGroup
	// 代理执行任务
	wg.Add(waitGroupAddNum)
	go func() {
		defer wg.Done()
		restoreServerSessions(deadInstance)
	}()
	go func() {
		defer wg.Done()
		restoreClientSessions(deadInstance)
	}()
	wg.Wait()
	// 任务结束后，释放锁
	instanceLock := &lock.Lock{
		Category: common.LockInstanceCategory,
		Holder:   deadInstance,
		Name:     deadInstance,
	}
	err := w.lockDao.Release(instanceLock)
	if err != nil {
		log.RunLogger.Errorf("[takeover worker] failed to release dead instance %s lock", deadInstance)
		return
	}
	log.RunLogger.Infof("[takeover worker] success to take over %s instance task", deadInstance)

}

func restoreClientSessions(instance string) {
	log.RunLogger.Infof("[restorer] start to restore client sessions")
	clientSessionDao := client_session.NewClientSessionDao(models.MySqlOrm)
	serviceSessionDao := server_session.NewServerSessionDao(models.MySqlOrm)

	// 获取状态为RESERVED的client session, 如果超时，则重新拉起，否则
	css, err := clientSessionDao.ListReservedClientSessionForInstance(instance)
	if err != nil {
		log.RunLogger.Errorf("[restorer] failed to restore client session %v", err)
		return
	}

	for _, csDB := range css {
		go func(csDB client_session.ClientSession) {
			log.RunLogger.Infof("[restorer] start to active client session %s", csDB.ID)
			ssDB, err := serviceSessionDao.GetOneByID(csDB.ServerSessionID)
			if ssDB.State != common.ServerSessionStateActive {
				log.RunLogger.Infof("[restorer] failed to restorer client session " +
					"for serverSession state is not active")
				return
			}
			if err != nil {
				log.RunLogger.Errorf("[restorer] failed to query server session by id for %v", err)
				return
			}
			services.ListenClientSession(&csDB, log.RunLogger)
		}(csDB)
	}

}

// 对server session的状态做修复，包括
// 1. 对于已经过期的，状态为Activating的server session设置为Error
// 2. 对于没过期的，状态为Activating的server session，重新启动激活流程
func restoreServerSessions(instance string) {
	log.RunLogger.Infof("[restorer] start to restore server sessions ")

	serverSessionDao := server_session.NewServerSessionDao(models.MySqlOrm)
	appProcessDao := app_process.NewAppProcessDao(models.MySqlOrm)

	// 设置状态为activating，且已经超时的server session状态为ERROR
	err := models.TerminateOutOfDateServerSession(instance)
	if err != nil {
		log.RunLogger.Errorf("[restorer] failed to restore server session %v", err)
		return
	}

	// 获取状态为activating，且没超时的server session； 前面已经把超时的都处理了，剩下的都是没有超时的
	sss, err := serverSessionDao.QueryActivatingServerSessionByWorknode(instance)
	if err != nil {
		log.RunLogger.Errorf("[restorer] failed to restore server session %v", err)
		return
	}
	for _, ssDB := range sss {
		// 为每个server session下发请求到auxproxy, 并启动状态监控定时器
		// TODO 这里需要做数量控制
		go func(ssDB server_session.ServerSession) {
			log.RunLogger.Infof("[restorer] start to active server session %s", ssDB.ID)
			apDB, err := appProcessDao.GetAppProcessByID(ssDB.ProcessID)
			if err != nil {
				log.RunLogger.Errorf("[restorer] failed to query process by id for %v", err)
				return
			}

			ss := apis.TransferSSFromModel2Api(&ssDB)
			err = services.ActivateServerSession(apDB, &ssDB, ss, log.RunLogger)
			if err != nil {
				log.RunLogger.Errorf("[restorer] failed to active server session %s for %v", ssDB.ID, err)
				return
			}
			log.RunLogger.Infof("[restorer] success to start active server session %s", ssDB.ID)
		}(ssDB)
	}
}
