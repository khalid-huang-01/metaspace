// Copyright (c) Huawei Technologies Co., Ltd. 2022-2022. All rights reserved.

// 锁操作
package lock

import (
	"fmt"
	"time"

	"github.com/beego/beego/v2/client/orm"
)

type Dao struct {
	sqlSession orm.Ormer
}

func NewLockDao(sqlSession orm.Ormer) *Dao {
	return &Dao{sqlSession: sqlSession}
}

// NewAndLock Lock 直接创建并加锁
func (l *Dao) NewAndLock(lock *Lock) error {
	_, err := l.sqlSession.Insert(lock)
	return err
}

// Acquire 使用条件更新的原子性和rc隔离性保证只有一个可以获取到锁
func (l *Dao) Acquire(lock *Lock) error {
	rsl, err := l.sqlSession.Raw("UPDATE DISTRIBUTED_LOCK SET EXPIRED_AT=?, HOLDER=? WHERE NAME=? AND "+
		"EXPIRED_AT<? AND CATEGORY=?",
		lock.ExpiredAt, lock.Holder, lock.Name, time.Now().String(), lock.Category).Exec()
	if err != nil {
		return err
	}
	num, _ := rsl.RowsAffected()
	if num == 0 {
		return fmt.Errorf("lock %s is holded by other instance", lock.Holder)
	}
	return nil
}

// Release 解锁
func (l *Dao) Release(lock *Lock) error {
	sqlStr := fmt.Sprintf("DELETE FROM %s WHERE NAME=? AND HOLDER=? AND CATEGORY=?", TableNameLock)
	rsl, err := l.sqlSession.Raw(sqlStr, lock.Name, lock.Holder, lock.Category).Exec()
	if err != nil {
		return nil
	}
	num, _ := rsl.RowsAffected()
	if num == 0 {
		return fmt.Errorf("there is not lock %+v to release", lock)
	}

	return err
}

// RefreshLease 续期, refresh 要确保护是自己的锁
func (l *Dao) RefreshLease(lock *Lock) error {
	sqlStr := fmt.Sprintf("update %s set EXPIRED_AT=? where NAME=? and HOLDER=? AND CATEGORY=?", TableNameLock)
	rsl, err := l.sqlSession.Raw(sqlStr, lock.ExpiredAt, lock.Name, lock.Holder, lock.Category).Exec()
	if err != nil {
		return err
	}
	num, _ := rsl.RowsAffected()
	if num == 0 {
		return fmt.Errorf("there is no valid lock")
	}
	return nil
}

// GetLock 获取锁的信息
func (l *Dao) GetLock(name, category string) (*Lock, error) {
	var lock Lock
	cond := orm.NewCondition()
	cond = cond.And(FieldNameName, name)
	cond = cond.And(FieldNameCategory, category)
	err := l.sqlSession.QueryTable(&Lock{}).SetCond(cond).One(&lock)
	return &lock, err
}

// GetExpiredLocks 获取某一类别下全部过期的锁
func (l *Dao) GetExpiredLocks(category string) ([]Lock, error) {
	var locks []Lock
	_, err := l.sqlSession.Raw("SELECT * FROM DISTRIBUTED_LOCK WHERE CATEGORY=? AND EXPIRED_AT<?",
		category, time.Now().String()).QueryRows(&locks)
	return locks, err
}
