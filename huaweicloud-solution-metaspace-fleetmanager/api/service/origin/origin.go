// Copyright (c) Huawei Technologies Co., Ltd. 2022-2022. All rights reserved.

// 获取origin信息
package origin

import (
	"encoding/json"
	"fleetmanager/api/errors"
	"fleetmanager/api/model/user"
	"fleetmanager/db/dao"
	"fleetmanager/logger"
	"fleetmanager/setting"
	"net/http"
	"strings"

	"github.com/beego/beego/v2/server/web/context"
)

type OriginService struct {
	ctx                  *context.Context
	logger               *logger.FMLogger
}

func NewOriginService(ctx *context.Context, logger *logger.FMLogger) *OriginService {
	s := &OriginService{
		ctx:    ctx,
		logger: logger,
	}
	return s
}
func (s *OriginService) List() (code int, rsp []byte, e *errors.CodedError) {
	s.logger.Info("receive a request to list origin info")
	username, err := s.CheckUser()
	if err != nil {
		return 400, nil, err
	}
	queryParams := make(map[string]string)
	queryParams["username"] = username
	resConf, errNew := dao.GetAllResConfigByMap(queryParams)
	if errNew != nil {
		return 500, nil, errors.NewError(errors.ServerInternalError)
	}
	totalCount := len(resConf)
	queryParams["region"] = s.ctx.Input.Query("region")
	queryParams["origin_domain_name"] = s.ctx.Input.Query("domain_name")
	resConf, errNew = dao.GetAllResConfigByMap(queryParams)
	if errNew != nil {
		return 500, nil, errors.NewError(errors.ServerInternalError)
	}
	originDomainIds := make(map[string]user.OriginInfoResponse)
	originResp := &user.ListOriginInfoResponse{
		TotalCount: totalCount,		// 该用户下支持的所有Domain所有区域的总数量
		Count: 		0,					// 该用户下domain用户数量
		Origins: 	originDomainIds,
	}
	for _, res := range resConf {
		GenerateByUserConf(&res, originResp)
	}
	originResp.Count = len(originResp.Origins)
	rsp, errNew = json.Marshal(originResp)
	if errNew != nil {
		s.logger.Error("marshal response error: %v, rsp: %s", errNew, originResp)
		return 500, nil, errors.NewError(errors.ServerInternalError)
	}
	return http.StatusOK, rsp, nil
}

func (s *OriginService) ListRegions() (code int, rsp []byte, e *errors.CodedError) {
	supportRegions := setting.SupportRegions
	listRegions := strings.Split(supportRegions, ",")
	if len(listRegions) <= 0 {
		s.logger.Error("support regions is not existed")
		return 0, nil, errors.NewErrorF(errors.ServerInternalError, "support regions is not existed")
	}
	listRegionsRsp := &user.ListRegions{
		Count:	len(listRegions),
		Regions: listRegions,
	}
	rsp, err := json.Marshal(listRegionsRsp)
	if err != nil {
		s.logger.Error("marshal response error: %v, rsp: %s", err, listRegionsRsp)
		return 500, nil, errors.NewError(errors.ServerInternalError)
	}
	return 200, rsp, nil
}

func (s *OriginService) CheckUser() (string, *errors.CodedError) {
	user := s.ctx.Input.Query("username")
	if user != "" {
		return user, nil
	}
	return user, errors.NewError(errors.UserNotFound)
}

func GenerateByUserConf(res *dao.UserResConf, resp *user.ListOriginInfoResponse) {
	if _, ok := resp.Origins[res.OriginDomainId]; !ok {
		originRegion := make(map[string]user.OriginProject)
		originInfo := &user.OriginInfoResponse{
			OriginDomainName: 	res.OriginDomainName,
			OriginDomainId:		res.OriginDomainId,
			Region: 			originRegion,
		}
		resp.Origins[res.OriginDomainId] = *originInfo
	}
	originProject := &user.OriginProject{
		OriginProjectId: res.OriginProjectId,
	}
	resp.Origins[res.OriginDomainId].Region[res.Region] = *originProject
}