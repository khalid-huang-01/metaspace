// Copyright (c) Huawei Technologies Co., Ltd. 2022-2022. All rights reserved.

// 应用进程恢复
package restorer

import (
	"encoding/json"
	"sync"

	"github.com/mitchellh/go-ps"

	"codehub-g.huawei.com/videocloud/mediaprocesscenter/auxproxy/pkg/apis"
	"codehub-g.huawei.com/videocloud/mediaprocesscenter/auxproxy/pkg/common"
	"codehub-g.huawei.com/videocloud/mediaprocesscenter/auxproxy/pkg/configmanager"
	"codehub-g.huawei.com/videocloud/mediaprocesscenter/auxproxy/pkg/processmanager"
	"codehub-g.huawei.com/videocloud/mediaprocesscenter/auxproxy/pkg/utils/clients"
	"codehub-g.huawei.com/videocloud/mediaprocesscenter/auxproxy/pkg/utils/log"
)

// Work 启动restorer
func Work() {
	takeOverAppProcesses()
}

func takeOverAppProcesses() {
	// 从gateway获取本节点非terminated的节点
	resp, err := clients.GWClient.ListProcesses(configmanager.ConfMgr.Config.InstanceID)
	if err != nil {
		log.RunLogger.Errorf("[restorer] failed to fetch all processes for %s", err)
		return
	}

	// 识别error/activating/active的节点中已经是不存在的进程，加入到terminating中等待后续处理
	processes, tobeTerminate, tobeTakeOver := resp.AppProcesses, make([]apis.AppProcess, 0), make([]apis.AppProcess, 0)
	for _, process := range processes {
		if process.State == common.AppProcessStateTerminated {
			continue
		}
		if process.State == common.AppProcessStateTerminating {
			tobeTerminate = append(tobeTerminate, process)
			continue
		}

		// 必须pid和bizpid都存在才可以
		if process.PID == process.BizPID {
			pro, err := ps.FindProcess(process.PID)
			if err != nil || pro == nil {
				log.RunLogger.Errorf("[restore] failed to find process %d, err %s or no exist", process.PID, err)
				tobeTerminate = append(tobeTerminate, process)
				continue
			} else {
				tobeTakeOver = append(tobeTakeOver, process)
			}
		} else {
			pro, err := ps.FindProcess(process.PID)
			if err != nil || pro == nil {
				log.RunLogger.Errorf("[restore] failed to find process %d, err %s or no exist", process.PID, err)
				tobeTerminate = append(tobeTerminate, process)
				continue
			}
			pro, err = ps.FindProcess(process.BizPID)
			if err != nil || pro == nil {
				log.RunLogger.Errorf("[restore] failed to find process %d, err %s or no exist", process.BizPID, err)
				tobeTerminate = append(tobeTerminate, process)
				continue
			}
			tobeTakeOver = append(tobeTakeOver, process)
		}
	}

	var wg sync.WaitGroup
	wg.Add(2) // 2种类型的数据
	// 处理terminating状态的节点,状态不为terminated的设置状态为terminated
	go handleNoExistProcesses(tobeTerminate, &wg)

	// 重新纳管是存活的节点
	go handleExistProcesses(tobeTakeOver, &wg)
	wg.Wait()
}

func handleNoExistProcesses(processes []apis.AppProcess, wg *sync.WaitGroup) {
	defer wg.Done()

	if len(processes) == 0 {
		log.RunLogger.Infof("[restorer] there no any process need to terminated")
		return
	} else {
		log.RunLogger.Infof("[restorer] start to handle terminate process")
	}

	// 直接设置状态为Terminated
	var selfwg sync.WaitGroup
	selfwg.Add(len(processes))
	for _, pro := range processes {
		go func(pro apis.AppProcess) {
			defer selfwg.Done()

			r := &apis.UpdateAppProcessStateRequest{
				State: common.AppProcessStateTerminated,
			}
			_, err := clients.GWClient.UpdateProcessState(pro.ID, r)
			if err != nil {
				log.RunLogger.Errorf("[restorer] failed to update process %s state to %s for "+
					"%v", pro.ID, common.AppProcessStateTerminated, err)
				return
			}
			log.RunLogger.Infof("[restorer] suceed to update process %s to state %s ",
				pro.ID, common.AppProcessStateTerminated)
		}(pro)
	}
	selfwg.Wait()
}

// 1. 加入pm管理(pm针对这种游离的需要每10s探测一次，使用一个独立的协程就可以了)
// 2. 手动注册到health checker
func handleExistProcesses(processes []apis.AppProcess, wg *sync.WaitGroup) {
	defer wg.Done()

	if len(processes) == 0 {
		log.RunLogger.Infof("[restorer] there no any process need to take over")
		return
	} else {
		log.RunLogger.Infof("[restorer] start to handle take over process")
	}

	var selfwg sync.WaitGroup
	selfwg.Add(len(processes))
	for _, pro := range processes {
		go func(pro apis.AppProcess) {
			defer selfwg.Done()
			var logPath []string
			err := json.Unmarshal([]byte(pro.LogPath), &logPath)
			if err != nil {
				log.RunLogger.Errorf("[restorer] failed to unmarshal logpath")
				return
			}
			takeOverProcess := processmanager.NewProcess(pro.LaunchPath, pro.Parameters, pro.PID)
			takeOverProcess.BizPid = pro.BizPID
			takeOverProcess.Id = pro.ID
			takeOverProcess.Ip = pro.PublicIP
			takeOverProcess.ClientPort = pro.ClientPort
			takeOverProcess.GrpcPort = pro.GrpcPort
			takeOverProcess.Status = pro.State
			takeOverProcess.LogPath = logPath
			err = processmanager.ProcessMgr.TakeOverRunningProcess(takeOverProcess)
			if err != nil {
				log.RunLogger.Errorf("[restorer] failed to create process to health checker for %v", err)
				return
			}
			log.RunLogger.Infof("[restoer] success handle existed processes %s", pro.ID)
		}(pro)
	}
}
