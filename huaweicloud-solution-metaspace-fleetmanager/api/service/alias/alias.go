// Copyright (c) Huawei Technologies Co., Ltd. 2022-2022. All rights reserved.

// alias 操作方法
package alias

import (
	"encoding/json"
	"fleetmanager/api/errors"
	"fleetmanager/api/model/alias"
	"fleetmanager/api/params"
	"fleetmanager/api/service/base"
	"fleetmanager/db/dao"
	"fleetmanager/logger"
	"fmt"
	"math/rand"
	"time"

	"github.com/beego/beego/v2/server/web/context"
	"github.com/google/uuid"
	"github.com/mroth/weightedrand"
)

type Service struct {
	base.FleetService
	createRequest *alias.CreateRequest
	updateReq     *alias.UpdateAliasRequest
	alias         *dao.Alias
}

// fleet的权重最多支持三位小数
const WeightPrecision = 1000

// NewAliasService 新建Alisa管理服务
func NewAliasService(ctx *context.Context, logger *logger.FMLogger) *Service {
	s := &Service{
		FleetService: base.FleetService{
			Ctx:    ctx,
			Logger: logger,
		},
	}
	return s
}

// buildAlias: 构建别名结构
func (s *Service) buildAlias() error {
	u, _ := uuid.NewUUID()
	aliasId := u.String()
	associatedFleetByte, err := json.Marshal(s.createRequest.AssociatedFleets)
	if err != nil {
		return fmt.Errorf("marshal associated fleets: %+v err: %+v", s.createRequest.AssociatedFleets, err)
	}
	a := &dao.Alias{
		Id:           aliasId,
		ProjectId:    s.Ctx.Input.Param(params.ProjectId),
		Name:         s.createRequest.Name,
		Description:  s.createRequest.Description,
		CreationTime: time.Now().Local(),
		UpdateTime:   time.Now().Local(),
		AssociatedFleets: string(associatedFleetByte),
		Type:         s.createRequest.Type,
		Message:      s.createRequest.Message,
	}
	s.alias = a
	return nil
}

// setAlias: 检查Alias是否存在
func (s *Service) setAlias() error {
	aliasId := s.Ctx.Input.Param(params.AliasId)
	projectId := s.Ctx.Input.Param(params.ProjectId)
	filter := dao.Filters{
		"Id":        aliasId,
		"ProjectId": projectId,
	}
	f, err := dao.GetAliasStorage().Get(filter)
	if err != nil {
		return err
	}
	s.alias = f
	return nil
}

// 根据alias关联的fleet权重判断会话应该在哪个fleet下创建，不判断fleet状态
func GenerateFleetByAssociateFleetWeight(afts string) (string, *errors.CodedError) {
	associatedFleets := &[]alias.AssociatedFleet{}
	err := json.Unmarshal([]byte(afts), associatedFleets)
	if err != nil {
		return "", errors.NewErrorF(errors.ServerInternalError, err.Error())
	}
	if len(*associatedFleets) < 1 {
		return "", errors.NewError(errors.AliasNoAvailableFleet)
	}
	// 随机种子
	rand.Seed(time.Now().UTC().UnixNano())
	availableFleetsChoices := []weightedrand.Choice{}
	for _, aft := range *associatedFleets {
		availableFleetsChoices = append(availableFleetsChoices, 
			weightedrand.NewChoice(aft.FleetId, uint(aft.Weight*WeightPrecision)))
	}
	chooser, err := weightedrand.NewChooser(availableFleetsChoices...)
	if err != nil {
		return "", errors.NewErrorF(errors.ServerInternalError, err.Error())
	}
	return chooser.Pick().(string), nil
}