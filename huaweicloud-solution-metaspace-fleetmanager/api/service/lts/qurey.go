package lts

import (
	"fleetmanager/api/errors"
	ltsmodel "fleetmanager/api/model/lts"
	"fleetmanager/api/params"
	"fleetmanager/api/service/base"
	"fleetmanager/api/service/constants"
	"fleetmanager/client"
	"fleetmanager/db/dao"
	"fmt"
	"net/http"
	"strconv"
	"strings"
)

// list 日志组
func (s *LtsService) ListLogGroup(projectId string) (*ltsmodel.ListLogGroups, *errors.CodedError) {
	var projectInfo dao.ResProject
	err := dao.Filters{"OriginProjectId": projectId}.Filter(dao.ResProjectTable).One(&projectInfo)
	if err != nil {
		s.logger.Error("get projectid err:%+v", err)
		return nil, errors.NewError(errors.DBError)
	}

	url := client.GetServiceEndpoint(client.ServiceNameAASS, projectInfo.Region) +
		fmt.Sprintf(constants.ScalingGroupListLogGroup, projectInfo.ResProjectId)
	req := client.NewRequest(client.ServiceNameAASS, url, http.MethodGet, nil)
	code, rsp, err := req.DoRequest()
	respMsg := &ltsmodel.ListLogGroups{}
	e := base.AcceptResp(s.logger, &respMsg, code, rsp, err)
	if e != nil {
		s.logger.Error("ListLogGroup err:%+v", e.ErrC.Msg())
		return nil, e
	}
	return respMsg, nil
}

// list 日志接入
func (s *LtsService) ListAccessConfig(projectId string, limit int, offset int) (*ltsmodel.ListAccessConfig, *errors.CodedError) {
	var projectInfo dao.ResProject
	err := dao.Filters{"OriginProjectId": projectId}.Filter(dao.ResProjectTable).One(&projectInfo)
	if err != nil {
		s.logger.Error("get projectid err:%+v", err)
		return nil, errors.NewError(errors.DBError)
	}

	url := client.GetServiceEndpoint(client.ServiceNameAASS, projectInfo.Region) +
		fmt.Sprintf(constants.ScalingGroupListAccessConfig, projectInfo.ResProjectId)
	req := client.NewRequest(client.ServiceNameAASS, url, http.MethodGet, nil)
	req.SetQuery(params.QueryLimit, strconv.Itoa(limit))
	req.SetQuery(params.QueryOffset, strconv.Itoa(offset))
	code, rsp, err := req.DoRequest()
	respMsg := ltsmodel.ListAccessConfig{}
	e := base.AcceptResp(s.logger, &respMsg, code, rsp, err)
	if e != nil {
		s.logger.Error("ListLogGroup err:%+v", e.ErrC.Msg())
		return nil, e
	}
	for index, msg := range respMsg.AccessConfigList {
		if msg.ObsTransferPath != "" {
			obsBucketName := strings.Split(msg.ObsTransferPath, "/")[1]
			respMsg.AccessConfigList[index].ObsTransferPathLink = fmt.Sprintf(getObsEndpoint(projectInfo.Region), obsBucketName)
		}
		respMsg.AccessConfigList[index].LogStreamLink = fmt.Sprintf(getLogStreamEndpoint(projectInfo.Region)+constants.LogStreamPattern, msg.LogGroupId,
			msg.LogGroupName, msg.LogStreamId, msg.LogStreamName)
	}
	return &respMsg, nil
}

// list 日志转储
func (s *LtsService) ListLogTransfer(projectId string, limit int, offset int) (*ltsmodel.ListTransfersInfo, *errors.CodedError) {
	var projectInfo dao.ResProject
	err := dao.Filters{"OriginProjectId": projectId}.Filter(dao.ResProjectTable).One(&projectInfo)
	if err != nil {
		s.logger.Error("get projectid err:%+v", err)
		return nil, errors.NewError(errors.DBError)
	}

	url := client.GetServiceEndpoint(client.ServiceNameAASS, projectInfo.Region) +
		fmt.Sprintf(constants.ScalingGroupListLogTransfer, projectInfo.ResProjectId)
	req := client.NewRequest(client.ServiceNameAASS, url, http.MethodGet, nil)
	req.SetQuery(params.QueryLogStreamId, s.ctx.Input.Query(params.QueryLogStreamId))
	req.SetQuery(params.QueryLimit, strconv.Itoa(limit))
	req.SetQuery(params.QueryOffset, strconv.Itoa(offset))
	code, rsp, err := req.DoRequest()
	respMsg := ltsmodel.ListTransfersInfo{}
	e := base.AcceptResp(s.logger, &respMsg, code, rsp, err)
	if e != nil {
		s.logger.Error("ListLogGroup err:%+v", e.ErrC.Msg())
		return nil, e
	}
	return &respMsg, nil
}

// 查询单条日志转储
func (s *LtsService) QueryLogTransfer(projectId string, logStreamId string) (*ltsmodel.ListTransferResp, *errors.CodedError) {
	var projectInfo dao.ResProject
	err := dao.Filters{"OriginProjectId": projectId}.Filter(dao.ResProjectTable).One(&projectInfo)
	if err != nil {
		s.logger.Error("get projectid err:%+v", err)
		return nil, errors.NewError(errors.DBError)
	}

	url := client.GetServiceEndpoint(client.ServiceNameAASS, projectInfo.Region) +
		fmt.Sprintf(constants.ScalingGroupLtsLogTransfer, projectInfo.ResProjectId)
	req := client.NewRequest(client.ServiceNameAASS, url, http.MethodGet, nil)
	req.SetQuery("log_stream_id", logStreamId)
	code, rsp, err := req.DoRequest()
	respMsg := ltsmodel.ListTransferResp{}
	e := base.AcceptResp(s.logger, &respMsg, code, rsp, err)
	if e != nil {
		s.logger.Error("ListLogGroup err:%+v", e.ErrC.Msg())
		return nil, e
	}

	return &respMsg, nil
}

// 查询 单条日志接入
func (s *LtsService) QueryAccessConfig(projectId string, accessConfigId string) (*ltsmodel.CreateAccessConfigResp, *errors.CodedError) {
	var projectInfo dao.ResProject
	err := dao.Filters{"OriginProjectId": projectId}.Filter(dao.ResProjectTable).One(&projectInfo)
	if err != nil {
		s.logger.Error("get projectid err:%+v", err)
		return nil, errors.NewError(errors.DBError)
	}

	url := client.GetServiceEndpoint(client.ServiceNameAASS, projectInfo.Region) +
		fmt.Sprintf(constants.ScalingGroupLtsAccessConfig, projectInfo.ResProjectId)
	req := client.NewRequest(client.ServiceNameAASS, url, http.MethodGet, nil)
	req.SetQuery("access_config_id", accessConfigId)
	code, rsp, err := req.DoRequest()
	respMsg := ltsmodel.CreateAccessConfigResp{}
	e := base.AcceptResp(s.logger, &respMsg, code, rsp, err)
	if e != nil {
		s.logger.Error("QueryAccessConfig err:%+v", e.ErrC.Msg())
		return nil, e
	}
	return &respMsg, nil
}

func getObsEndpoint(region string) string {
	if region == "cn-north-7" {
		return fmt.Sprintf("%s/console/#%s", constants.UlanqabConsoleEndpoint, constants.ObsListPattern)
	}
	return fmt.Sprintf("%s/console/#%s", constants.NormalConsoleEndpoint, constants.ObsListPattern)
}

func getLogStreamEndpoint(region string) string {
	if region == "cn-north-7" {
		return fmt.Sprintf("%s/lts/#/cts/logEventsLeftMenu/events?", constants.UlanqabConsoleEndpoint)
	}
	return fmt.Sprintf("%s/lts/#/cts/logEventsLeftMenu/events?", constants.NormalConsoleEndpoint)
}
