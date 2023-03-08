// Copyright (c) Huawei Technologies Co., Ltd. 2022-2022. All rights reserved.

// 服务端会话与客户端会话连接操作
package models

import (
	"fmt"
	"sync"

	"github.com/beego/beego/v2/client/orm"

	"codehub-g.huawei.com/videocloud/mediaprocesscenter/application-gateway/pkg/common"
	app_process "codehub-g.huawei.com/videocloud/mediaprocesscenter/application-gateway/pkg/models/appprocess"
	client_session "codehub-g.huawei.com/videocloud/mediaprocesscenter/application-gateway/pkg/models/clientsession"
	server_session "codehub-g.huawei.com/videocloud/mediaprocesscenter/application-gateway/pkg/models/serversession"
	"codehub-g.huawei.com/videocloud/mediaprocesscenter/application-gateway/pkg/utils/log"
)

func DispatchServerSession2Process(ss *server_session.ServerSession, ap *app_process.AppProcess) error {
	tx, err := MySqlOrm.Begin()
	if err != nil {
		return err
	}
	sqlStr := fmt.Sprintf("update %s set SERVER_SESSION_COUNT = SERVER_SESSION_COUNT + 1 "+
		"where ID=? and SERVER_SESSION_COUNT < MAX_SERVER_SESSION_NUM", app_process.TableNameAppProcess)
	rsl1, err1 := tx.Raw(sqlStr, ap.ID).Exec()
	if err1 != nil {
		log.RunLogger.Errorf("[transaction] failed to dispatch server session %s to app process "+
			"%d in sql1 because error %v", ss.ID, ap.ID, err1)
		err := tx.Rollback()
		if err != nil {
			log.RunLogger.Errorf("[transaction] sqlStr0 rollback failed for %v", err)
			return err
		}
		return err1
	}
	num, _ := rsl1.RowsAffected()
	if num == 0 {
		log.RunLogger.Errorf("[transaction] failed to dispatch server session %s to app process "+
			"%d because sql1 do not affected any row %v", ss.ID, ap.ID)
		err := tx.Rollback()
		if err != nil {
			log.RunLogger.Errorf("[transaction] sqlStr0 rollback failed for %v", err)
			return err
		}
		return fmt.Errorf("there is no available process")
	}
	_, err2 := tx.Update(ss)
	if err2 != nil {
		log.RunLogger.Errorf("[transaction] failed to dispatch server session %s to app process "+
			"%d in ss update because error %v", ss.ID, ap.ID, err2)
		err := tx.Rollback()
		if err != nil {
			log.RunLogger.Errorf("[transaction] update server session %s rollback failed for %v", ss.ID, err)
			return err
		}
		return err1
	}
	return tx.Commit()

}

// UpdateServerSessionState 变更为Error或者Terminated的时候需要同步去释放Process的server session名额
func UpdateServerSessionState(ss *server_session.ServerSession, tLogger *log.FMLogger) error {
	if ss.State != common.ServerSessionStateError && ss.State != common.ServerSessionStateTerminated {
		return nil
	}

	tx, err := MySqlOrm.Begin()
	if err != nil {
		return err
	}

	// 这里加了行锁，所以并发不好，但是单个process的server session不会太多，所以可以先这样处理
	// 修改process, 释放名额
	sqlStr := fmt.Sprintf("update %s set SERVER_SESSION_COUNT = SERVER_SESSION_COUNT - 1 "+
		"where ID= ? and SERVER_SESSION_COUNT > 0 ", app_process.TableNameAppProcess)
	rsl1, err1 := tx.Raw(sqlStr, ss.ProcessID).Exec()
	if err1 == nil {
		num, _ := rsl1.RowsAffected()
		if num == 0 {
			tLogger.Infof("[transaction] reduce server session count for server session %s in process %s do not "+
				"affected", ss.ID, ss.ProcessID)
		}
	}

	// 修改server session, 并更状态
	sqlStr2 := fmt.Sprintf("update %s set STATE=?,STATE_REASON=? where ID=?",
		server_session.TableNameServerSession)
	rsl2, err2 := tx.Raw(sqlStr2, ss.State, ss.StateReason, ss.ID).Exec()
	if err2 == nil {
		num, _ := rsl2.RowsAffected()
		if num == 0 {
			tLogger.Infof("[transaction] server session state update SQL2 %s: do not affected any row, "+
				"state %s, reason %s, id %v, start to rollback", sqlStr2, ss.State, ss.StateReason, ss.ID)
			// 如果修改无效，证明已经修改过了，也是需要回滚的
			return tx.Rollback()
		}
	}

	if err1 != nil || err2 != nil {
		tLogger.Errorf("[transaction] err1: %v", err1)
		tLogger.Errorf("[transaction] err2: %v", err2)
		tLogger.Errorf("[transaction] start to rollback for server session %s state update", ss.ID)
		return tx.Rollback()
	} else {
		return tx.Commit()
	}

}

// CreateServerSessionAndUpdateProcess 使用事务的方式创建server session并更新process的server session计数
func CreateServerSessionAndUpdateProcess(ss *server_session.ServerSession, tLogger *log.FMLogger) error {

	tx, err := MySqlOrm.Begin()
	if err != nil {
		return err
	}
	// 获取App Process写锁，如果失败直接退出
	// 这里加了行锁，所以并发不好，但是单个process的server session创建不会太多，所以可以先这样处理
	sqlStr0 := fmt.Sprintf("select * from %s where ID=? for update", app_process.TableNameAppProcess)
	_, err0 := tx.Raw(sqlStr0, ss.ProcessID).Exec()
	if err0 != nil {
		tLogger.Errorf("err0: %v", err0)
		err := tx.Rollback()
		if err != nil {
			tLogger.Errorf("[transaction] server session %v sqlStr0 rollback failed for %v", ss.ID, err)
			return err
		}
		return err0
	}

	// 占用一个名额
	// 修改process, 释放名额
	sqlStr := fmt.Sprintf("update %s set SERVER_SESSION_COUNT = SERVER_SESSION_COUNT + 1 "+
		"where ID=? and SERVER_SESSION_COUNT < MAX_SERVER_SESSION_NUM", app_process.TableNameAppProcess)
	rsl, err1 := tx.Raw(sqlStr, ss.ProcessID).Exec()
	if err1 == nil {
		num, _ := rsl.RowsAffected()
		if num == 0 {
			tLogger.Infof("server session %v SQL1 %s: do not affected any row", ss.ID, sqlStr)
		}
	}

	// 创建sever session
	_, err2 := tx.Insert(ss)
	if err1 != nil || err2 != nil {
		tLogger.Errorf("err1: %v", err1)
		tLogger.Errorf("err2: %v", err2)
		err := tx.Rollback()
		if err != nil {
			return err
		}
		tLogger.Errorf("[transaction] server session %v sqlStr1 rollback failed for %v", ss.ID, err)
		return fmt.Errorf("failed to finish create server session and update process")
	} else {
		err := tx.Commit()
		if err != nil {
			return err
		}
		tLogger.Infof("Finish transaction CreateServerSessionAndUpdateProcess for server session %v", ss.ID)
		return nil
	}
}

// CreateServerSessionAndUpdateProcess 使用事务的方式创建server session并更新process的server session计数
func CreateServerSessionAndUpdateProcess1(ss *server_session.ServerSession, tLogger *log.FMLogger) error {

	tx, err := MySqlOrm.Begin()
	if err != nil {
		return err
	}
	// 获取App Process写锁，如果失败直接退出
	// 这里加了行锁，所以并发不好，但是单个process的server session创建不会太多，所以可以先这样处理
	sqlStr0 := fmt.Sprintf("select * from %s where ID=? for update", app_process.TableNameAppProcess)
	_, err0 := tx.Raw(sqlStr0, ss.ProcessID).Exec()
	if err0 != nil {
		tLogger.Errorf("err0: %v", err0)
		err := tx.Rollback()
		if err != nil {
			tLogger.Errorf("[transaction] sqlStr0 rollback failed for %v", err)
			return err
		}
		return err0
	}

	// 占用一个名额
	// 修改process, 释放名额
	sqlStr := fmt.Sprintf("update %s set SERVER_SESSION_COUNT = SERVER_SESSION_COUNT + 1 "+
		"where ID=? and SERVER_SESSION_COUNT < MAX_SERVER_SESSION_NUM", app_process.TableNameAppProcess)
	rsl, err1 := tx.Raw(sqlStr, ss.ProcessID).Exec()
	if err1 == nil {
		num, _ := rsl.RowsAffected()
		if num == 0 {
			tLogger.Infof("SQL1 %s: do not affected any row", sqlStr)
		}
	}

	// 创建sever session
	_, err2 := tx.Insert(ss)
	if err1 != nil || err2 != nil {
		tLogger.Errorf("err1: %v", err1)
		tLogger.Errorf("err2: %v", err2)
		err := tx.Rollback()
		if err != nil {
			return err
		}
		tLogger.Errorf("[transaction] sqlStr1 rollback failed for %v", err)
		return fmt.Errorf("failed to finish create server session and update process")
	} else {
		err := tx.Commit()
		if err != nil {
			return err
		}
		tLogger.Infof("Finish transaction CreateServerSessionAndUpdateProcess")
		return nil
	}
}

// TerminateAllResourcesForServerSession 以事务的方式终止指定server session以及与其相关的所有client session
func TerminateAllResourcesForServerSession(ssID string, tLogger *log.FMLogger) error {
	tx, err := MySqlOrm.Begin()
	if err != nil {
		return nil
	}

	// 修改client session的状态
	sqlStr := fmt.Sprintf("update %s set STATE=? where SERVER_SESSION_ID=?",
		client_session.TableNameClientSession)
	rsl1, err1 := tx.Raw(sqlStr, common.ClientSessionStateCompleted, ssID).Exec()
	if err1 == nil {
		num, _ := rsl1.RowsAffected()
		if num == 0 {
			tLogger.Infof("SQL1 %s: server session %v do not affected any row", sqlStr, ssID)
		}
	}

	// 修改server session, 并更状态
	sqlStr2 := fmt.Sprintf("update %s set CLIENT_SESSION_COUNT=0,STATE=?,STATE_REASON=? where ID=?",
		server_session.TableNameServerSession)
	rsl2, err2 := tx.Raw(sqlStr2, common.ServerSessionStateTerminated,
		"terminate by TerminateAllRelativeResources", ssID).Exec()
	if err2 == nil {
		num, _ := rsl2.RowsAffected()
		if num == 0 {
			tLogger.Infof("SQL2 %s: server session %v do not affected any row", sqlStr2, ssID)
		}
	}

	if err1 != nil || err2 != nil {
		tLogger.Errorf("err1: %v", err1)
		tLogger.Errorf("err2: %v", err2)
		return tx.Rollback()
	} else {
		tLogger.Infof("Finish transaction TerminateAllRelativeResources for server session %v", ssID)
		return tx.Commit()
	}
}

// TerminateOutOfDateServerSession 终止所有超时的server session
func TerminateOutOfDateServerSession(instance string) error {
	// 先获取全部的超时server session，如果只剩下3秒就过期也设置为超时
	sqlStr := fmt.Sprintf("select * FROM %s where WORK_NODE_ID=? and STATE=? and ACTIVATION_TIMEOUT_SECONDS "+
		"< TIMESTAMPDIFF(SECOND, CREATED_AT, DATE_ADD(NOW(),INTERVAL 3 SECOND))", server_session.TableNameServerSession)
	log.RunLogger.Infof("sqlStr: %v", sqlStr)
	var sss []server_session.ServerSession
	_, err := MySqlOrm.Raw(sqlStr, instance, common.ServerSessionStateActivating).QueryRows(&sss)
	if err == orm.ErrNoRows {
		log.RunLogger.Infof("[server session dao] there is no any out of date server session")
		return nil
	} else {
		log.RunLogger.Infof("[server session dao] need to handler %d out of date server sessions", len(sss))
	}
	if err != nil {
		return err
	}
	// 针对每个都调用事务处理
	var wg sync.WaitGroup
	for _, ss := range sss {
		// 做数量控制
		wg.Add(1)
		go func(ss server_session.ServerSession) {
			defer wg.Done()
			log.RunLogger.Infof("[transaction] start to error out of date server session %d", ss.ID)
			ss.State = common.ServerSessionStateError
			ss.StateReason = "terminate server session out fo date in restart service"
			err := UpdateServerSessionState(&ss, log.RunLogger)
			if err != nil {
				log.RunLogger.Errorf("[transaction] failed to update server session %v state in terminate "+
					"out of date server session", ss.ID)
			}
		}(ss)
	}
	wg.Wait()
	return nil
}

// CreateClientSessionAndUpdateServerSession  创建client session和修改数据库的事务
func CreateClientSessionAndUpdateServerSession(cs *client_session.ClientSession, tLogger *log.FMLogger) error {
	tx, err := MySqlOrm.Begin()
	if err != nil {
		return nil
	}

	// 这里加了行锁，所以并发不好
	// 如果后期并发起来了，要使用channel做并发控制
	sqlStr0 := fmt.Sprintf("select * from %s where ID=? for update", server_session.TableNameServerSession)
	_, err0 := tx.Raw(sqlStr0, cs.ServerSessionID).Exec()
	if err0 != nil {
		tLogger.Errorf("err0: %v", err0)
		err := tx.Rollback()
		if err != nil {
			tLogger.Errorf("[transaction] sqlStr0 rollback failed for %v", err)
			return err
		}
		return err0
	}

	// 修改server session的client session count
	sqlStr := fmt.Sprintf("update %s set CLIENT_SESSION_COUNT = CLIENT_SESSION_COUNT + 1 "+
		"where ID=? and CLIENT_SESSION_COUNT < MAX_CLIENT_SESSION_NUM", server_session.TableNameServerSession)
	rsl, err1 := tx.Raw(sqlStr, cs.ServerSessionID).Exec()
	if err1 == nil {
		num, _ := rsl.RowsAffected()
		if num == 0 {
			tLogger.Infof("SQL1 %s: do not affected any row", sqlStr)
		}
	}

	// 创建sever session
	_, err2 := tx.Insert(cs)
	if err1 != nil || err2 != nil {
		tLogger.Errorf("err1: %v", err1)
		tLogger.Errorf("err2: %v", err2)
		return tx.Rollback()
	} else {
		tLogger.Infof("Finish transaction CreateClientSessionAndUpdateServerSession")
		return tx.Commit()
	}
}

// UpdateClientSessionState 更新client session状态与写入数据库的事务
func UpdateClientSessionState(cs *client_session.ClientSession, tLogger *log.FMLogger) error {
	if cs.State != common.ClientSessionStateCompleted && cs.State != common.ClientSessionStateTimeout {
		return nil
	}

	tx, err := MySqlOrm.Begin()
	if err != nil {
		return err
	}
	// 这里加了行锁，所以并发不好
	// 如果后期并发起来了，要使用channel做并发控制
	sqlStr0 := fmt.Sprintf("select * from %s where ID=? for update", server_session.TableNameServerSession)
	_, err0 := tx.Raw(sqlStr0, cs.ServerSessionID).Exec()
	if err0 != nil {
		tLogger.Errorf("err0: %v", err0)
		err := tx.Rollback()
		if err != nil {
			tLogger.Errorf("[transaction] sqlStr0 rollback failed for %v", err)
			return err
		}
		return err0
	}

	// 修改SERVER SESSION, client session count  减少 1
	sqlStr := fmt.Sprintf("update %s set CLIENT_SESSION_COUNT = CLIENT_SESSION_COUNT - 1 "+
		"where ID=? and CLIENT_SESSION_COUNT > 0", server_session.TableNameServerSession)
	rsl1, err1 := tx.Raw(sqlStr, cs.ServerSessionID).Exec()
	if err1 == nil {
		num, _ := rsl1.RowsAffected()
		if num == 0 {
			tLogger.Infof("SQL1 %s: do not affected any row", sqlStr)
		}
	}

	// 修改client  session, 并更状态
	// TODO 这个地方需要加一个行锁
	sqlStr2 := fmt.Sprintf(`update %s set STATE=? where ID=?`, client_session.TableNameClientSession)
	rsl2, err2 := tx.Raw(sqlStr2, common.ClientSessionStateTimeout, cs.ID).Exec()
	if err2 == nil {
		num, _ := rsl2.RowsAffected()
		if num == 0 {
			tLogger.Infof("SQL2 %s: do not affected any row", sqlStr2)
		}
	}

	if err1 != nil || err2 != nil {
		tLogger.Errorf("err1: %v", err1)
		tLogger.Errorf("err2: %v", err2)
		return tx.Rollback()
	} else {
		tLogger.Infof("Finish transaction UpdateClientSessionState")
		return tx.Commit()
	}

}

// CreateClientSessionsAndUpdateServerSession 用于批量创建client session
func CreateClientSessionsAndUpdateServerSession(css []*client_session.ClientSession, tLogger *log.FMLogger) error {
	tx, err := MySqlOrm.Begin()
	if err != nil {
		return nil
	}
	numOfClientSession := len(css)
	ssID := css[0].ServerSessionID
	sqlStr0 := fmt.Sprintf("select * from SERVER_SESSION where ID=? for update")
	_, err0 := tx.Raw(sqlStr0, ssID).Exec()
	if err0 != nil {
		tLogger.Errorf("err0: %v", err0)
		err := tx.Rollback()
		if err != nil {
			tLogger.Errorf("[transaction] sqlStr0 rollback failed for %v", err)
			return err
		}
		return err0
	}
	// 修改server session的client session count
	sqlStr := fmt.Sprintf("update SERVER_SESSION set CLIENT_SESSION_COUNT = CLIENT_SESSION_COUNT + ? " +
		"where ID=? and CLIENT_SESSION_COUNT + ? <= MAX_CLIENT_SESSION_NUM")
	rsl, err1 := tx.Raw(sqlStr, numOfClientSession, ssID, numOfClientSession).Exec()
	if err1 == nil {
		num, _ := rsl.RowsAffected()
		if num == 0 {
			tLogger.Infof("SQL1 %s: do not affected any row", sqlStr)
		}
	}

	// 创建client session
	_, err2 := tx.InsertMulti(numOfClientSession, css)
	if err1 != nil || err2 != nil {
		tLogger.Errorf("err1: %v", err1)
		tLogger.Errorf("err2: %v", err2)
		return tx.Rollback()
	} else {
		tLogger.Infof("Finish transaction CreateClientSessionsAndUpdateServerSession")
		return tx.Commit()
	}
}
