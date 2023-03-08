package lts

import (
	"encoding/json"
	"fleetmanager/api/errors"
	ltsmodel "fleetmanager/api/model/lts"
	"fleetmanager/api/service/base"
	"fleetmanager/api/service/constants"
	"fleetmanager/db/dao"

	"fleetmanager/client"
	"fleetmanager/logger"
	"fmt"
	"net/http"

	"github.com/beego/beego/v2/server/web/context"
)

type FleetIdList struct {
	Count       int64
	FleetIdList []string
}

type LtsService struct {
	base.FleetService
	ctx    *context.Context
	logger *logger.FMLogger
}

func NewLtsService(ctx *context.Context, logger *logger.FMLogger) *LtsService {
	s := &LtsService{
		ctx:    ctx,
		logger: logger,
	}
	return s
}

// 创建日志接入
func (s *LtsService) CreateLTSAccessConfig(projectId string,
	acConfig ltsmodel.CreateAccessConfigReq) (*ltsmodel.CreateAccessConfigResp, *errors.CodedError) {
	
	var projectInfo dao.ResProject
	err := dao.Filters{"OriginProjectId": projectId}.Filter(dao.ResProjectTable).One(&projectInfo)
	if err != nil {
		s.logger.Error("get projectid err:%+v", err)
		return nil, errors.NewError(errors.DBError)
	}
	var iamInfo dao.ResAgency
	err = dao.Filters{"Id": projectInfo.Id}.Filter(dao.ResAgencyTable).One(&iamInfo)
	if err != nil {
		s.logger.Error("get res agency err:%+v", err)
		return nil, errors.NewError(errors.DBError)
	}
	// 需要配置了iam agency才能创建
	if iamInfo.IamAgencyName == "" {
		s.logger.Error("get iam agency name err:%+v", err)
		return nil, errors.NewErrorF(errors.DBError, "IAM Agency invalid or not set")
	}
	acConfigToAASS, errC := s.BuildCreateAccessConfigReqToAASS(acConfig, projectInfo.OriginProjectId)
	if errC != nil {
		return nil, errC
	} 
	body, err := json.Marshal(acConfigToAASS)
	if err != nil {
		return nil, errors.NewError(errors.InvalidParameterValue)
	}
	url := client.GetServiceEndpoint(client.ServiceNameAASS, projectInfo.Region) +
		fmt.Sprintf(constants.ScalingGroupLtsAccessConfig, projectInfo.ResProjectId)
	req := client.NewRequest(client.ServiceNameAASS, url, http.MethodPost, body)
	code, rsp, err := req.DoRequest()
	respMsg := &ltsmodel.CreateAccessConfigResp{}
	e := base.AcceptResp(s.logger, &respMsg, code, rsp, err)
	if e != nil {
		s.logger.Error("CreateLTSAccessConfig err:%+v", e.ErrC.Msg())
		return nil, e
	}
	return respMsg, e
}

func (s *LtsService) BuildCreateAccessConfigReqToAASS(acConfig ltsmodel.CreateAccessConfigReq, projectId string) (
	*ltsmodel.CreateAccessConfigReqToAASS, *errors.CodedError) {
	filter := dao.Filters{
		"Id":         acConfig.FleetId,
		"ProjectId":  projectId,
		"Terminated": false,
	}
	f, err := dao.GetFleetStorage().Get(filter)
	if err != nil {
		s.logger.Error("get fleet info db error: %s", err.Error())
		return nil, errors.NewError(errors.DBError)
	}
	return &ltsmodel.CreateAccessConfigReqToAASS{
		CreateAccessConfigReq: acConfig,
		EnterpriseProjectId: f.EnterpriseProjectId,
	}, nil
}
// 创建日志转储
func (s *LtsService) CreateTransfer(projectId string, config ltsmodel.LogTransferReq) (*ltsmodel.TransferResp,
	*errors.CodedError) {
	var projectInfo dao.ResProject
	err := dao.Filters{"OriginProjectId": projectId}.Filter(dao.ResProjectTable).One(&projectInfo)
	if err != nil {
		s.logger.Error("get projectid err:%+v", err)
		return nil, errors.NewError(errors.DBError)
	}

	body, err := json.Marshal(config)
	if err != nil {
		return nil, nil
	}
	url := client.GetServiceEndpoint(client.ServiceNameAASS, projectInfo.Region) +
		fmt.Sprintf(constants.ScalingGroupLtsLogTransfer, projectInfo.ResProjectId)
	req := client.NewRequest(client.ServiceNameAASS, url, http.MethodPost, body)
	code, rsp, err := req.DoRequest()
	respMsg := ltsmodel.TransferResp{}
	e := base.AcceptResp(s.logger, &respMsg, code, rsp, err)
	if e != nil {
		s.logger.Error("CreateTransfer err:%+v", e.ErrC.Msg())
		return nil, e
	}
	return &respMsg, nil
}

// 创建日志组
func (s *LtsService) CreateLogGroup(projectId string, config ltsmodel.CreateLogGroup) (*ltsmodel.CreateLogGroupResp, *errors.CodedError) {
	var projectInfo dao.ResProject
	err := dao.Filters{"OriginProjectId": projectId}.Filter(dao.ResProjectTable).One(&projectInfo)
	if err != nil {
		s.logger.Error("get projectid err:%+v", err)
		return nil, errors.NewError(errors.DBError)
	}

	body, err := json.Marshal(config)
	if err != nil {
		return nil, nil
	}
	url := client.GetServiceEndpoint(client.ServiceNameAASS, projectInfo.Region) +
		fmt.Sprintf(constants.ScalingGroupLtsLogGroup, projectInfo.ResProjectId)
	req := client.NewRequest(client.ServiceNameAASS, url, http.MethodPost, body)
	code, rsp, err := req.DoRequest()
	respMsg := ltsmodel.CreateLogGroupResp{}
	e := base.AcceptResp(s.logger, &respMsg, code, rsp, err)
	if e != nil {
		s.logger.Error("CreateLogGroup err:%+v", e.ErrC.Msg())
		return nil, e
	}
	return &respMsg, nil
}
