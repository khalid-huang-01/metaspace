// Copyright (c) Huawei Technologies Co., Ltd. 2022-2022. All rights reserved.

// 异步任务管理
package taskmgmt

import (
	"context"
	"fmt"
	"runtime"
	"time"

	"scase.io/application-auto-scaling-service/pkg/taskmgmt/asynctask/interfaces"
	"scase.io/application-auto-scaling-service/pkg/utils/logger"
)

const (
	// chanMaxLength chan缓存最大长度
	chanMaxLength = 1024
	// taskRetryInterval 任务重试时间间隔
	taskRetryInterval = time.Second * 30
)

var (
	asyncTaskMgmt *AsyncTaskMgmt
)

type AsyncTaskMgmt struct {
	// 正在运行的任务记录
	*taskCache
	// 接收新增异步任务chan
	addTaskChan chan interfaces.AsyncTaskInf
}

// GetTaskMgmt ...
func GetTaskMgmt() *AsyncTaskMgmt {
	if asyncTaskMgmt == nil {
		logger.R.Error("AsyncTaskMgmt is not initialized correctly")
		return nil
	}
	return asyncTaskMgmt
}

// AddTask 将异步任务交由 AsyncTaskMgmt 接管
func (m *AsyncTaskMgmt) AddTask(task interfaces.AsyncTaskInf) {
	asyncTaskMgmt.addTaskChan <- task
}

// RunAsyncTaskMgmt ...
func RunAsyncTaskMgmt(ctx context.Context) {
	asyncTaskMgmt = &AsyncTaskMgmt{
		taskCache:   newTaskCache(),
		addTaskChan: make(chan interfaces.AsyncTaskInf, chanMaxLength),
	}
	go asyncTaskMgmt.run(ctx)
}

func (m *AsyncTaskMgmt) run(ctx context.Context) {
	for {
		select {
		// 接收发布的异步任务
		case task, ok := <-m.addTaskChan:
			if !ok {
				return
			}
			if m.IsTaskExist(task.GetType(), task.GetKey()) {
				logger.R.Info("Async task[%s:%s] is running, do not need to be added",
					task.GetType(), task.GetKey())
				continue
			}
			logger.R.Info("Add async task[%s:%s]", task.GetType(), task.GetKey())
			go m.handleTask(task)
		case <-ctx.Done():
			logger.R.Info("=== AsyncTaskMgmt.run exit ===")
			return
		}
	}
}

// handleTask 处理任务的启动和失败重启
func (m *AsyncTaskMgmt) handleTask(task interfaces.AsyncTaskInf) {
	m.addTask(task)
	defer m.delTask(task)

	// 记录panic错误
	log := logger.R
	defer func() {
		if err := recover(); err != nil {
			var stack string
			for i := 1; ; i++ {
				_, file, line, ok := runtime.Caller(i)
				if !ok {
					break
				}
				stack += fmt.Sprintf("\n %s:%d", file, line)
			}
			log.Error("Recover panic err: %+v;\n Stack info: %s", err, stack)
		}
	}()

	// 执行任务，直至任务成功
	for !task.IsComplete() {
		log = logger.R.WithFields(map[string]interface{}{
			logger.AsyncTask:      fmt.Sprintf("%s[%s]", task.GetType(), task.GetKey()),
			logger.TaskRetryTimes: task.GetRetryTimes(),
			logger.TaskLastError:  task.GetLastErr(),
		})
		log.Info("Task starts running……")
		err := task.Run(log)
		if err != nil {
			task.SetStatusFailed(err)
			log.Error("The task exits in error and will be restarted later, err info: %+v", err)
			// 重启等待时长 = 30s * 任务重试次数
			time.Sleep(taskRetryInterval * time.Duration(task.GetRetryTimes()))
			continue
		}
		task.SetStatusComplete()
		log.Info("Task completed successfully")
		break
	}
}
