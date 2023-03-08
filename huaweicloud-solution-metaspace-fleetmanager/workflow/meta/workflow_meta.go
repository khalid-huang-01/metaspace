// Copyright (c) Huawei Technologies Co., Ltd. 2022-2022. All rights reserved.

// workflow_meta
package meta

type WorkflowMeta struct {
	Name        string     `json:"name"`
	Description string     `json:"description"`
	Version     string     `json:"version"`
	Tasks       []TaskMeta `json:"tasks"`
}
