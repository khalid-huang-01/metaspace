// Copyright (c) Huawei Technologies Co., Ltd. 2022-2022. All rights reserved.

// 工作节点数据表
package dao

import (
	"fleetmanager/db/dbm"
	"fmt"
	"github.com/beego/beego/v2/client/orm"
	"time"
)

// Work node 状态积变化
//                    ←——
//                 ↓         ↑
//  Running  →  Error → Taking Over  →  Finished
//      |                     ↑
//       →   Terminated    --
//
const (
	WorkNodeStateRunning    = "RUNNING"
	WorkNodeStateError      = "ERROR"
	WorkNodeStateTerminated = "TERMINATED"
	WorkNodeStateTakingOver = "TAKINGOVER"
	WorkNodeStateFinished   = "FINISHED"
)

type WorkNode struct {
	Id           string    `orm:"column(id);size(64);pk" json:"id"`
	State        string    `orm:"column(state);size(32)" json:"state"`
	TakeOverId   string    `orm:"column(take_over_id);size(64)" json:"take_over_id"`
	CreationTime time.Time `orm:"column(creation_time);type(datetime);auto_now_add" json:"creation_time"`
	UpdateTime   time.Time `orm:"column(update_time);type(datetime);auto_now" json:"update_time"`
}

func (wn *WorkNode) String() string {
	return fmt.Sprintf("WorkNode(id=%s, state=%s, take_over_id=%s, creation_time=%v, update_time=%v)",
		wn.Id, wn.State, wn.TakeOverId, wn.CreationTime, wn.UpdateTime)
}

// Insert 插入work node
func InsertWorkNode(wn *WorkNode) error {
	_, err := dbm.Ormer.Insert(wn)
	return err
}

// Get 获取work node
func GetWorkNodes(f Filters) ([]WorkNode, error) {
	var wns []WorkNode
	_, err := f.Filter(WorkNodeTable).All(&wns)
	return wns, err
}

// TakeOver 进行work node接管, 原子性操作
func TakeOverWorkNode(originWorkNodeId string, originState string, takeOverId string) (int64, error) {
	return dbm.Ormer.QueryTable(WorkNodeTable).
		Filter("Id", originWorkNodeId).
		Filter("State", originState).
		Update(orm.Params{
			"State":      WorkNodeStateTakingOver,
			"TakeOverId": takeOverId,
			"UpdateTime": time.Now().UTC(),
		})
}

// UpdateWorkNodeState 更新节点状态
func UpdateWorkNodeState(wnId string, state string) error {
	_, err := dbm.Ormer.QueryTable(WorkNodeTable).
		Filter("Id", wnId).
		Update(orm.Params{
			"State":      state,
			"UpdateTime": time.Now().UTC(),
		})
	return err
}

// UpdateWorkNodeState 更新节点状态
func UpdateDeadWorkNodeState(deadTime string) (int64, error) {
	return dbm.Ormer.QueryTable(WorkNodeTable).
		Filter("State__in", []string{WorkNodeStateRunning, WorkNodeStateTakingOver}).
		Filter("UpdateTime__lt", deadTime).
		Update(orm.Params{
			"State": WorkNodeStateError,
		})
}

// QueryDeadWorkNode 查询Dead WorkNode列表
func QueryDeadWorkNode(deadTime string) ([]WorkNode, error) {
	var wns []WorkNode
	_, err := dbm.Ormer.QueryTable(WorkNodeTable).
		Filter("State__in", []string{WorkNodeStateRunning, WorkNodeStateTakingOver}).
		Filter("UpdateTime__lt", deadTime).All(&wns)
	return wns, err
}

// HeartBeat 节点心跳
func HeartBeat(workNodeId string) error {
	_, err := dbm.Ormer.QueryTable(WorkNodeTable).
		Filter("Id", workNodeId).
		Update(orm.Params{
			"UpdateTime": time.Now().UTC(),
		})
	return err
}

// TerminateWorkNode 终止节点
func TerminateWorkNode(workNodeId string) error {
	_, err := dbm.Ormer.QueryTable(WorkNodeTable).
		Filter("Id", workNodeId).
		Update(orm.Params{
			"State":         WorkNodeStateTerminated,
		})
	return err
}
