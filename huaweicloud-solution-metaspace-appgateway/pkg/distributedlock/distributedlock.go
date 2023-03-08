// Copyright (c) Huawei Technologies Co., Ltd. 2022-2022. All rights reserved.

// 分布式锁
package distributedlock

import (
	"crypto/rand"
	"math/big"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/beego/beego/v2/client/orm"

	"codehub-g.huawei.com/videocloud/mediaprocesscenter/application-gateway/config"
	"codehub-g.huawei.com/videocloud/mediaprocesscenter/application-gateway/pkg/models"
	"codehub-g.huawei.com/videocloud/mediaprocesscenter/application-gateway/pkg/models/lock"
	"codehub-g.huawei.com/videocloud/mediaprocesscenter/application-gateway/pkg/utils/log"
)

type Role string

const (
	// RoleHolder 获取到锁是Holder，
	RoleHolder Role = "Holder"
	// RoleCompetitor 没有获取到锁是Competitor
	RoleCompetitor                    Role = "Competitor"
	defaultInternalLockLeaseTime           = 60 * time.Second // 每次续期60秒
	defaultTryLockOrLeaseBaseInterval      = 40               // 尝试获取锁或者续期的间隔基础时间（实际时间是这个时间加上10s内的随机时间）
)

type LockProcessor interface {
	start()                                    // 开始选举
	stop()                                     // 退出，如果是leader，要负责释放
	setLockLeaseTime(t time.Duration)          // 锁的续期时间
	setTryLockOrLeaseInterval(t time.Duration) // 获取锁/续期锁的操作间隔时间
	exit()                                     // 只用与test，模拟crash
	getLockName() string                       // 获取锁名称
}

// BizWorker 业务worker
type BizWorker interface {
	HolderHook()     // 成为Holder之后需要调用的方法
	CompetitorHook() // 成为Competitor之后需要调用的方法
}

type MySQLLockProcessor struct {
	name      string     // 锁名
	category  string     // 分类
	lock      *lock.Lock // 缓存当前的lock对象
	isSuccess bool       // 当前是否获取到锁
	holder    string
	stopCh    chan struct{} // stop信号，优雅退出
	exitCh    chan struct{} // exit信号，模拟crash
	role      chan Role     // 与控制器的交互信道
	// 时间，一般使用默认配置就可以了
	lockLeaseTime            time.Duration // 每次续期的时间
	retryLockOrLeaseInterval time.Duration // 尝试获取锁或者续期的间隔基础时间（实际时间是这个时间加上10s内的随机时间）
}

// NewMySQLLockProcessor 实例化一个基于mysql分布式锁的选举器
func NewMySQLLockProcessor(role chan Role, lockName, category, holder string) *MySQLLockProcessor {
	return &MySQLLockProcessor{
		name:                     lockName,
		category:                 category,
		isSuccess:                false,
		holder:                   holder,
		stopCh:                   make(chan struct{}),
		exitCh:                   make(chan struct{}),
		role:                     role,
		lockLeaseTime:            defaultInternalLockLeaseTime,
		retryLockOrLeaseInterval: randomizedTryLockOrLeaseInterval(),
	}
}

func (p *MySQLLockProcessor) start() {
	lockDao := lock.NewLockDao(models.MySqlOrm)

	// 直接启动一轮选举
	p.execOneRoundElection()

	// 周期性触发续期或者获取锁
	t := time.NewTicker(p.retryLockOrLeaseInterval)

	for {
		select {
		// for test,直接退出，不清理锁，模拟crash
		case <-p.exitCh:
			t.Stop()
			return

		case <-p.stopCh:
			t.Stop()
			if p.isSuccess != true {
				return
			}
			// 持有者，需要释放锁
			err := lockDao.Release(p.lock)
			if err != nil {
				log.RunLogger.Infof("[%s lock processor] failed to unlock lock %s when stop", p.holder, p.name)
			} else {
				log.RunLogger.Infof("[%s lock processor] success to unlock lock %s when stop", p.holder, p.name)
			}
			return

		case <-t.C:
			if !p.isSuccess {
				p.tryAcquireLock()
				continue
			}
			// 如果已经是获取到锁的，就直接续期
			p.lock.ExpiredAt = time.Now().Add(p.lockLeaseTime)
			err := lockDao.RefreshLease(p.lock)
			if err != nil {
				// 如果续期失败
				log.RunLogger.Infof("[%s lock processor] failed to continue refresh lock %s", p.holder, p.name)
				p.isSuccess = false
				p.role <- RoleCompetitor
				continue
			}
			log.RunLogger.Debugf("[%s lock processor] success to continue refresh lock %s", p.holder, p.name)
		}
	}
}

func (p *MySQLLockProcessor) exit() {
	close(p.exitCh)
}

func (p *MySQLLockProcessor) getLockName() string {
	return p.name
}

func (p *MySQLLockProcessor) stop() {
	close(p.stopCh)
}

func (p *MySQLLockProcessor) setLockLeaseTime(t time.Duration) {
	p.lockLeaseTime = t
}

func (p *MySQLLockProcessor) setTryLockOrLeaseInterval(t time.Duration) {
	p.retryLockOrLeaseInterval = t
}

func (p *MySQLLockProcessor) tryAcquireLock() {
	p.execOneRoundElection()
}

func (p *MySQLLockProcessor) execOneRoundElection() {
	lockDao := lock.NewLockDao(models.MySqlOrm)

	// 这里需要先查，如果查不到才可以开始创建
	var err error
	p.lock, err = lockDao.GetLock(p.name, p.category)
	if err == orm.ErrNoRows {
		log.RunLogger.Infof("[%s lock processor] lock %s is no exit, start to create", p.holder, p.name)
		p.lock = &lock.Lock{
			ExpiredAt: time.Now().Add(p.lockLeaseTime),
			Name:      p.name,
			Holder:    p.holder,
			Category:  p.category,
		}
		err = lockDao.NewAndLock(p.lock)
		if err == nil {
			p.isSuccess = true
			log.RunLogger.Infof("[%s lock processor] success to create and acquire lock %s", p.holder, p.name)
			p.role <- RoleHolder
		} else {
			p.isSuccess = false
			log.RunLogger.Infof("[%s lock processor] failed to create and acquire lock %s", p.holder, p.name)
			p.role <- RoleCompetitor
		}
	} else {
		// 尝试去获取锁
		if p.lock.ExpiredAt.After(time.Now()) {
			log.RunLogger.Debugf("[%s lock processor] failed to acquire lock %s, because the lock is "+
				"holder by other instance", p.holder, p.name)
		} else {
			// 如果过期，就加行锁，修改占用
			p.lock.ExpiredAt = time.Now().Add(p.lockLeaseTime)
			p.lock.Holder = p.holder
			p.lock.Name = p.name
			p.lock.Category = p.category
			err = lockDao.Acquire(p.lock)
			if err != nil {
				log.RunLogger.Debugf("[%s lock processor] failed to acquire lock %s for %v", p.holder,
					err, p.name)
			} else {
				// 如果成功加锁
				log.RunLogger.Infof("[%s lock processor] success to acquire lock %s", p.holder, p.name)
				p.isSuccess = true
				p.role <- RoleHolder
			}
		}
	}
}

type Controller struct {
	name          string    // 一般对应holder
	worker        BizWorker // 负责承接业务，主要是钩子函数
	role          Role
	lockProcessor LockProcessor
	// test使用，模拟crash
	exitCh chan struct{}
	// controller 优雅退出信号
	stopCh chan struct{}
	// 监控关机信号，做资源释放
	signalCh chan os.Signal
	// 接收角色信息变化
	roleCh chan Role
}

// NewDistributedLockController 实例化分布式锁控制器
// 分布式锁可以用于做主备、做实例存活判断
func NewDistributedLockController(lockName string, category string, worker BizWorker) *Controller {
	m := &Controller{
		name:     config.GlobalConfig.InstanceName,
		role:     RoleCompetitor,
		exitCh:   make(chan struct{}),
		stopCh:   make(chan struct{}),
		signalCh: make(chan os.Signal),
		roleCh:   make(chan Role),
		worker:   worker,
	}
	m.lockProcessor = NewMySQLLockProcessor(m.roleCh, lockName, category, m.name)
	return m
}

// NewDistributedLockControllerWithName 只用于测试
func NewDistributedLockControllerWithName(lockName, holder, category string, worker BizWorker) *Controller {
	m := &Controller{
		name:     holder,
		role:     RoleCompetitor,
		exitCh:   make(chan struct{}),
		stopCh:   make(chan struct{}),
		signalCh: make(chan os.Signal),
		roleCh:   make(chan Role),
		worker:   worker,
	}
	m.lockProcessor = NewMySQLLockProcessor(m.roleCh, lockName, category, m.name)
	return m
}

// Work 启动指标监控器
func (m *Controller) Work() {
	go m.work()
}

func (m *Controller) Stop() {
	m.stop()
}

func (m *Controller) SetLockLeaseTime(t time.Duration) {
	m.lockProcessor.setLockLeaseTime(t)
}

func (m *Controller) SetTryLockOrLeaseInterval(t time.Duration) {
	m.lockProcessor.setTryLockOrLeaseInterval(t)
}

func (m *Controller) work() {
	log.RunLogger.Infof("[%s lock controller] start to work as %v for lock %s",
		m.name, m.role, m.lockProcessor.getLockName())
	// 监控关机信号，做解锁操作
	signal.Notify(m.signalCh, syscall.SIGINT, syscall.SIGTERM)

	// 立即触发选主的逻辑
	go m.startElection()

	t := time.NewTicker(120 * time.Second) // 每隔120s打印一些数据

	for {
		select {
		case <-t.C:
			log.RunLogger.Infof("[%s lock controller] my role is %v for lock %s",
				m.name, m.role, m.lockProcessor.getLockName())
		case <-m.exitCh:
			m.lockProcessor.exit()
			return
		// monitor的停止唯一入口
		case <-m.stopCh:
			log.RunLogger.Infof("[%s lock controller] shutting down self (lock %s)", m.name,
				m.lockProcessor.getLockName())
			t.Stop()

			// 确保业务可以关闭
			if m.role == RoleHolder {
				m.worker.CompetitorHook()
			}

			time.Sleep(2 * time.Second) // sleep 2 秒确保业务停止
			m.lockProcessor.stop()

			log.RunLogger.Infof("[%s lock controller] finish shutdown (lock %s)", m.name,
				m.lockProcessor.getLockName())
			return
		// 接收到关机信息
		case signalVal := <-m.signalCh:
			log.RunLogger.Infof("[%s lock controller] receive stop signal %v, will trigger stop", m.name,
				signalVal)
			m.stop()
		// 接收到身份变化
		case role := <-m.roleCh:
			if role == m.role {
				continue
			}
			log.RunLogger.Infof("[%s lock controller] change role from %v to %v for lock %s",
				m.name, m.role, role, m.lockProcessor.getLockName())
			m.role = role
			go func(role Role) {
				if role == RoleHolder {
					m.worker.HolderHook()
				} else if role == RoleCompetitor {
					m.worker.CompetitorHook()
				}
			}(m.role)
		}
	}
}

func (m *Controller) stop() {
	close(m.stopCh)
}

func (m *Controller) exit() {
	close(m.exitCh)
}

func (m *Controller) startElection() {
	// 尝试去获取锁
	m.lockProcessor.start()
}

func randomizedTryLockOrLeaseInterval() time.Duration {
	maxVal := new(big.Int).SetInt64(10)
	randomVal, err := rand.Int(rand.Reader, maxVal)
	if err != nil {
		log.RunLogger.Errorf("[distributed lock] failed to use rand.Read")
		return defaultTryLockOrLeaseBaseInterval * time.Second
	}
	return time.Duration(defaultTryLockOrLeaseBaseInterval+randomVal.Int64()) * time.Second // 随机10秒，避免羊群
}
