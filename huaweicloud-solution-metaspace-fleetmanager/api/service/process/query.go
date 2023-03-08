// Copyright (c) Huawei Technologies Co., Ltd. 2022-2022. All rights reserved.

// 进程查询服务
package process

import (
	"encoding/json"
	"fleetmanager/api/errors"
	"fleetmanager/api/model/process"
	"fleetmanager/api/params"
	"fleetmanager/api/service/constants"
	"fleetmanager/logger"
	"fleetmanager/client"
	"fleetmanager/utils"
	"net/http"
	"fmt"
)

// transProcess 构建应用进程数据结构
func (s *Service) transProcess(ap process.AppProcess) process.Process {
	p := process.Process{
		Id:                  ap.Id,
		Ip:                  ap.IpAddress,
		Port:                ap.Port,
		State:               ap.State,
		ServerSessionCount:  ap.ServerSessionCount,
		MaxServerSessionNum: ap.MaxServerSessionNum,
	}
	return p
}

// transProcesses 数据结构转换
func (s *Service) transProcesses(appProcesses []process.AppProcess) []process.Process {
	pList := make([]process.Process, len(appProcesses))
	for i, app := range appProcesses {
		pList[i] = s.transProcess(app)
	}

	return pList
}

// List 查询应用进程列表
func (s *Service) List() (code int, rsp []byte, e *errors.CodedError) {
	if err := s.SetFleetById(s.Ctx.Input.Query(params.QueryFleetId)); err != nil {
		s.Logger.Error("get fleet in list server process error, fleetId:%s",
			s.Ctx.Input.Query(params.QueryFleetId))
		e = err
		return
	}
	code, rsp, err := s.forwardListToAPPGW()
	if err != nil {
		s.Logger.Error("forward get request to app gateway error %v", err)
		e = errors.NewErrorF(errors.ServerInternalError, "internal network error")
		return
	}
	// appgw接口返回数据结构
	appProcess := process.ListAppProcessesResponse{}
	if err := json.Unmarshal(rsp, &appProcess); err != nil {
		s.Logger.Error("read appgw body error %v", err)
		return 0, nil, errors.NewError(errors.ServerInternalError)
	}
	if appProcess.Count == 0 {
		r := &process.ListProcess{
			Count:     0,
			Processes: []process.Process{},
		}
		rsp, err := json.Marshal(r)
		if err != nil {
			s.Logger.Error("read appgw body error %v", err)
			return 0, nil, errors.NewError(errors.ServerInternalError)
		}
		return code, rsp, nil
	}

	p := &process.ListProcess{
		Count:     appProcess.Count,
		Processes: s.transProcesses(appProcess.AppProcesses),
	}
	rsp, err = json.Marshal(p)
	if err != nil {
		s.Logger.Error("read appgw body error %v", err)
		return 0, nil, errors.NewError(errors.ServerInternalError)
	}
	return code, rsp, nil
}
func (s *Service) forwardListToAPPGW() (code int, rsp []byte, err error) {

	url := client.GetServiceEndpoint(client.ServiceNameAPPGW, s.Fleet.Region) + constants.ProcessesUrl
	req := client.NewRequest(client.ServiceNameAPPGW, url, http.MethodGet, nil)
	req.SetQuery(params.QueryFleetId, s.Ctx.Input.Query(params.QueryFleetId))
	req.SetQuery(params.QueryOffset,
		utils.GetStringIfNotEmpty(s.Ctx.Input.Query(params.QueryOffset), params.DefaultOffset))
	req.SetQuery(params.QueryLimit, utils.GetStringIfNotEmpty(s.Ctx.Input.Query(params.QueryLimit), params.DefaultLimit))
	req.SetHeader(map[string]string{
		logger.RequestId: fmt.Sprintf("%s", s.Ctx.Input.GetData(logger.RequestId)),
	})
	return req.DoRequest()
}
