// Copyright (c) Huawei Technologies Co., Ltd. 2022-2022. All rights reserved.

// 应用进程管理
package processmanager

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"sync"
	"time"

	"github.com/mitchellh/go-ps"
	"google.golang.org/grpc"

	"codehub-g.huawei.com/videocloud/mediaprocesscenter/auxproxy/pkg/apis"
	"codehub-g.huawei.com/videocloud/mediaprocesscenter/auxproxy/pkg/common"
	"codehub-g.huawei.com/videocloud/mediaprocesscenter/auxproxy/pkg/configmanager"
	"codehub-g.huawei.com/videocloud/mediaprocesscenter/auxproxy/pkg/utils/clean"
	"codehub-g.huawei.com/videocloud/mediaprocesscenter/auxproxy/pkg/utils/clients"
	"codehub-g.huawei.com/videocloud/mediaprocesscenter/auxproxy/pkg/utils/log"
	"codehub-g.huawei.com/videocloud/mediaprocesscenter/auxproxy/pkg/utils/sdk/processservice"
)

const (
	newProcessLaunchInterval   = 50 * time.Second
	processHealthCheckInterval = 30 * time.Second
	maxLevel                   = 3
)

type BasicProcessInfo struct {
	// 启动路径的信息
	LaunchPath string
	Parameters string
}

type Process struct {
	// 启动路径的信息
	LaunchPath string
	Parameters string

	// 运行时信息
	isTakeOver   bool // 是否是接管过来的
	isRegistered bool // 是否是已注册的，只有已注册的才可以做健康检查
	Pid          int
	BizPid       int
	Id           string
	Ip           string
	ClientPort   int
	GrpcPort     int
	Status       string
	LogPath      []string
	Client       processservice.ProcessGrpcSdkServiceClient

	// 记录server session是否启动过
	ServerSessionStartedMap map[string]bool
	Mux                     sync.RWMutex
}

// NewProcess 传入基本信息构建Process对象，这里是构建Process的唯一入口
func NewProcess(launchPath, Parameters string, pid int) *Process {
	return &Process{
		Pid:                     pid,
		LaunchPath:              launchPath,
		Parameters:              Parameters,
		ServerSessionStartedMap: make(map[string]bool),
		Mux:                     sync.RWMutex{},
	}
}

// NewBasicProcessInfo 创建只包含launchPath和Parameters两个基本信息的对象
func NewBasicProcessInfo(launchPath, Parameters string) *BasicProcessInfo {
	return &BasicProcessInfo{
		LaunchPath: launchPath,
		Parameters: Parameters,
	}
}

type ProcessManager struct {
	ToBeStartedProcess []*BasicProcessInfo
	Processes          []*Process

	ToBeStartedProcessMux sync.RWMutex
	ProcessMux            sync.RWMutex

	WaitChan chan int
	stopCh   chan struct{}
}

// ProcessMgr process manager
var ProcessMgr *ProcessManager

// InitProcessManager init process manager
func InitProcessManager() {
	var once sync.Once
	once.Do(func() {
		ProcessMgr = &ProcessManager{
			ToBeStartedProcessMux: sync.RWMutex{},
			ToBeStartedProcess:    []*BasicProcessInfo{},
			WaitChan:              make(chan int, 50), // concurrent process num
			ProcessMux:            sync.RWMutex{},
			Processes:             []*Process{},
			stopCh:                make(chan struct{}, 0),
		}
	})
}

// TakeOverRunningProcess 接管游离进程
func (p *ProcessManager) TakeOverRunningProcess(process *Process) error {
	p.ProcessMux.Lock()
	defer p.ProcessMux.Unlock()

	process.isTakeOver = true
	process.isRegistered = true

	p.Processes = append(p.Processes, process)
	conn, err := grpc.Dial(fmt.Sprintf("%s:%d", "localhost", process.GrpcPort), grpc.WithInsecure())
	if err != nil {
		return fmt.Errorf("failed to create process client for %v in take over "+
			"process %d", err, process.Pid)
	}

	cli := processservice.NewProcessGrpcSdkServiceClient(conn)
	process.Client = cli

	log.RunLogger.Infof("[process manager] success take overprocess, the pid is %d and "+
		"bizpid is %d", process.Pid, process.BizPid)
	return nil

}

// AddToBeStartedProcess add a path and parameters to ToBeStartedProcess queue
func (p *ProcessManager) AddToBeStartedProcess(path, parameters string) {
	p.ToBeStartedProcessMux.Lock()
	defer p.ToBeStartedProcessMux.Unlock()

	p.ToBeStartedProcess = append(p.ToBeStartedProcess, NewBasicProcessInfo(path, parameters))

	log.RunLogger.Infof("[process manager] succeed to add process %s %s to be started processes", path, parameters)
}

// IsProcessExisted judge if process exists
func (p *ProcessManager) IsProcessExisted(pid int) bool {
	p.ProcessMux.RLock()
	defer p.ProcessMux.RUnlock()

	for _, pro := range p.Processes {
		if pro.Pid == pid {
			return true
		}
	}

	return false
}

// GetAllBasicProcessInfos 返回所有process的拷贝
func (p *ProcessManager) GetAllBasicProcessInfos() []*BasicProcessInfo {
	var pros []*BasicProcessInfo

	// both check toBeStartedProcesses and Processes
	p.ProcessMux.RLock()
	for _, pro := range p.Processes {
		pros = append(pros, NewBasicProcessInfo(pro.LaunchPath, pro.Parameters))
	}
	p.ProcessMux.RUnlock()

	p.ToBeStartedProcessMux.RLock()
	for _, pro := range p.ToBeStartedProcess {
		pros = append(pros, NewBasicProcessInfo(pro.LaunchPath, pro.Parameters))
	}
	p.ToBeStartedProcessMux.RUnlock()

	return pros
}

// GetAllRunningProcesses 获取全部的运行进程
func (p *ProcessManager) GetAllRunningProcesses() []*Process {
	processCopy := make([]*Process, len(p.Processes))

	p.ProcessMux.RLock()
	copy(processCopy, p.Processes)
	p.ProcessMux.RUnlock()

	return processCopy
}

func (p *ProcessManager) consistProcessByConfiguration() {
	log.RunLogger.Infof("[process manager] consist process, check existed processes")

	pcs := configmanager.ConfMgr.Config.InstanceConfig.RuntimeConfiguration.ProcessConfiguration
	log.RunLogger.Infof("[process manager] get config : %v", pcs)

	var toBeStartedProcesses []*BasicProcessInfo

	// get to be started process, do not care about additional running process
	existedProcesses := ProcessMgr.GetAllBasicProcessInfos()
	for _, pc := range pcs {
		for i := 0; i < pc.ConcurrentExecutions; i++ {
			findFlag := false
			findIndex := -1

			findFlag, findIndex = checkExistedProcess(existedProcesses, pc.LaunchPath, pc.Parameters)

			if findFlag {
				existedProcesses = append(existedProcesses[:findIndex], existedProcesses[findIndex+1:]...)
			} else {
				toBeStartedProcesses = append(toBeStartedProcesses, NewBasicProcessInfo(pc.LaunchPath, pc.Parameters))
			}
		}
	}

	if len(toBeStartedProcesses) == 0 {
		log.RunLogger.Infof("[process manager] there is no any process need to be launch")
		return
	}
	log.RunLogger.Infof("[process manager] such processes %v will be started", toBeStartedProcesses)

	for _, toBeStartedPro := range toBeStartedProcesses {
		ProcessMgr.AddToBeStartedProcess(toBeStartedPro.LaunchPath, toBeStartedPro.Parameters)
	}
}

func (p *ProcessManager) checkExistedProcess(launchPath, parameters string) (bool, int) {
	p.ProcessMux.RLocker()
	defer p.ProcessMux.RUnlock()

	for i, pro := range p.Processes {
		if pro.LaunchPath == launchPath && pro.Parameters == parameters {
			return true, i
		}
	}
	return false, -1
}

// RegisterProcess 业务进程注册
func (p *ProcessManager) RegisterProcess(pid int, clientPort, grpcPort int, logPath []string) error {
	// 设置bizPid，并校验是否是合法
	process := p.InitBizPid(pid)
	if process == nil {
		return fmt.Errorf("invalid process")
	}

	// 上报appgateway注册
	res, err := registerProcess(process.Pid, process.BizPid, process.LaunchPath, process.Parameters)
	if err != nil {
		log.RunLogger.Errorf("[process manager] failed to create process to gateway for %v", err)
		return fmt.Errorf("failed to add process to gateway")
	}
	log.RunLogger.Infof("[process manager] succeed to register an app process to app gateway")

	// 上报appgateway补充process信息
	_, err = updateProcess(clientPort, grpcPort, logPath, res.AppProcess.ID)

	// 绑定health checker
	process.BizPid = pid
	process.Id = res.AppProcess.ID
	process.GrpcPort = grpcPort
	process.ClientPort = clientPort
	process.LogPath = logPath
	conn, err := grpc.Dial(fmt.Sprintf("%s:%d", "localhost", process.GrpcPort), grpc.WithInsecure())
	if err != nil {
		return fmt.Errorf("failed to create process client for %v", err)
	}

	cli := processservice.NewProcessGrpcSdkServiceClient(conn)
	process.Client = cli
	process.isRegistered = true

	log.RunLogger.Infof("[process manager] success register process, the pid is %d and "+
		"bizpid is %d", process.Pid, process.BizPid)

	return nil
}

// InitBizPid 如果返回为空，代表不是合法的进程
func (p *ProcessManager) InitBizPid(pid int) *Process {
	// 判断是否是自己管理的进程
	for _, process := range p.Processes {
		if process.Pid == pid {
			process.BizPid = pid
			log.RunLogger.Infof("[process manager] pid %d is auxproxy's child process", pid)
			return process
		}
	}

	// 上面都没有获取到，就通过系统调用查询
	// 判断是否是自己管理的进程的子进程
	log.RunLogger.Infof("[process manager] can not fetch process by pid %d, try to use parent pid", pid)
	curPid := pid
	// 为了避免父子进程关系过深，先设置最多可以三层
	for i := 0; i < maxLevel; i++ {
		pro, err := ps.FindProcess(curPid)
		if err != nil || pro == nil {
			log.RunLogger.Errorf("[process manager] failed to fetch parent id for %d", pid)
			return nil
		}
		ppid := pro.PPid()

		for _, process := range p.Processes {
			if process.Pid == ppid {
				log.RunLogger.Infof("[process manager] pid %d's parent pid %d is "+
					"auxproxy's child process", pid, ppid)
				process.BizPid = pid
				return process
			}
		}
		curPid = ppid
	}

	log.RunLogger.Infof("[process manager] can not fetch process or parent process by pid %d,"+
		"the process is invalid", pid)
	return nil
}

// GetProcess 根据pid或者业务id查询
func (p *ProcessManager) GetProcess(pid int) *Process {
	// 判断是否是自己管理的进程或者是自己的bizPid（一般是进程的子进程）
	for _, process := range p.Processes {
		if process.Pid == pid || process.BizPid == pid {
			return process
		}
	}
	return nil
}

// RemoveProcess 根据pid清理
func (p *ProcessManager) RemoveProcess(pid int) {
	p.ProcessMux.Lock()
	defer p.ProcessMux.Unlock()

	idx := -1
	for i, pro := range p.Processes {
		if pro.Pid == pid || pro.BizPid == pid {
			idx = i
			break
		}
	}
	if idx == -1 {
		log.RunLogger.Errorf("[process manager] remove not exist process %d", pid)
		return
	}
	p.Processes = append(p.Processes[:idx], p.Processes[idx+1:]...)

	log.RunLogger.Infof("[process manager] success remove process %d", pid)
}

// Work let process manager work
func (p *ProcessManager) Work() {
	go p.receiveWaitSig()

	go p.work()

	log.RunLogger.Infof("[process manager] succeed to start process manager")
}

// Stop 停止processmanager
func (p *ProcessManager) Stop() {
	close(p.stopCh)
}

func (p *ProcessManager) work() {
	processTicker := time.NewTicker(newProcessLaunchInterval)
	healthCheckTicker := time.NewTicker(processHealthCheckInterval)

	// 立即执行一次
	p.consistProcessByConfiguration()
	p.launchProcess()

	for {
		select {
		case <-p.stopCh:
			processTicker.Stop()
			healthCheckTicker.Stop()
			log.RunLogger.Infof("[process manager] stop the process manager")
			return
		case <-processTicker.C:
			// 先一致性配置
			p.consistProcessByConfiguration()
			// 在启动新的进程
			p.launchProcess()
		case <-healthCheckTicker.C:
			// 复制一份避免加锁,可以这样做是因为health check，如果去health check一个已经remove的process
			// 不会造成任何影响；可以被移除的一定是已经出现退出的进程，对一个退出的进程做health checker没有影响
			processCopy := make([]*Process, len(p.Processes))
			p.ProcessMux.RLock()
			copy(processCopy, p.Processes)
			p.ProcessMux.RUnlock()

			wg := sync.WaitGroup{}
			wg.Add(len(p.Processes))
			for _, process := range processCopy {
				// 自管理进程
				go func(process *Process) {
					defer wg.Done()

					p.healthCheck(process)
				}(process)
			}
			wg.Wait()
		}
	}
}

func (p *ProcessManager) launchProcess() {
	// start new process
	log.RunLogger.Infof("[process manager] luanch process")
	var newToBeStartedPros []*BasicProcessInfo
	p.ToBeStartedProcessMux.Lock()
	for _, toBeStartedPro := range p.ToBeStartedProcess {
		cmd := exec.Command(toBeStartedPro.LaunchPath, strings.Split(toBeStartedPro.Parameters, " ")...)

		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		err := cmd.Start()
		if err != nil {
			log.RunLogger.Errorf("[process manager] failed to exec command \"%s %s\" for %v",
				toBeStartedPro.LaunchPath, toBeStartedPro.Parameters, err)
			newToBeStartedPros = append(newToBeStartedPros, toBeStartedPro)
			continue
		}

		go wait(cmd, p)

		p.ProcessMux.Lock()
		p.Processes = append(p.Processes, NewProcess(toBeStartedPro.LaunchPath,
			toBeStartedPro.Parameters, cmd.Process.Pid))
		p.ProcessMux.Unlock()

		log.RunLogger.Infof("[process manager] start process \"%s %s\"",
			toBeStartedPro.LaunchPath, toBeStartedPro.Parameters)
	}
	p.ToBeStartedProcess = newToBeStartedPros
	p.ToBeStartedProcessMux.Unlock()

	log.RunLogger.Infof("[process manager] launch process finish, processes %+v", p.Processes)
}

// 监听自己创建的进程，可以实现进程的快速退出
func (p *ProcessManager) receiveWaitSig() {
	for {
		select {
		case <-p.stopCh:
			return
		case pid := <-p.WaitChan:
			// 接收到管理进程的退出信号，自动删除进程（快速退出）
			log.RunLogger.Infof("[process manager] received a process exit signal with pid %d", pid)
			pro := p.GetProcess(pid)
			if pro == nil {
				log.RunLogger.Infof("[process manager] the %d process have removed from process manager, "+
					"do not need to do remove in receiveWaitSig", pid)
				continue
			}
			r := &apis.UpdateAppProcessStateRequest{State: common.AppProcessStateTerminated}
			_, err := clients.GWClient.UpdateProcessState(pro.Id, r)
			if err != nil {
				log.RunLogger.Errorf("[health checker] set process %v to terminated to gateway", p)
			}
			err = clean.DeleteLogFiles(pro.LogPath)
			if err != nil {
				log.RunLogger.Errorf("[health checker] failed to clean process %v logs for %v", p, err)
			}
			p.RemoveProcess(pro.Pid)
		}
	}
}

func wait(cmd *exec.Cmd, p *ProcessManager) {
	err := cmd.Wait()
	if err != nil {
		log.RunLogger.Errorf("[process manager] cmd wait error for %v", err)
	}
	p.WaitChan <- cmd.Process.Pid
}

func (p *ProcessManager) healthCheck(process *Process) {
	if !process.isRegistered {
		return
	}
	ctx := context.Background()
	res, err := process.Client.OnHealthCheck(ctx, &processservice.HealthCheckRequest{})
	if err != nil {
		// 无响应，打印信息，直接跳过
		log.RunLogger.Errorf("[process manager] health checker for process %d failed "+
			"for %v", process.Pid, err)
		// 如果是接管的进程，需要做下进程的探测，如果确实是挂掉了，要终止掉
		if !process.isTakeOver {
			return
		}
		pro, err := ps.FindProcess(process.Pid)
		if err != nil || pro == nil {
			r := &apis.UpdateAppProcessStateRequest{State: common.AppProcessStateTerminated}
			_, err := clients.GWClient.UpdateProcessState(process.Id, r)
			if err != nil {
				log.RunLogger.Errorf("[health checker] set process %v to terminated to gateway", p)
			}
			err = clean.DeleteLogFiles(process.LogPath)
			if err != nil {
				log.RunLogger.Errorf("[health checker] failed to clean process %v logs for %v", p, err)
			}
			p.RemoveProcess(process.Pid)
		}
		return
	}
	if res.HealthStatus {
		process.Status = common.AppProcessStateActive
	} else {
		process.Status = common.AppProcessStateError
	}

	r := &apis.UpdateAppProcessStateRequest{
		State: process.Status,
	}
	_, err = clients.GWClient.UpdateProcessState(process.Id, r)
	if err != nil {
		log.RunLogger.Errorf("[process manager] failed to update process state for %v"+
			" in health check", err)
		return
	}

	log.RunLogger.Infof("[process manager] suceed to update process state for %v in health check", process.Id)
}

func checkExistedProcess(processes []*BasicProcessInfo, launchPath, parameters string) (bool, int) {
	for i, pro := range processes {
		if pro.LaunchPath == launchPath && pro.Parameters == parameters {
			return true, i
		}
	}
	return false, -1
}

func (p *ProcessManager) IsServerSessionStarted(process *Process, serverSessionId string) bool {
	if nil == process {
		log.RunLogger.Errorf("[process manager] failed to check server session started with "+
			"server session id: %s, process is nil", serverSessionId)
		return false
	}

	log.RunLogger.Infof("[process manager] check is server session started on process: %d "+
		"server session id: %s", process.Pid, serverSessionId)
	process.Mux.RLock()
	defer process.Mux.RUnlock()

	isStarted, ok := process.ServerSessionStartedMap[serverSessionId]
	return ok && isStarted
}

func (p *ProcessManager) RecordServerSessionStarted(process *Process, serverSessionId string) {
	if nil == process {
		log.RunLogger.Errorf("[process manager] failed to record server session started with "+
			"server session id: %s, process is nil", serverSessionId)
		return
	}

	log.RunLogger.Infof("[process manager] record server session started on process: %d "+
		"server session id: %s", process.Pid, serverSessionId)
	process.Mux.Lock()
	defer process.Mux.Unlock()

	process.ServerSessionStartedMap[serverSessionId] = true
	return
}
