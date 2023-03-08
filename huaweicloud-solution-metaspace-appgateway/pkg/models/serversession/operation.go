// Copyright (c) Huawei Technologies Co., Ltd. 2022-2022. All rights reserved

// 服务端会话相关操作
package serversession

import (
	"fmt"
	"strconv"

	"github.com/beego/beego/v2/client/orm"

	"codehub-g.huawei.com/videocloud/mediaprocesscenter/application-gateway/pkg/common"
	"codehub-g.huawei.com/videocloud/mediaprocesscenter/application-gateway/pkg/utils/log"
)

type ServerSessionDao struct {
	sqlSession orm.Ormer
}

// NewServerSessionDao 创建一个server session dao
func NewServerSessionDao(sqlSession orm.Ormer) *ServerSessionDao {
	return &ServerSessionDao{sqlSession: sqlSession}
}

// Insert 持久化一个新的Server Session
func (s *ServerSessionDao) Insert(ss *ServerSession) (*ServerSession, error) {
	_, err := s.sqlSession.Insert(ss)
	return ss, err
}

// Delete 删除指定的server session
func (s *ServerSessionDao) Delete(ss *ServerSession) error {
	_, err := s.sqlSession.Delete(ss)
	return err
}

// Update 更新指定的server session
func (s *ServerSessionDao) Update(ss *ServerSession) (*ServerSession, error) {
	_, err := s.sqlSession.Update(ss)
	if err != nil {
		return nil, fmt.Errorf("failed to update app process for %v", err)
	}
	return ss, err
}

func (s *ServerSessionDao) UpdateStateAndReason(ss *ServerSession) (*ServerSession, error) {
	_, err := s.sqlSession.Update(ss, FieldNameState, FieldNameStateReason)
	if err != nil {
		return nil, fmt.Errorf("failed to update app process for %v", err)
	}
	return ss, err
}

// GetOneByID 根据ID查找server session
func (s *ServerSessionDao) GetOneByID(id string) (*ServerSession, error) {
	var ss ServerSession

	cond := orm.NewCondition()
	cond = cond.And(FieldNameServerSessionID, id)
	err := s.sqlSession.QueryTable(&ServerSession{}).SetCond(cond).One(&ss)
	return &ss, err
}

// ListByState1 根据指定信息查找server session列表
func (s *ServerSessionDao) ListByState1(state,
	sort string, offset, limit int) (*[]ServerSession, error) {
	var sss []ServerSession

	cond := orm.NewCondition()
	if state != "" {
		cond = cond.And("STATE", state)
	}

	_, err := s.sqlSession.QueryTable(&ServerSession{}).SetCond(cond).OrderBy(sort).
		Offset(offset).Limit(limit).All(&sss)
	return &sss, err
}

// ListByFleetIDAndInstanceIDAndProcessIDAndState 根据指定信息查找server session列表
func (s *ServerSessionDao) ListByFleetIDAndInstanceIDAndProcessIDAndState(fleetID, instanceID, processID, state,
	sort string, offset, limit int) (*[]ServerSession, error) {
	var sss []ServerSession

	cond := orm.NewCondition()
	if fleetID != "" {
		cond = cond.And("FLEET_ID", fleetID)
	}
	if instanceID != "" {
		cond = cond.And("INSTANCE_ID", instanceID)
	}
	if processID != "" {
		cond = cond.And("PROCESS_ID", processID)
	}
	// 外层查询的时候，需要把指定activating查询的时候，把creating也给返回
	switch state {
	case "":
	case common.ServerSessionStateActivating:
		cond2 := orm.NewCondition()
		cond2 = cond2.Or("STATE", common.ServerSessionStateCreating).Or("STATE", common.ServerSessionStateActivating)
		cond = cond.AndCond(cond2)
	default:
		cond = cond.And("STATE", state)
	}

	_, err := s.sqlSession.QueryTable(&ServerSession{}).SetCond(cond).OrderBy(sort).
		Offset(offset).Limit(limit).All(&sss)
	return &sss, err
}

// GetAllServerSessionCountGroupByFleetID 根据fleetID获取所有server session的计数
func (s *ServerSessionDao) GetAllServerSessionCountGroupByFleetID() (map[string]int, error) {
	serverSessionCountsByFleet := map[string]int{}

	var rows []orm.Params
	_, err := s.sqlSession.Raw("SELECT FLEET_ID, COUNT(*) as COUNT FROM SERVER_SESSION "+
		"WHERE STATE=? OR STATE=? GROUP BY FLEET_ID",
		common.ServerSessionStateActivating, common.ServerSessionStateActive).Values(&rows)
	if err != nil {
		return nil, err
	}

	for i, _ := range rows {
		fleetID, _ := rows[i]["FLEET_ID"].(string)
		countStr, _ := rows[i]["COUNT"].(string)

		count, err := strconv.Atoi(countStr)
		if err != nil {
			return nil, err
		}
		serverSessionCountsByFleet[fleetID] = count
	}

	return serverSessionCountsByFleet, err
}

// QueryActivatingServerSession 查询所有处于active状态的server session
func (s *ServerSessionDao) QueryActivatingServerSession() ([]ServerSession, error) {
	var sss []ServerSession
	sqlStr := fmt.Sprintf(`select * from %s where STATE=?`, TableNameServerSession)
	log.RunLogger.Infof("sqlStr: %v", sqlStr)
	_, err := s.sqlSession.Raw(sqlStr, common.ServerSessionStateActivating).QueryRows(&sss)
	return sss, err

}

// QueryActivatingServerSessionByWorknode 查询所有处于active状态的server session
func (s *ServerSessionDao) QueryActivatingServerSessionByWorknode(worknode string) ([]ServerSession, error) {
	var sss []ServerSession
	sqlStr := fmt.Sprintf(`select * from %s where STATE=? and WORK_NODE_ID=?`, TableNameServerSession)
	log.RunLogger.Infof("sqlStr: %v", sqlStr)
	_, err := s.sqlSession.Raw(sqlStr, common.ServerSessionStateActivating, worknode).QueryRows(&sss)
	return sss, err

}

// CleanServerSession 用于清理数据库中终止的ClientSession
func (s *ServerSessionDao) CleanServerSession() error {
	sqlStr := fmt.Sprintf(`update SERVER_SESSION SET IS_DELETE = 1 where DATEDIFF(NOW(),CREATED_AT) > 14 
                                   AND STATE = "TERMINATED"`)
	_, err := s.sqlSession.Raw(sqlStr).Exec()
	if err != nil {
		return err
	}
	return nil

}
