package lts

import (
	"fleetmanager/api/errors"
	"fleetmanager/api/service/constants"
	"fleetmanager/db/dao"

	"fleetmanager/client"
	"fmt"
	"net/http"
)

// 删除日志接入
func (s *LtsService) DeleteTransfer(projectId string, transferId string) (code int,
	rsp []byte, e *errors.CodedError) {
	var projectInfo dao.ResProject
	err := dao.Filters{"OriginProjectId": projectId}.Filter(dao.ResProjectTable).One(&projectInfo)
	if err != nil {
		s.logger.Error("get projectid err:%+v", err)
		return 0, nil, errors.NewError(errors.DBError)
	}

	url := client.GetServiceEndpoint(client.ServiceNameAASS, projectInfo.Region) +
		fmt.Sprintf(constants.ScalingGroupLtsLogTransfer, projectInfo.ResProjectId)
	req := client.NewRequest(client.ServiceNameAASS, url, http.MethodDelete, nil)
	req.SetQuery("log_transfer_id", transferId)
	code, rsp, err = req.DoRequest()
	if code < http.StatusOK || code >= http.StatusBadRequest {
		return code, rsp, errors.NewErrorF(errors.LtsAccessConfigError, string(rsp))
	}
	return code, rsp, nil
}

// 删除日志接入
func (s *LtsService) DeleteAccessConfig(projectId string, accessconfigId string)  (code int,
	rsp []byte, e *errors.CodedError) {
	var projectInfo dao.ResProject
	err := dao.Filters{"OriginProjectId": projectId}.Filter(dao.ResProjectTable).One(&projectInfo)
	if err != nil {
		s.logger.Error("get projectid err:%+v", err)
		return 0, nil, errors.NewError(errors.DBError)
	}

	url := client.GetServiceEndpoint(client.ServiceNameAASS, projectInfo.Region) +
		fmt.Sprintf(constants.ScalingGroupLtsAccessConfig, projectInfo.ResProjectId)
	req := client.NewRequest(client.ServiceNameAASS, url, http.MethodDelete, nil)
	req.SetQuery("access_config_id", accessconfigId)
	code, rsp, err = req.DoRequest()
	if code < http.StatusOK || code >= http.StatusBadRequest {
		return code, rsp, errors.NewErrorF(errors.LtsAccessConfigError, string(rsp))
	}
	return code, rsp, nil
}
