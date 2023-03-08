// Copyright (c) Huawei Technologies Co., Ltd. 2022-2022. All rights reserved.

// alias删除方法
package alias

import (
	"encoding/json"
	"fleetmanager/api/errors"
	"fleetmanager/api/model/alias"
	FleetModel "fleetmanager/api/model/fleet"
	"fleetmanager/api/params"
	"fleetmanager/api/service/constants"
	"fleetmanager/api/service/fleet"
	"fleetmanager/db/dao"
	"sync"

	"github.com/beego/beego/v2/client/orm"
)

// Show: 查询alias详情
func (s *Service) ShowAlias() (*alias.ShowAliasResponse, *errors.CodedError) {

	aliasId := s.Ctx.Input.Param(params.AliasId)
	projectId := s.Ctx.Input.Param(params.ProjectId)
	filter := dao.Filters{
		"Id":        aliasId,
		"ProjectId": projectId,
	}
	a, err := dao.GetAliasStorage().Get(filter)
	if err != nil {
		s.Logger.Error("show alias db error: %v", err)
		if err == orm.ErrNoRows {
			return nil, errors.NewError(errors.AliasNotFound)
		}
		return nil, errors.NewError(errors.DBError)
	}
	associatedFleets, errCode := GenerateAliasFleetRsp(*s, *a)
	if errCode != nil {
		return nil, errCode
	}
	rsp := GenerateAliasRsp(*a, *associatedFleets)
	return &rsp, nil
}

// List: 查询alias列表na
func (s *Service) List(offset int, limit int) (alias.List, *errors.CodedError) {

	var list alias.List
	total_count, err := dao.GetAliasStorage().Count(s.aliasFilter())
	if err != nil {
		s.Logger.Error("list alias count from db error: %v", err)
		return list, errors.NewError(errors.DBError)
	}
	list.TotalCount = int(total_count)
	if total_count == 0 {
		list.Alias = []alias.ShowAliasResponse{}
		return list, nil
	}
	if e := s.checkOffsetLimit(total_count, offset, limit); e != nil {
		return list, e
	}
	aliases, err1 := dao.GetAliasStorage().List(s.aliasFilter(), (offset)*limit, limit)
	if err1 != nil {
		s.Logger.Error("list alias from db error: %v", err1)
		return list, errors.NewError(errors.DBError)
	}
	list.Count = len(aliases)
	for _, a := range aliases {
		associatedFleets, err := GenerateAliasFleetRsp(*s, a)
		if err != nil {
			return list, errors.NewErrorF(errors.ServerInternalError, err.Error())
		}
		list.Alias = append(list.Alias, GenerateAliasRsp(a, *associatedFleets))
	}
	return list, nil
}

// checkOffsetLimit: 校验分页不超总数
func (s *Service) checkOffsetLimit(count int64, offset int, limit int) *errors.CodedError {
	if int64(offset*limit) >= count {
		return errors.NewErrorF(errors.InvalidParameterValue, " offset and limit over total count")
	}
	return nil
}

// aliasFilter: 过滤条件查询
func (s *Service) aliasFilter() dao.Filters {
	projectId := s.Ctx.Input.Param(params.ProjectId)
	filterMap := make(map[string]interface{})
	filterMap["project_id"] = projectId
	if AliasType := s.Ctx.Input.Query(params.QueryType); AliasType != "" {
		filterMap["type"] = AliasType
	} else {
		filterMap["type__ne"] = dao.AliasTypeTerminated
	}
	if AliasName := s.Ctx.Input.Query(params.QueryName); AliasName != "" {
		filterMap["name__contains"] = AliasName
	}
	if AliasFleetId := s.Ctx.Input.Query(params.QueryFleetId); AliasFleetId != "" {
		filterMap["associated_fleets__contains"] = AliasFleetId
	}
	return filterMap
}

func GenerateAliasFleetRsp(s Service, als dao.Alias) (*[]alias.AssociatedFleetRsp, *errors.CodedError) {
	fleetService := fleet.NewFleetService(s.Ctx, s.Logger)
	aFleets := []alias.AssociatedFleetRsp{}

	// 格式化关联fleet信息
	associatedFleets := &[]alias.AssociatedFleet{}
	err := json.Unmarshal([]byte(als.AssociatedFleets), associatedFleets)
	if err != nil {
		return nil, errors.NewErrorF(errors.ServerInternalError, err.Error())
	}

	// 异步获取每个alias所关联的alias的信息
	wg := sync.WaitGroup{}
	wg.Add(len(*associatedFleets))
	for _, aft := range *associatedFleets {
		ft, err := dao.GetFleetStorage().Get(dao.Filters{"Id": aft.FleetId})
		if err != nil {
			return nil, errors.NewError(errors.DBError)
		}
		go func(oneFt FleetModel.Fleet, aaft alias.AssociatedFleet) {
			defer wg.Done()
			mftp, err := fleetService.GetMonitorFleet(&oneFt)
			if err != nil {
				s.Logger.Info("get fleet: %s info error: %+v", oneFt.FleetId, err)
			} else {
				aFleets = append(aFleets, GenerateAssociatedFleetRsp(*mftp, aaft))
			}
		}(fleet.BuildFleetModel(ft), aft)
	}
	wg.Wait()
	return &aFleets, nil
}

func GenerateAssociatedFleetRsp(ft FleetModel.ShowMonitorFleetResponse, af alias.AssociatedFleet) alias.AssociatedFleetRsp {
	return alias.AssociatedFleetRsp{
		Name:                ft.Name,
		FleetId:             ft.FleetId,
		Weight:              af.Weight,
		State:               ft.State,
		InstanceCount:       ft.InstanceCount,
		ProcessCount:        ft.ProcessCount,
		ServerSessionCount:  ft.ServerSessionCount,
		MaxServerSessionNum: ft.MaxServerSessionNum,
	}
}

func GenerateAliasRsp(als dao.Alias, alsrsp []alias.AssociatedFleetRsp) alias.ShowAliasResponse {
	if len(alsrsp) == 0 {
		alsrsp = []alias.AssociatedFleetRsp{}
	}
	return alias.ShowAliasResponse{
		AliasId:          als.Id,
		Name:             als.Name,
		Description:      als.Description,
		CreationTime:     als.CreationTime.Format(constants.TimeFormatLayout),
		UpdateTime:       als.UpdateTime.Format(constants.TimeFormatLayout),
		AssociatedFleets: alsrsp,
		Type:             als.Type,
		Message:          als.Message,
	}
}
