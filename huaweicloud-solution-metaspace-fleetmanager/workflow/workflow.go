// Copyright (c) Huawei Technologies Co., Ltd. 2022-2022. All rights reserved.

// workflow
package workflow

import (
	"encoding/json"
	"fleetmanager/api/errors"
	"fleetmanager/config"
	"fleetmanager/db/dao"
	"fleetmanager/logger"
	"fleetmanager/workflow/components"
	"fleetmanager/workflow/directer"
	"fleetmanager/workflow/meta"
	"fmt"
	"github.com/google/uuid"
	"io/ioutil"
	"time"
)

const (
	DefaultExecChLength = 128
)

type Workflow struct {
	logger     *logger.FMLogger
	dbInfo     *dao.Workflow
	execCh     chan *directer.ExecuteContext
	Context    *directer.WorkflowContext
	Tasks      *taskCollection
	Id         string
	Meta       meta.WorkflowMeta
	Parameter  string
	EntryTask  components.Task
	resourceId string
	err        error
	rollback   bool
	projectId  string
	worknodeId string
}

// Process 执行任务
func (wf *Workflow) Process(ctx *directer.ExecuteContext) {
	wf.execCh <- ctx
}

// GetLogger 获取日志对象
func (wf *Workflow) GetLogger() *logger.FMLogger {
	return wf.logger
}

// GetContext 获取Context对象
func (wf *Workflow) GetContext() *directer.WorkflowContext {
	return wf.Context
}

func (wf *Workflow) failed(err error) {
	wf.err = err
	wf.logger.WithFields(map[string]interface{}{
		logger.Stage: "workflow_failed",
		logger.Error: fmt.Sprintf("%v", err),
	}).Error("workflow failed")
}

func (wf *Workflow) setRollbackFlag(err error) {
	if wf.rollback {
		return
	}
	wf.rollback = true
	wf.failed(err)
}

func (wf *Workflow) error(err error, output interface{}, direction directer.WorkflowDirection) {
	wf.logger.WithFields(map[string]interface{}{
		logger.Stage:             "workflow_task_failed",
		logger.Error:             err.Error(),
		logger.WorkflowDirection: direction,
	}).Error("task failed")
}

// 记录数据库
func (wf *Workflow) insertDB() error {
	metaStr, err := json.Marshal(wf.Meta)
	if err != nil {
		return err
	}
	w := &dao.Workflow{
		Id:           wf.Id,
		State:        dao.WorkflowStateCreate,
		ResourceId:   wf.resourceId,
		Parameter:    string(wf.Context.Parameter),
		Meta:         string(metaStr),
		ProjectId:    wf.projectId,
		CreationTime: time.Now().UTC(),
		UpdateTime:   time.Now().UTC(),
		WorkNodeId:   wf.worknodeId,
	}
	wf.dbInfo = w
	if err := dao.InsertWorkflow(w); err != nil {
		return errors.NewError(errors.DBError)
	}

	return nil
}

func (wf *Workflow) save(state string) error {
	metaStr, err := json.Marshal(wf.Meta)
	if err != nil {
		return err
	}
	wf.dbInfo.Meta = string(metaStr)
	wf.dbInfo.Parameter = string(wf.Context.Parameter)
	wf.dbInfo.State = state
	wf.dbInfo.UpdateTime = time.Now().UTC()
	return dao.UpdateWorkflow(wf.dbInfo, "State", "Meta", "Parameter", "UpdateTime")
}

func (wf *Workflow) load(logger *logger.FMLogger) error {
	filter := dao.Filters{"Id": wf.Id}
	wfdb, err := dao.GetWorkflow(filter)
	if err != nil {
		return err
	}
	wf.dbInfo = wfdb
	wf.Parameter = wf.dbInfo.Parameter
	wf.Tasks = newTaskCollection()
	wf.execCh = make(chan *directer.ExecuteContext, DefaultExecChLength)
	wf.logger = logger
	wf.Context = &directer.WorkflowContext{}
	wf.worknodeId = wfdb.WorkNodeId

	m := meta.WorkflowMeta{}
	if err = json.Unmarshal([]byte(wf.dbInfo.Meta), &m); err != nil {
		return err
	}
	wf.Meta = m

	if err := wf.parseParameterStr(wf.dbInfo.Parameter); err != nil {
		return err
	}

	if err := wf.parseTasks(); err != nil {
		return err
	}

	return nil
}

func (wf *Workflow) processExecEvent(ctx *directer.ExecuteContext) (exitWorkflow bool) {
	if ctx.Ended() {
		if ctx.Err != nil {
			wf.failed(ctx.Err)
			if err := wf.save(dao.WorkflowStateError); err != nil {
				return true
			}
		}

		if ctx.Direction == directer.PositiveDirection {
			if err := wf.save(dao.WorkflowStateFinished); err != nil {
				return true
			}
		} else {
			if err := wf.save(dao.WorkflowStateRollbacked); err != nil {
				return true
			}
		}
		return true
	}

	t, err := wf.Tasks.getTask(ctx.Next)
	if err != nil {
		wf.failed(err)
		return true
	}

	if ctx.Direction == directer.PositiveDirection {
		if output, e := t.Execute(ctx); e != nil {
			wf.error(e, output, ctx.Direction)
		}
		if err := wf.save(dao.WorkflowStateRunning); err != nil {
			return true
		}
	} else {
		wf.setRollbackFlag(ctx.Err)
		if output, e := t.Rollback(ctx); e != nil {
			wf.error(e, output, ctx.Direction)
		}
		if err := wf.save(dao.WorkflowStateRollbacking); err != nil {
			return true
		}
	}

	return false
}

func (wf *Workflow) run() {
	wf.logger.WithField(logger.Stage, "workflow_begin").Info("workflow begin")

	ticker := time.NewTicker(5 * time.Second)
	wf.Process(&directer.ExecuteContext{
		Next: wf.EntryTask.TaskStep(),
	})

	for {
		select {
		case <-ticker.C:
			wf.logger.WithField(logger.Stage, "workflow_heartbeat").Info("workflow is running")
		case execCtx := <-wf.execCh:
			if wf.processExecEvent(execCtx) {
				return
			}
		}
	}
}

// Run 执行工作流6
func (wf *Workflow) Run() {
	go func() {
		defer func() {
			log := wf.logger.WithField(logger.Stage, "workflow_finish")
			success := 1
			if wf.err != nil {
				success = 0
				log = log.WithField(logger.Error, wf.err.Error())
			}
			log.WithField(logger.Success, success).Info("workflow finish")
		}()

		wf.run()
	}()
}

func (wf *Workflow) parseMeta(metaFile string) error {
	data, err := ioutil.ReadFile(metaFile)
	if err != nil {
		return err
	}
	m := meta.WorkflowMeta{}
	if err = json.Unmarshal(data, &m); err != nil {
		return err
	}
	wf.Meta = m

	return nil
}

func (wf *Workflow) parseParameter(parameter interface{}) error {
	data, err := json.Marshal(parameter)
	if err != nil {
		return err
	}
	wf.Context.Parameter = data

	m := make(map[string]interface{}, 0)
	if err = json.Unmarshal(data, &m); err != nil {
		return err
	}
	wf.Context.Config = config.NewConfig(m)

	return nil
}

func (wf *Workflow) parseParameterStr(parameter string) error {
	wf.Context.Parameter = []byte(parameter)
	m := make(map[string]interface{}, 0)
	if err := json.Unmarshal(wf.Context.Parameter, &m); err != nil {
		return err
	}
	wf.Context.Config = config.NewConfig(m)

	return nil
}

func (wf *Workflow) parseTasks() error {
	if len(wf.Meta.Tasks) < 1 {
		return fmt.Errorf("must have more than one task")
	}
	entry, err := newComponent(wf.Meta.Tasks[0], wf, 1)
	if err != nil {
		return err
	}
	curTask := entry
	wf.EntryTask = entry
	wf.Tasks.addTask(entry)

	for i := 1; i < len(wf.Meta.Tasks); i++ {
		tm := wf.Meta.Tasks[i]
		t, err := newComponent(tm, wf, i+1)
		if err != nil {
			return err
		}
		t.LinkPrev(curTask.TaskStep())
		curTask.LinkNext(t.TaskStep())
		wf.Tasks.addTask(t)
		curTask = t
	}

	return nil
}

// LoadWorkflow 加载工作流
func LoadWorkflow(id string) (*Workflow, error) {
	wf := &Workflow{
		Id: id,
	}

	err := wf.load(logger.R.WithField(logger.WorkflowId, id))
	return wf, err
}

// StartWorkflowFailed 启动工作流失败
func StartWorkflowFailed(wf *dao.Workflow) error {
	// 更新workflow为error, 更新对应fleet状态为error
	wf.State = dao.WorkflowStateError

	// 可忽略错误
	err := dao.UpdateWorkflow(wf, "State")
	if err != nil {
		logger.R.Warn("update workflow state to error db error: %v", err)
	}

	f := &dao.Fleet{
		Id:         wf.ResourceId,
		State:      dao.FleetStateError,
		UpdateTime: time.Now().UTC(),
	}
	if err = dao.GetFleetStorage().Update(f, "State", "UpdateTime"); err != nil {
		return err
	}

	return nil
}

// CreateWorkflow 新建workflow
func CreateWorkflow(metaFile string,
	parameter interface{},
	resourceId string,
	projectId string,
	log *logger.FMLogger,
	worknodeId string) (*Workflow, error) {
	u, _ := uuid.NewUUID()
	wf := &Workflow{
		Id:         u.String(),
		Tasks:      newTaskCollection(),
		execCh:     make(chan *directer.ExecuteContext, DefaultExecChLength),
		logger:     log.WithField(logger.WorkflowId, u.String()),
		Context:    &directer.WorkflowContext{},
		projectId:  projectId,
		resourceId: resourceId,
		worknodeId: worknodeId,
	}
	if err := wf.parseMeta(metaFile); err != nil {
		return nil, err
	}
	if err := wf.parseParameter(parameter); err != nil {
		return nil, err
	}

	if err := wf.parseTasks(); err != nil {
		return nil, err
	}

	if err := wf.insertDB(); err != nil {
		return nil, err
	}

	return wf, nil
}
