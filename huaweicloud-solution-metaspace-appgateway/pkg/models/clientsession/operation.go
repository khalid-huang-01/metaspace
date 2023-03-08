// Copyright (c) Huawei Technologies Co., Ltd. 2022-2022. All rights reserved.

// 客户端会话操作
package clientsession

import (
	server_session "codehub-g.huawei.com/videocloud/mediaprocesscenter/application-gateway/pkg/common"
	"codehub-g.huawei.com/videocloud/mediaprocesscenter/application-gateway/pkg/utils/log"
	"fmt"
	"github.com/beego/beego/v2/client/orm"
)

type ClientSessionDao struct {
	sqlSession orm.Ormer
}

// NewClientSessionDao new client session dao
func NewClientSessionDao(sqlSession orm.Ormer) *ClientSessionDao {
	return &ClientSessionDao{sqlSession: sqlSession}
}

// Insert 插入新的数据到数据库表
func (c *ClientSessionDao) Insert(cs *ClientSession) (*ClientSession, error) {
	_, err := c.sqlSession.Insert(cs)
	return cs, err
}
func (c *ClientSessionDao) InsertMulti(num int, css []*ClientSession) ([]*ClientSession, error) {
	_, err := c.sqlSession.InsertMulti(num, css)
	return css, err

}

// Delete  从数据库中client session表删除数据
func (c *ClientSessionDao) Delete(cs *ClientSession) error {
	_, err := c.sqlSession.Delete(cs)
	return err
}

// Update 更新数据库中client session表的数据
func (c *ClientSessionDao) Update(cs *ClientSession) (*ClientSession, error) {

	_, err := c.sqlSession.Update(cs)
	return cs, err
}

// GetClientSessionByID  通过一个client session的id获取一个client session
func (c *ClientSessionDao) GetClientSessionByID(id string) (*ClientSession, error) {
	var cs ClientSession
	cond := orm.NewCondition()
	cond = cond.And("ID", id)
	err := c.sqlSession.QueryTable(&ClientSession{}).SetCond(cond).One(&cs)
	return &cs, err
}

// ListClientSessionByServerSessionID 通过一个Server Session ID获取client session列表
func (c *ClientSessionDao) ListClientSessionByServerSessionID(ID, sort string,
	offset, limit int) (*[]ClientSession, error) {
	var css []ClientSession

	cond := orm.NewCondition()

	if ID != "" {
		cond = cond.And("SERVER_SESSION_ID", ID)
	}

	_, err := c.sqlSession.QueryTable(&ClientSession{}).SetCond(cond).OrderBy(sort).
		Offset(offset).Limit(limit).All(&css)
	return &css, err
}

// ListAllClientSessionsByServerSessionID 通过一个Server Session ID相关的全部 client session
func (c *ClientSessionDao) ListAllClientSessionsByServerSessionID(ID, sort string) (*[]ClientSession, error) {
	var css []ClientSession

	cond := orm.NewCondition()

	if ID != "" {
		cond = cond.And("SERVER_SESSION_ID", ID)
	}

	_, err := c.sqlSession.QueryTable(&ClientSession{}).SetCond(cond).OrderBy(sort).All(&css)
	return &css, err
}

// CleanClientSession 用于清理数据库中终止的ClientSession
func (c *ClientSessionDao) CleanClientSession() error {
	sqlStr := fmt.Sprintf(`update CLIENT_SESSION SET IS_DELETE = 1 where DATEDIFF(NOW(),CREATED_AT) > 14 
                  and STATE IN ("COMPLETED", "TIMEOUT")`)
	_, err := c.sqlSession.Raw(sqlStr).Exec()
	if err != nil {
		return err
	}
	return nil
}

// ListReservedClientSessionForInstance 查询指定节点所有处于状态的server session
func (c *ClientSessionDao) ListReservedClientSessionForInstance(instance string) ([]ClientSession, error) {
	var css []ClientSession
	sqlStr := fmt.Sprintf(`select * from CLIENT_SESSION where WORK_NODE_ID=? and STATE=?`)
	log.RunLogger.Infof("sqlStr: %v", sqlStr)
	_, err := c.sqlSession.Raw(sqlStr, instance, server_session.ClientSessionStateReserved).QueryRows(&css)
	return css, err

}
