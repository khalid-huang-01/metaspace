// Copyright (c) Huawei Technologies Co., Ltd. 2022-2022. All rights reserved.

// 工作节点
package worknode

import (
	"regexp"
	"time"

	"scase.io/application-auto-scaling-service/pkg/common"
	"scase.io/application-auto-scaling-service/pkg/db"
	"scase.io/application-auto-scaling-service/pkg/metricmonitor"
	"scase.io/application-auto-scaling-service/pkg/setting"
	"scase.io/application-auto-scaling-service/pkg/taskmgmt"
	"scase.io/application-auto-scaling-service/pkg/taskmgmt/reload"
	"scase.io/application-auto-scaling-service/pkg/utils"
	"scase.io/application-auto-scaling-service/pkg/utils/logger"
	"scase.io/application-auto-scaling-service/pkg/utils/wait"
)

const (
	formatDateTime    = "2006-01-02 15:04:05"
	timeFormatPattern = `^20[\d]{2}-\d{2}-\d{2} \d{2}:\d{2}:\d{2}$`
)

var (
	log *logger.FMLogger

	taskOverTaskIntervalSeconds  int
	heartBeatTaskIntervalSeconds int
	deadCheckTaskIntervalSeconds int
	maxDeadMinutes               int
)

// Init WorkNode
func Init() error {
	taskOverTaskIntervalSeconds = setting.GetWorkNodeTakeOverTaskIntervalSeconds()
	heartBeatTaskIntervalSeconds = setting.GetWorkNodeHeartBeatTaskIntervalSeconds()
	deadCheckTaskIntervalSeconds = setting.GetWorkNodeDeadCheckTaskIntervalSeconds()
	maxDeadMinutes = setting.GetWorkNodeTaskMaxDeadMinutes()
	log = logger.R.WithField(logger.WorkNodeId, common.LocalWorkNodeId)

	return db.InsertWorkNode(&db.WorkNode{
		Id: common.LocalWorkNodeId,
		IP: utils.GetLocalIP(),
	})
}

// StartWorkNodeTakeOverTask 周期性检测可执行的工作流任务并启动
func StartWorkNodeTakeOverPeriodTask(stopCh <-chan struct{}) {
	go wait.Until(func() {
		log.Info("TakeOverWorkNodeTask start, WorkNode: %s", common.LocalWorkNodeId)
		if err := takeOverWorkNode(); err != nil {
			log.Warn("work node take over error: %v", err)
			return
		}

		log.Debug("TakeOverWorkNodeTask finished, WorkNode: %s", common.LocalWorkNodeId)
	}, time.Duration(taskOverTaskIntervalSeconds)*time.Second, stopCh)
}

// StartWorkNodeHeartBeatPeriodTask 周期性心跳任务
func StartWorkNodeHeartBeatPeriodTask(stopCh <-chan struct{}) {
	go wait.Until(func() {
		log.Debug("HeartBeatTask start, WorkNode: %s ", common.LocalWorkNodeId)
		if err := heartBeat(); err != nil {
			log.Warn("heart beat error: %v", err)
			return
		}

		log.Debug("HeartBeatTask start, WorkNode: %s ", common.LocalWorkNodeId)
	}, time.Duration(heartBeatTaskIntervalSeconds)*time.Second, stopCh)

	go (func() {
		if err := stopWorkNode(stopCh); err != nil {
			log.Warn("stop work node error: %+v", err)
		}
	})()
}

// StartDeadWorkNodeUpdateTask 僵死WorkNode实例检测任务
func StartDeadWorkNodeUpdatePeriodTask(stopCh <-chan struct{}) {
	go wait.Until(func() {
		deadTime := time.Now().UTC().Add(-time.Minute * time.Duration(maxDeadMinutes))
		deadTimeString := deadTime.Format(formatDateTime)
		log.Info("DeadWorkNodeUpdateTask start, deadTime: %v, deadTime: %s", deadTime, deadTimeString)

		if !checkDeadTime(deadTimeString) {
			// NOTE: 本地测试过程中, 曾经意外发现转换出来的deadTimeString是乱码, 导致后续判断错误, 但是无法复现,
			// 		 所以在这里增加字符串格式校验, 防止该情况发生
			log.Warn("DeadWorkNodeUpdateTask check deadTime: %s error", deadTimeString)
			return
		}

		// 记录所有dead的work node
		logDeadWorkNode(deadTimeString)

		// 执行dead work node更新
		updated, err := deadWorkNodeUpdate(deadTimeString)
		if err != nil {
			log.Warn("update dead work node error: %+v", err)
			return
		}

		log.Debug("DeadWorkNodeUpdateTask finished, %d are updated", updated)
	}, time.Duration(deadCheckTaskIntervalSeconds)*time.Second, stopCh)
}

func checkDeadTime(deadTime string) bool {
	isOk, err := regexp.MatchString(timeFormatPattern, deadTime)
	if err != nil {
		log.Error("CheckDeadTime error: %v", err)
		return false
	}
	return isOk
}

// 任务接管
func takeOverWorkNode() error {
	// 获取所有状态是Error以及Deleting的WorkNode列表
	states := []string{db.WorkNodeStateTerminated, db.WorkNodeStateError}
	wns, err := db.GetWorkNodesByFilterStates(states)
	if err != nil {
		return err
	}

	log.Warn("%d work nodes need to be take over, wns: %s", len(wns), wns)

	// 尝试Take Over, 如果成功, 则启动workflow恢复, 如果失败则执行下一个WorkNode的Take Over, 直到成功一个
	for _, wn := range wns {
		updated, err := db.TakeOverWorkNode(wn.Id, wn.State, common.LocalWorkNodeId)
		if err != nil {
			log.Warn("work node[%s] try to take over wn[%s] error: %+v", common.LocalWorkNodeId, wn, err)
		} else {
			if updated != 1 {
				// 没有接管成功
				log.Warn("work node[%s] take over wn[%s] failed by other node", common.LocalWorkNodeId, wn)
				continue
			}

			// 接管成功，启动接管流程
			go (func() {
				if err := startTakingOverAsyncTaskAndMetricAlarms(wn.Id); err != nil {
					log.Warn("Taking over async task or metric monitor task error: %+v", err)
				}
			})()

			// 一个period仅接管一个WorkNode
			break
		}
	}

	return nil
}

func startTakingOverAsyncTaskAndMetricAlarms(workNodeId string) error {
	// 1. 查询出对应workNode所有未删除的异步任务与指标监控告警
	asyncTasks, err := db.GetAsyncTasksByWorkNodeId(workNodeId)
	if err != nil {
		return err
	}
	metricTasks, err := db.GetMetricMonitorTasksByWorkNodeId(workNodeId)
	if err != nil {
		return err
	}

	// 2. 开始接管异步任务与指标监控告警
	asyncTaskIds := make([]int, 0)
	for _, t := range asyncTasks {
		asyncTaskIds = append(asyncTaskIds, t.Id)
	}
	if err = db.TakeOverAsyncTasks(asyncTaskIds, common.LocalWorkNodeId); err != nil {
		return err
	}
	monitorTaskIds := make([]string, 0)
	for _, t := range metricTasks {
		monitorTaskIds = append(monitorTaskIds, t.Id)
	}
	if err = db.TakeOverMetricMonitorTasks(monitorTaskIds, common.LocalWorkNodeId); err != nil {
		return err
	}
	// 更新成功, 更新WorkNode状态
	if err := db.DeleteWorkNodeState(workNodeId); err != nil {
		// 打印日志, 不影响后续异步任务执行
		log.Warn("update work node state to deleted db err: %+v", err)
	}

	// 3. 重启异步任务与指标监控告警，理论上可以忽略错误。
	taskMgmt := taskmgmt.GetTaskMgmt()
	for _, at := range asyncTasks {
		if err = reload.AddAsyncTask(metricmonitor.GetMgmt(), taskMgmt, at); err != nil {
			log.Error("reload add async task[%s:%s] err: %+v", at.TaskKey, at.Id, err)
			continue
		}
	}
	for _, mt := range metricTasks {
		if err = metricmonitor.GetMgmt().AddTask(mt); err != nil {
			log.Error("reload add metric monitor task[%s] err: %+v", mt.Id, err)
			continue
		}
	}

	return nil
}

func deadWorkNodeUpdate(deadTime string) (int64, error) {
	return db.UpdateDeadWorkNodeState(deadTime)
}

func logDeadWorkNode(deadTime string) {
	wns, err := db.QueryDeadWorkNode(deadTime)
	if err != nil {
		log.Error("QueryDeadWorkNode db error: %v", err)
		return
	}

	log.Info("QueryDeadWorkNode work node: %s", wns)
}

func heartBeat() error {
	return db.HeartBeat(common.LocalWorkNodeId)
}

func stopWorkNode(stopCh <-chan struct{}) error {
	for {
		select {
		case <-stopCh:
			return db.TerminateWorkNode(common.LocalWorkNodeId)
		default:
		}
	}
}
