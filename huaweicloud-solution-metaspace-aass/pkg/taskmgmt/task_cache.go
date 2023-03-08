// Copyright (c) Huawei Technologies Co., Ltd. 2022-2022. All rights reserved.

// 任务缓存
package taskmgmt

import (
	"sync"

	"scase.io/application-auto-scaling-service/pkg/taskmgmt/asynctask/interfaces"
)

type taskCache struct {
	lock sync.Mutex
	// 任务记录map：[taskType][taskKey]asynctask
	taskMap map[string]map[string]interfaces.AsyncTaskInf
}

func newTaskCache() *taskCache {
	return &taskCache{taskMap: make(map[string]map[string]interfaces.AsyncTaskInf)}
}

func (c *taskCache) IsTaskExist(taskType, taskKey string) bool {
	c.lock.Lock()
	defer c.lock.Unlock()

	if _, exists := c.taskMap[taskType]; !exists {
		return false
	}
	if _, exists := c.taskMap[taskType][taskKey]; !exists {
		return false
	}
	return true
}

func (c *taskCache) addTask(task interfaces.AsyncTaskInf) {
	c.lock.Lock()
	defer c.lock.Unlock()

	if _, exist := c.taskMap[task.GetType()]; !exist {
		c.taskMap[task.GetType()] = make(map[string]interfaces.AsyncTaskInf)
	}

	c.taskMap[task.GetType()][task.GetKey()] = task
}

func (c *taskCache) delTask(task interfaces.AsyncTaskInf) {
	c.lock.Lock()

	delete(c.taskMap[task.GetType()], task.GetKey())

	c.lock.Unlock()
}
