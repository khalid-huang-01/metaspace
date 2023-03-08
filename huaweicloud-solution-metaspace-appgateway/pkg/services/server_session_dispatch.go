package services

import (
	"sync"
	"time"

	"codehub-g.huawei.com/videocloud/mediaprocesscenter/application-gateway/pkg/apis"
	"codehub-g.huawei.com/videocloud/mediaprocesscenter/application-gateway/pkg/common"
	"codehub-g.huawei.com/videocloud/mediaprocesscenter/application-gateway/pkg/models"
	server_session "codehub-g.huawei.com/videocloud/mediaprocesscenter/application-gateway/pkg/models/serversession"
	"codehub-g.huawei.com/videocloud/mediaprocesscenter/application-gateway/pkg/services/stragegy"
	"codehub-g.huawei.com/videocloud/mediaprocesscenter/application-gateway/pkg/utils/log"
)

const (
	roundThreshold = 1000
)

type ServerSessionDispatcher struct {
	dispatcher *stragegy.BatchDispatch
	roundCount int32
	mu         sync.Mutex
}

func (d *ServerSessionDispatcher) Work(stopCh chan struct{}) {
	d.dispatcher = stragegy.NewBatchDispatch()
	d.roundCount = 0
	go d.work(stopCh)
}

func (d *ServerSessionDispatcher) work(stopCh chan struct{}) {
	for {
		select {
		case <-stopCh:
			d.dispatcher.Clear()
			return
		default:
			d.executeOneDispatchRound()
		}
	}
}

func (d *ServerSessionDispatcher) executeOneDispatchRound() {
	// 获取所有server session
	serverSessionDao := server_session.NewServerSessionDao(models.MySqlOrm)
	// 先到先处理，按创建时间升序获取
	sssDB, err := serverSessionDao.ListByState1(common.ServerSessionStateCreating, common.ASCSort, 0, 10000)
	if err != nil {
		log.RunLogger.Errorf("[dispatch] failed to get server sessions with "+ " for %v", err)
		return
	}

	log.RunLogger.Debugf("[dispatch] start to dispatch %d server session", len(*sssDB))
	wg := sync.WaitGroup{}
	wg.Add(len(*sssDB))
	for _, ssDB := range *sssDB {
		go d.dispatchOneProcess(ssDB, serverSessionDao, &wg)
	}
	wg.Wait()
	d.monitor()
	time.Sleep(1 * time.Second)
}

func (d *ServerSessionDispatcher) dispatchOneProcess(ssDB server_session.ServerSession, 
	serverSessionDao *server_session.ServerSessionDao, wg *sync.WaitGroup) {
	defer wg.Done()
	// 获取可用process
	dispatchProcess, err := d.dispatcher.Pick(ssDB.FleetID)
	if err != nil {
		log.RunLogger.Errorf("[dispatch] get available process by fleet id "+
			"%s for %s failed because %v", ssDB.FleetID, ssDB.ID, err)
		ssDB.State = common.ServerSessionStateError
		ssDB.StateReason = err.Error()
		_, err := serverSessionDao.Update(&ssDB)
		if err != nil {
			log.RunLogger.Errorf("[dispatch] failed to update error server session %s to db, for %v", ssDB.ID, err)
		}
		return
	}

	log.RunLogger.Infof("[dispatch] success pick process %v for server session %s in fleetID %s",
		dispatchProcess.AppProcess.ID, ssDB.ID, ssDB.FleetID)

	// 填充数据
	ssDB.ProcessID = dispatchProcess.AppProcess.ID
	ssDB.InstanceID = dispatchProcess.AppProcess.InstanceID
	ssDB.PublicIP = dispatchProcess.AppProcess.PublicIP
	ssDB.ClientPort = dispatchProcess.AppProcess.ClientPort
	ssDB.PID = dispatchProcess.AppProcess.PID
	ssDB.State = common.ServerSessionStateActivating
	ssDB.ProtectionPolicy = dispatchProcess.AppProcess.NewServerSessionProtectionPolicy
	ssDB.ProtectionTimeLimitMinutes = dispatchProcess.AppProcess.ServerSessionProtectionTimeLimitMinutes
	ssDB.ActivationTimeoutSeconds = dispatchProcess.AppProcess.ServerSessionActivationTimeoutSeconds

	// 执行事务
	err = models.DispatchServerSession2Process(&ssDB, dispatchProcess.AppProcess)
	if err != nil {
		log.RunLogger.Errorf("[dispatch] failed to dispatch server session %s to process %s in fleetID %s "+
			"because %v", ssDB.ID, dispatchProcess.AppProcess.ID, ssDB.FleetID, err)
		ssDB.State = common.ServerSessionStateError
		ssDB.StateReason = err.Error()
		_, err := serverSessionDao.Update(&ssDB)
		if err != nil {
			log.RunLogger.Errorf("[dispatch] failed to update error server session %s to db, for %v", ssDB.ID, err)
		}
		// 入库失败的实例倾向剔除
		d.dispatcher.FinishHandleDispatch(dispatchProcess)
		return
	}

	go func() {
		ss := apis.TransferSSFromModel2Api(&ssDB)
		err = ActivateServerSession(dispatchProcess.AppProcess, &ssDB, ss, log.RunLogger)
		if err != nil {
			log.RunLogger.Errorf("[dispatch] activate server session error %v", err)
			return
		}
	}()
	log.RunLogger.Infof("[dispatch] success dispatch server session %s to process %s in fleetID %s",
		ssDB.ID, dispatchProcess.AppProcess.ID, ssDB.FleetID)
	d.dispatcher.FinishHandleDispatch(dispatchProcess)
}

func (d *ServerSessionDispatcher) monitor() {
	d.mu.Lock()
	defer d.mu.Unlock()

	d.roundCount += 1
	// 每隔1000轮清理一次，避免内存泄露
	if d.roundCount == roundThreshold {
		log.RunLogger.Infof("start to clear dispatcher data and start new round")
		d.dispatcher.Clear()
		d.roundCount = 0
	}
}
