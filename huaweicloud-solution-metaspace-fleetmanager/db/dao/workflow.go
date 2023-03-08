// Copyright (c) Huawei Technologies Co., Ltd. 2022-2022. All rights reserved.

// 工作流数据表
package dao

import (
	"fleetmanager/db/dbm"
	"github.com/beego/beego/v2/client/orm"
	"time"
)

const (
	WorkflowStateCreate      = "CREATE"
	WorkflowStateRunning     = "RUNNING"
	WorkflowStateError       = "ERROR"
	WorkflowStateRollbacking = "ROLLBACKING"
	WorkflowStateRollbacked  = "ROLLBACKED"
	WorkflowStateFinished    = "FINISHED"
)

type Workflow struct {
	Id           string    `orm:"column(id);size(64);pk" json:"id"`
	ProjectId    string    `orm:"column(project_id);size(64)" json:"project_id"`
	State        string    `orm:"column(state);size(32)" json:"state"`
	ResourceId   string    `orm:"column(resource_id);size(64)" json:"resource_id"`
	Meta         string    `orm:"column(meta);type(text)" json:"meta"`
	Parameter    string    `orm:"column(parameter);type(text)" json:"parameter"`
	WorkNodeId   string    `orm:"column(work_node_id);size(64)" json:"work_node_id"`
	CreationTime time.Time `orm:"column(creation_time);type(datetime);auto_now_add" json:"creation_time"`
	UpdateTime   time.Time `orm:"column(update_time);type(datetime);auto_now" json:"update_time"`
}

// InsertWorkflow 插入工作流
func InsertWorkflow(w *Workflow) error {
	_, err := dbm.Ormer.Insert(w)
	return err
}

// GetWorkflow 获取工作流
func GetWorkflow(f Filters) (*Workflow, error) {
	var workflow Workflow
	err := f.Filter(WorkflowTable).One(&workflow)
	if err != nil {
		return nil, err
	}

	return &workflow, err
}

// GetAllWorkflows 获取所有的工作流
func GetAllWorkflows(f Filters) ([]Workflow, error) {
	var wfs []Workflow
	_, err := f.Filter(WorkflowTable).All(&wfs)
	return wfs, err
}

// UpdateWorkflow 更新工作流
func UpdateWorkflow(f *Workflow, cols ...string) error {
	_, err := dbm.Ormer.Update(f, cols...)
	return err
}

// TakeOverWorkFlow 接管工作流
func TakeOverWorkFlow(wfIds []string, takeOverId string) error {
	if len(wfIds) == 0 {
		return nil
	}
	_, err := dbm.Ormer.QueryTable(WorkflowTable).
		Filter("Id__in", wfIds).
		Update(orm.Params{
			"WorkNodeId": takeOverId,
		})
	return err
}
