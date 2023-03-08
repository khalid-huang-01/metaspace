// Copyright (c) Huawei Technologies Co., Ltd. 2022-2022. All rights reserved.

// 工作节点定义
package db

import (
	"time"

	"github.com/beego/beego/v2/client/orm"
	"github.com/pkg/errors"
)

const (
	tableNameWorkNode = "work_node"

	// WorkNode 状态变化
	//              ← ← ← ← ←
	//             ↓          ↑
	// running → error → takingOver → deleted
	//    ↓                   ↑
	//    → → → terminated → →
	WorkNodeStateRunning    = "running"
	WorkNodeStateError      = "error"
	WorkNodeStateTerminated = "terminated"
	WorkNodeStateTakingOver = "takingOver"
	WorkNodeStateDeleted    = "deleted"
)

type WorkNode struct {
	Id         string `orm:"column(id);size(64);pk"`
	IP         string `orm:"column(ip);size(32)"`
	State      string `orm:"column(state);size(32)"`
	TakeOverId string `orm:"column(take_over_id);size(64)"`
	TimeModel
}

// Insert 插入work node
func InsertWorkNode(wn *WorkNode) error {
	if wn == nil {
		return errors.New("func InsertWorkNode has invalid args")
	}
	wn.IsDeleted = notDeletedFlag
	wn.State = WorkNodeStateRunning
	_, err := ormer.Insert(wn)
	return err
}

// Get 获取work node
func GetWorkNodesByFilterStates(states []string) ([]WorkNode, error) {
	var wns []WorkNode
	_, err := ormer.QueryTable(tableNameWorkNode).Filter(fieldNameStateIn, states).All(&wns)
	return wns, err
}

// UpdateWorkNodeState 更新节点状态
func DeleteWorkNodeState(wnId string) error {
	timestamp := time.Now().UTC()
	_, err := ormer.QueryTable(tableNameWorkNode).
		Filter(fieldNameId, wnId).
		Update(orm.Params{
			fieldNameState:     WorkNodeStateDeleted,
			fieldNameIsDeleted: deletedFlag,
			fieldNameUpdateAt:  timestamp,
			fieldNameDeleteAt:  timestamp,
		})
	return err
}

// TakeOver 进行work node接管, 原子性操作
func TakeOverWorkNode(originWorkNodeId string, originState string, takeOverId string) (int64, error) {
	return ormer.QueryTable(tableNameWorkNode).
		Filter(fieldNameId, originWorkNodeId).
		Filter(fieldNameState, originState).
		Update(orm.Params{
			fieldNameState:      WorkNodeStateTakingOver,
			fieldNameTakeOverId: takeOverId,
			fieldNameUpdateAt:   time.Now().UTC(),
		})
}

// UpdateWorkNodeState 更新节点状态
func UpdateWorkNodeState(wnId string, state string) error {
	_, err := ormer.QueryTable(tableNameWorkNode).
		Filter(fieldNameId, wnId).
		Update(orm.Params{
			fieldNameState:    state,
			fieldNameUpdateAt: time.Now().UTC(),
		})
	return err
}

// UpdateWorkNodeState 更新节点状态
func UpdateDeadWorkNodeState(deadTime string) (int64, error) {
	return ormer.QueryTable(tableNameWorkNode).
		Filter(fieldNameStateIn, []string{WorkNodeStateRunning, WorkNodeStateTakingOver}).
		Filter(fieldNameUpdateAtLt, deadTime).
		Update(orm.Params{
			fieldNameState: WorkNodeStateError,
		})
}

// QueryDeadWorkNode 查询Dead WorkNode列表
func QueryDeadWorkNode(deadTime string) ([]WorkNode, error) {
	var wns []WorkNode
	_, err := ormer.QueryTable(tableNameWorkNode).
		Filter(fieldNameStateIn, []string{WorkNodeStateRunning, WorkNodeStateTakingOver}).
		Filter(fieldNameUpdateAtLt, deadTime).All(&wns)
	return wns, err
}

// HeartBeat 节点心跳
func HeartBeat(workNodeId string) error {
	_, err := ormer.QueryTable(tableNameWorkNode).
		Filter(fieldNameId, workNodeId).
		Update(orm.Params{
			fieldNameUpdateAt: time.Now().UTC(),
		})
	return err
}

// TerminateWorkNode 终止节点
func TerminateWorkNode(workNodeId string) error {
	_, err := ormer.QueryTable(tableNameWorkNode).
		Filter(fieldNameId, workNodeId).
		Update(orm.Params{
			fieldNameState: WorkNodeStateTerminated,
		})
	return err
}
