// Copyright (c) Huawei Technologies Co., Ltd. 2022-2022. All rights reserved.

// worknode
package worknode

import (
	"fleetmanager/db/dao"
	"fleetmanager/logger"
	"fleetmanager/utils/wait"
	"fleetmanager/workflow"
	"github.com/google/uuid"
	"regexp"
	"time"
)

const (
	DefaultTakeOverTaskInterval          = 5 * 60
	DefaultDeadWorkNodeCheckTaskInterval = 5 * 60
	DefaultHeartBeatTaskInterval         = 1 * 60
	MaxDeadMinites                       = 10
	FormatDateTime                       = "2006-01-02 15:04:05"
	TimeFormatPattern                    = `^20[\d]{2}-\d{2}-\d{2} \d{2}:\d{2}:\d{2}$`
)

// WorkNodeId is unified id for one work node
var WorkNodeId string
var TLogger *logger.FMLogger

// Init WorkNode
func Init() error {
	u, _ := uuid.NewUUID()
	WorkNodeId = u.String()
	wn := &dao.WorkNode{
		Id:    WorkNodeId,
		State: dao.WorkNodeStateRunning,
	}
	TLogger = logger.R.WithField(logger.WorkNodeId, WorkNodeId)
	return dao.InsertWorkNode(wn)
}

// StartWorkNodeTakeOverPeriodTask 周期性检测可执行的工作流任务并启动
func StartWorkNodeTakeOverPeriodTask(stopCh <-chan struct{}) {
	go wait.Until(func() {
		TLogger.Info("TakeOverWorkNodeTask start, WorkNode: %s", WorkNodeId)
		if err := takeOverWorkNode(); err != nil {
			logger.R.Warn("work node take over error: %v", err)
			return
		}

		TLogger.Debug("TakeOverWorkNodeTask finished, WorkNode: %s", WorkNodeId)
	}, time.Duration(DefaultTakeOverTaskInterval)*time.Second, stopCh)
}

// StartWorkNodeHeartBeatPeriodTask 周期性心跳任务
func StartWorkNodeHeartBeatPeriodTask(stopCh <-chan struct{}) {
	go wait.Until(func() {
		TLogger.Debug("HeartBeatTask start, WorkNode: %s ", WorkNodeId)
		if err := heartBeat(); err != nil {
			logger.R.Warn("heart beat error: %v", err)
			return
		}

		TLogger.Debug("HeartBeatTask start, WorkNode: %s ", WorkNodeId)
	}, time.Duration(DefaultHeartBeatTaskInterval)*time.Second, stopCh)

	go (func() {
		if err := stopWorkNode(stopCh); err != nil {
			TLogger.Warn("stop work node error: %v", err)
		}
	})()
}

// StartDeadWorkNodeUpdateTask 僵死WorkNode实例检测任务
func StartDeadWorkNodeUpdatePeriodTask(stopCh <-chan struct{}) {
	go wait.Until(func() {
		deadTime := time.Now().UTC().Add(-time.Minute * MaxDeadMinites)
		deadTimeString := deadTime.Format(FormatDateTime)
		TLogger.Info("DeadWorkNodeUpdateTask start, deadTime: %v, deadTime: %s", deadTime, deadTimeString)

		if !checkDeadTime(deadTimeString) {
			// NOTE: 本地测试过程中, 曾经意外发现转换出来的deadTimeString是乱码, 导致后续判断错误, 但是无法复现,
			// 		 所以在这里增加字符串格式校验, 防止该情况发生
			TLogger.Warn("DeadWorkNodeUpdateTask check deadTime: %s error", deadTimeString)
			return
		}

		// 记录所有dead的work node
		logDeadWorkNode(deadTimeString)

		// 执行dead work node更新
		updated, err := deadWorkNodeUpdate(deadTimeString)
		if err != nil {
			TLogger.Warn("update dead work node error: %v", err)
			return
		}

		TLogger.Debug("DeadWorkNodeUpdateTask finished, %d are updated", updated)
	}, time.Duration(DefaultDeadWorkNodeCheckTaskInterval)*time.Second, stopCh)
}

func checkDeadTime(deadTime string) bool {
	isOk, err := regexp.MatchString(TimeFormatPattern, deadTime)
	if err != nil {
		TLogger.Error("CheckDeadTime error: %v", err)
		return false
	}
	return isOk
}

// 任务接管
func takeOverWorkNode() error {
	// 获取所有状态是Error以及Terminated的WorkNode列表
	f := dao.Filters{"State__in": []string{dao.WorkNodeStateTerminated, dao.WorkNodeStateError}}
	wns, err := dao.GetWorkNodes(f)
	if err != nil {
		return err
	}

	TLogger.Warn("%d work nodes need to be take over, wns: %s", len(wns), wns)

	// 尝试Take Over, 如果成功, 则启动workflow恢复, 如果失败则执行下一个WorkNode的Take Over, 直到成功一个
	for _, wn := range wns {
		updated, err := dao.TakeOverWorkNode(wn.Id, wn.State, WorkNodeId)
		if err != nil {
			TLogger.Warn("work node %s try to take over wn %s error: %v", WorkNodeId, wn, err)
		} else {
			if updated != 1 {
				// 没有接管成功
				TLogger.Warn("work node %s take over wn %s failed by other node", WorkNodeId, wn)
				continue
			}

			// 接管成功，启动接管流程
			TLogger.Debug("Try to take over wfs, WorkNode:%s", WorkNodeId)
			go (func() {
				if err := startTakeOverWorkFlow(wn.Id); err != nil {
					logger.R.Warn("workflow take over error: %v", err)
				}
			})()

			// 一个period仅接管一个WorkNode
			break
		}
	}

	return nil
}

func updateWorkFlowNodeInfo(wfs []dao.Workflow, originWorkNodeId string) error {
	wfIds := getWorkflowIds(wfs)
	if err := dao.TakeOverWorkFlow(wfIds, WorkNodeId); err != nil {
		// 更新失败
		return err
	}

	// 更新成功, 更新worknode状态
	if err := dao.UpdateWorkNodeState(originWorkNodeId, dao.WorkNodeStateFinished); err != nil {
		// 打印日志, 不影响后续工作流任务执行
		TLogger.Warn("update work node state to finished db error: %v", err)
	}
	return nil
}

func getWorkflowIds(wfs []dao.Workflow) []string {
	wfIds := make([]string, 0)
	for _, wf := range wfs {
		wfIds = append(wfIds, wf.Id)
	}
	return wfIds
}

func startTakeOverWorkFlow(workNodeId string) error {
	// 1. 查询出对应workNode所有的未完成工作流
	filter := dao.Filters{
		"State__in":  []string{dao.WorkflowStateRollbacking, dao.WorkflowStateRunning, dao.WorkflowStateCreate},
		"WorkNodeId": workNodeId,
	}
	wfs, err := dao.GetAllWorkflows(filter)
	if err != nil {
		return err
	}

	TLogger.Info("Try to take over wfs:%+v, WorkNode:%s", wfs, WorkNodeId)

	// 2. 开始接管
	if err := updateWorkFlowNodeInfo(wfs, workNodeId); err != nil {
		return err
	}

	// 3. 执行工作流, 理论上可以忽略错误
	for _, wf := range wfs {
		tmp, err := workflow.LoadWorkflow(wf.Id)
		if err != nil {
			TLogger.Error("load workflow error: %v, try to update to error", err)
			if err = workflow.StartWorkflowFailed(&wf); err != nil {
				TLogger.Error("change workflow to error failed error: %v, try to ignore", err)
			}
			continue
		}

		// 启动工作流
		tmp.Run()
	}

	return nil
}

func deadWorkNodeUpdate(deadTime string) (int64, error) {
	return dao.UpdateDeadWorkNodeState(deadTime)
}

func logDeadWorkNode(deadTime string) {
	wns, err := dao.QueryDeadWorkNode(deadTime)
	if err != nil {
		TLogger.Error("QueryDeadWorkNode db error: %v", err)
		return
	}

	TLogger.Info("QueryDeadWorkNode work node: %s", wns)
}

func heartBeat() error {
	return dao.HeartBeat(WorkNodeId)
}

func stopWorkNode(stopCh <-chan struct{}) error {
	for {
		select {
		case <-stopCh:
			return dao.TerminateWorkNode(WorkNodeId)
		}
	}
}
