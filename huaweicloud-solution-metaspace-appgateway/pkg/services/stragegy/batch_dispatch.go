package stragegy

import (
	"codehub-g.huawei.com/videocloud/mediaprocesscenter/application-gateway/pkg/models"
	app_process "codehub-g.huawei.com/videocloud/mediaprocesscenter/application-gateway/pkg/models/appprocess"
	"codehub-g.huawei.com/videocloud/mediaprocesscenter/application-gateway/pkg/utils/log"
	"fmt"
	"sync"
)

type Process struct {
	//如果在处理中的process就不覆盖（分配出去还没入库，又被捞取回来，避免被重新分配出去，导致后者失败,map的获取是不定序的）；如果没在处理中的，可以覆盖（需要考虑server session termianted的释放问题）
	HandlingCount int64 // 当前处理者数目
	AppProcess    *app_process.AppProcess
}

type BatchDispatch struct {
	// fleetID -> process
	fleetProcessesMap map[string]map[string]*Process
	fleetLockMap      map[string]*sync.Mutex
	mu                sync.Mutex
}

func NewBatchDispatch() *BatchDispatch {
	return &BatchDispatch{
		fleetLockMap:      make(map[string]*sync.Mutex),
		fleetProcessesMap: make(map[string]map[string]*Process),
	}
}

func (b *BatchDispatch) Pick(fleetID string) (*Process, error) {
	ap, err := b.pickup(fleetID)
	if err != nil {
		return nil, err
	}

	return ap, nil
}

// Clear 清理资源
func (b *BatchDispatch) Clear() {
	if len(b.fleetProcessesMap) == 0 {
		return
	}
	b.fleetProcessesMap = make(map[string]map[string]*Process)
	b.fleetLockMap = make(map[string]*sync.Mutex)
}

func (b *BatchDispatch) pickup(fleetID string) (*Process, error) {
	b.mu.Lock()
	if b.fleetProcessesMap[fleetID] == nil {
		if b.fleetProcessesMap[fleetID] == nil {
			b.fleetLockMap[fleetID] = &sync.Mutex{}
			b.fleetProcessesMap[fleetID] = make(map[string]*Process, 0)
		}
	}
	b.mu.Unlock()

	b.fleetLockMap[fleetID].Lock()
	defer b.fleetLockMap[fleetID].Unlock()

	ap := b.fetchOneAvailableAppProcess(fleetID)
	if ap == nil {
		log.RunLogger.Infof("[batch dispatch] fleet %s's process list is empty, start to fetch from db", fleetID)
		appProcessDao := app_process.NewAppProcessDao(models.MySqlOrm)
		processesDB, err := appProcessDao.GetAllAvailableAppProcessByFleetID(fleetID)
		if err != nil {
			log.RunLogger.Errorf("[batch dispatch] failed to fetch all available process by fleetID %s for %v",
				fleetID, err)
			return nil, err
		}

		for _, process := range processesDB {
			if _, ok := b.fleetProcessesMap[fleetID][process.ID]; !ok {
				// 不存在的话，直接入
				b.fleetProcessesMap[fleetID][process.ID] = &Process{
					HandlingCount: 0,
					AppProcess:    process,
				}
			} else {
				// 如果没有已经在处理的程序,就以数据库的为主；如果有在处理的，就不覆盖，避免从数据库捞取回来已经分配出去但还是还没入库的
				if b.fleetProcessesMap[fleetID][process.ID].HandlingCount == 0 {
					b.fleetProcessesMap[fleetID][process.ID] = &Process{
						HandlingCount: 0,
						AppProcess:    process,
					}
				}
			}
		}
		ap = b.fetchOneAvailableAppProcess(fleetID)
	}

	if ap == nil {
		return nil, fmt.Errorf("there is no available process for fleet %s", fleetID)
	}
	ap.AppProcess.ServerSessionCount += 1
	ap.HandlingCount += 1

	return ap, nil
}

// 这里要保证这里的数目无论是process还是fleet的维度一定会比外寸的少，这样才不会实际存在但是说没有
func (b *BatchDispatch) fetchOneAvailableAppProcess(fleetID string) *Process {
	if len(b.fleetProcessesMap[fleetID]) == 0 {
		return nil
	}
	for _, ap := range b.fleetProcessesMap[fleetID] {
		if ap.AppProcess.ServerSessionCount < ap.AppProcess.MaxServerSessionNum {
			return ap
		}
	}
	return nil
}

// 不区分最后是否入库成功来进行AppProcess.ServerSessionCount -= 1是因为想以数据库为主；及时把正确的数目刷回来
// 入库失败的实例倾向剔除
func (b *BatchDispatch) FinishHandleDispatch(ap *Process) {
	b.fleetLockMap[ap.AppProcess.FleetID].Lock()
	defer b.fleetLockMap[ap.AppProcess.FleetID].Unlock()
	ap.HandlingCount -= 1

	// 判断是否需要移除
	if ap.HandlingCount == 0 && ap.AppProcess.ServerSessionCount >= ap.AppProcess.MaxServerSessionNum {
		log.RunLogger.Infof("[batch dispatch] delete process %s from fleet %s process list", ap.AppProcess.ID,
			ap.AppProcess.FleetID)
		delete(b.fleetProcessesMap[ap.AppProcess.FleetID], ap.AppProcess.ID)
	}
}
