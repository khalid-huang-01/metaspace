// Copyright (c) Huawei Technologies Co., Ltd. 2022-2022. All rights reserved.

// 进程清理
package httpserver

import (
	context2 "context"
	"net/http"
	"sync"
	"time"

	"github.com/beego/beego/v2/server/web/context"

	"codehub-g.huawei.com/videocloud/mediaprocesscenter/auxproxy/pkg/apis"
	"codehub-g.huawei.com/videocloud/mediaprocesscenter/auxproxy/pkg/common"
	"codehub-g.huawei.com/videocloud/mediaprocesscenter/auxproxy/pkg/configmanager"
	"codehub-g.huawei.com/videocloud/mediaprocesscenter/auxproxy/pkg/processmanager"
	"codehub-g.huawei.com/videocloud/mediaprocesscenter/auxproxy/pkg/utils/clients"
	"codehub-g.huawei.com/videocloud/mediaprocesscenter/auxproxy/pkg/utils/log"
	"codehub-g.huawei.com/videocloud/mediaprocesscenter/auxproxy/pkg/utils/sdk/processservice"
)

const (
	CleanUpStateReserved = "RESERVED"
	CleanUpStateCleaning = "CLEANING"
	CleanUpStateFinished = "FINISHED"
)

var state = CleanUpStateReserved

// StartCleanUp start cleanup
func StartCleanUp(ctx *context.Context) {
	if state == CleanUpStateCleaning || state == CleanUpStateFinished {
		ctx.Output.SetStatus(http.StatusAccepted)
		return
	}

	log.RunLogger.Infof("[cleaner] start to clean up resources")

	go func() {
		// change state to cleaning
		state = CleanUpStateCleaning

		// stop config manager
		configmanager.ConfMgr.Stop()
		log.RunLogger.Infof("[cleaner] stop config manager")

		// stop health checker
		processmanager.ProcessMgr.Stop()
		log.RunLogger.Infof("[cleaner] stop process manager")

		// update all processes state to terminating
		pros := processmanager.ProcessMgr.GetAllRunningProcesses()
		for _, pro := range pros {
			r := &apis.UpdateAppProcessStateRequest{State: common.AppProcessStateTerminating}

			_, err := clients.GWClient.UpdateProcessState(pro.Id, r)
			if err != nil {
				log.RunLogger.Errorf("[cleaner] failed to update process %d state for %v", pro.Pid, err)
			} else {
				log.RunLogger.Infof("[cleaner] succeed to update process %d state to terminating", pro.Pid)
			}
		}

		// invoke OnProcessTerminate func to all processes
		defaultTerminationTime := 60
		for _, pro := range pros {
			_, err := pro.Client.OnProcessTerminate(context2.Background(),
				&processservice.ProcessTerminateRequest{TerminationTime: int64(defaultTerminationTime)})
			if err != nil {
				log.RunLogger.Errorf("[cleaner] failed to invoke on process %d terminate to process for %v",
					pro.Pid, err)
			} else {
				log.RunLogger.Infof("[cleaner] succeed to invoke onProcessTerminate to process %d", pro.Pid)
			}
		}

		// 处理所有的进程
		log.RunLogger.Infof("[cleaner] start to handler all processes with len %d", len(pros))
		var wg sync.WaitGroup
		wg.Add(len(pros))
		for _, pro := range pros {
			// TODO: 是否要设置超时
			go handleProcess(pro, &wg)
		}

		wg.Wait()
		log.RunLogger.Infof("[cleaner] success to finish all cleanup job")

		// change state to finished
		state = CleanUpStateFinished
	}()

	ctx.Output.SetStatus(http.StatusOK)
	return
}

func handleProcess(pro *processmanager.Process, wg *sync.WaitGroup) {
	ctx, cancel := context2.WithCancel(context2.Background())
	go _handlerProcess(pro, cancel)
	select {
	case _, _ = <-ctx.Done():
		// 处理完成所有的server session和client session，设置进程为Terminated（可能已经成功了）
		r := &apis.UpdateAppProcessStateRequest{State: common.AppProcessStateTerminated}
		_, err := clients.GWClient.UpdateProcessState(pro.Id, r)
		if err != nil {
			log.RunLogger.Errorf("[cleaner] failed to update process %d state for %v", pro.Pid, err)
		} else {
			log.RunLogger.Infof("[cleaner] succeed to update process %d state to terminating", pro.Pid)
		}
		wg.Done()
	}
}

func _handlerProcess(process *processmanager.Process, cancelFunc context2.CancelFunc) {
	log.RunLogger.Infof("[cleaner] start to handler terminate process %s", process.Id)
	// 处理完毕所有的server session
	defer cancelFunc()

	// 获取这个pro的全部server session
	res, _ := clients.GWClient.ListServerSessions(process.Id)

	sss := res.ServerSessions
	var wg sync.WaitGroup
	wg.Add(len(sss))
	for _, ss := range sss {
		if ss.State == common.AppProcessStateError || ss.State == common.ServerSessionStateTerminated {
			wg.Done()
			continue
		}
		// 根据 server session的策略，做调用
		switch ss.ProtectionPolicy {
		case common.ProtectionPolicyNoProtection:
			go handlerServerSessionForNoProtection(ss.ID, &wg)

		case common.ProtectionPolicyTimeLimitProtection:
			go handlerServerSessionForTimeLimitProtection(ss.ID, ss.ProtectionTimeLimitMinutes, &wg)

		default:
			log.RunLogger.Infof("[cleaner] invalid ProtectionPolicy %s", ss.ProtectionPolicy)
			wg.Done()
		}
	}
	// 等待
	wg.Wait()
	log.RunLogger.Infof("[cleaner] success to handler terminate process %s", process.Id)
}

func handlerServerSessionForNoProtection(id string, wg *sync.WaitGroup) {
	log.RunLogger.Infof("[cleaner] start to handler server session %s with no protection", id)
	defer wg.Done()

	err := clients.GWClient.TerminateServerSessionAllRelativeResources(id)
	if err != nil {
		log.RunLogger.Errorf("[cleaner] handler server session for no protection failed %v", err)
		return
	}
	log.RunLogger.Infof("[cleaner] success to handler server session %s with no protection", id)
}

func handlerServerSessionForTimeLimitProtection(id string, limitTime int, wg *sync.WaitGroup) {
	log.RunLogger.Infof("[cleaner] start to handler server session %s with time limit protection", id)
	// 设置定时器
	t := time.NewTimer(time.Duration(limitTime) * time.Minute)
	stopCh := make(chan struct{})
	go func() {
		defer wg.Done()
		select {
		case <-t.C:
			// 超时了，直接关闭
			log.RunLogger.Infof("[cleaner] handler server session %s for time limit protection is timeout, "+
				"start to termiate directly", id)
			err := clients.GWClient.TerminateServerSessionAllRelativeResources(id)
			if err != nil {
				log.RunLogger.Errorf("[cleaner] handler server session for timeLimitProtection failed %v", err)
			}
		case _, _ = <-stopCh:
			t.Stop()
		}
	}()

	// 周期性监控
	ticker := time.NewTicker(30 * time.Second)
	for {
		<-ticker.C
		res, err := clients.GWClient.FetchServerSessionAllRelativeResources(id)
		if err != nil {
			log.RunLogger.Errorf("[cleaner] handler server session for timeLimitProtection failed%v", err)
			continue
		}

		if res.ServerSession.State != common.ServerSessionStateError && res.ServerSession.
			State != common.ServerSessionStateTerminated {
			continue
		}

		isContinue := false
		for _, cs := range res.ClientSessions {
			if cs.State != common.ClientSessionStateCompleted &&
				cs.State != common.ClientSessionStateTimeout {
				isContinue = true
				break
			}
		}
		if isContinue {
			continue
		}
		ticker.Stop()
		break
	}
	// 通知关闭定时器
	stopCh <- struct{}{}

	log.RunLogger.Infof("[cleaner] success to handler server session %s with time limit protection", id)
}

// ShowCleanUpState 暴露清理进度
func ShowCleanUpState(ctx *context.Context) {
	ctx.JSONResp(ShowCleanUpStateResponse{State: state})
	return
}
