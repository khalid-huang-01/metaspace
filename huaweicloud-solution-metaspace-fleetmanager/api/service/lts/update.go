package lts

import (
	"encoding/json"
	"fleetmanager/api/errors"
	ltsmodel "fleetmanager/api/model/lts"
	"fleetmanager/api/service/constants"
	"fleetmanager/client"
	"fleetmanager/db/dao"
	"fmt"
	"net/http"
)

// 更新主机组
func (s *LtsService) UpdateHostGroup(projectId string) *errors.CodedError {
	var projectInfo dao.ResProject
	err := dao.Filters{"OriginProjectId": projectId}.Filter(dao.ResProjectTable).One(&projectInfo)
	if err != nil {
		s.logger.Error("get projectid err:%+v", err)
		return errors.NewError(errors.DBError)
	}

	url := client.GetServiceEndpoint(client.ServiceNameAASS, projectInfo.Region) +
		fmt.Sprintf(constants.ScalingGroupLtsLogTransfer, projectInfo.ResProjectId)
	req := client.NewRequest(client.ServiceNameAASS, url, http.MethodPut, nil)
	_, _, err = req.DoRequest()

	if err != nil {
		s.logger.Error("update err:%+v", err.Error())
		return errors.NewError(errors.LtsLogGroupError)
	}
	return nil
}

// 更新日志接入
func (s *LtsService) UpdateAccessConfig(projectId string,
	config ltsmodel.UpdateAccessConfigToDB) (code int, rsp []byte, e *errors.CodedError) {
	var projectInfo dao.ResProject
	err := dao.Filters{"OriginProjectId": projectId}.Filter(dao.ResProjectTable).One(&projectInfo)
	if err != nil {
		s.logger.Error("get projectid err:%+v", err)
		return 0, nil, errors.NewError(errors.DBError)
	}
	body, err := json.Marshal(config)
	if err != nil {
		return 0, nil, errors.NewError(errors.ServerInternalError)
	}
	url := client.GetServiceEndpoint(client.ServiceNameAASS, projectInfo.Region) +
		fmt.Sprintf(constants.ScalingGroupLtsAccessConfig, projectInfo.ResProjectId)
	req := client.NewRequest(client.ServiceNameAASS, url, http.MethodPut, body)
	code, rsp, err = req.DoRequest()
	if code < http.StatusOK || code >= http.StatusBadRequest {
		return code, rsp, errors.NewErrorF(errors.LtsAccessConfigError, string(rsp))
	}
	return code, rsp, nil
}
