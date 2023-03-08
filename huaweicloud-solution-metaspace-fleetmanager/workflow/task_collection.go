// Copyright (c) Huawei Technologies Co., Ltd. 2022-2022. All rights reserved.

// task collection
package workflow

import (
	"fleetmanager/workflow/components"
	"fmt"
	"sync"
)

type taskCollection struct {
	All   []components.Task
	aLock sync.RWMutex
}

func (tc *taskCollection) addTask(t components.Task) {
	tc.aLock.Lock()
	defer tc.aLock.Unlock()

	tc.All = append(tc.All, t)
}

func (tc *taskCollection) getTask(step int) (components.Task, error) {
	tc.aLock.RLock()
	defer tc.aLock.RUnlock()

	if step > len(tc.All) {
		return nil, fmt.Errorf("task step over range")
	}
	t := tc.All[step-1]

	return t, nil
}

func newTaskCollection() *taskCollection {
	return &taskCollection{
	}
}
