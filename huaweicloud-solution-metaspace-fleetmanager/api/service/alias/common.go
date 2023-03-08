// Copyright (c) Huawei Technologies Co., Ltd. 2022-2022. All rights reserved.

// alias基础方法
package alias

import (
	"encoding/json"
	"fleetmanager/api/model/alias"
	"fleetmanager/api/service/constants"
	"fleetmanager/db/dao"
	"fmt"
)

func buildAliasModel(a *dao.Alias) (*alias.Alias, error) {
	associatedFleet := []alias.AssociatedFleet{}
	err := json.Unmarshal([]byte(a.AssociatedFleets), &associatedFleet)
	if err != nil {
		return nil, fmt.Errorf("build alias model error: %+v", err)
	}
	if associatedFleet == nil {
		associatedFleet= []alias.AssociatedFleet{}
	}
	aliasModel := alias.Alias{
		AliasId:      a.Id,
		Name:         a.Name,
		Description:  a.Description,
		CreationTime: a.CreationTime.Format(constants.TimeFormatLayout),
		AssociatedFleets: associatedFleet,
		Type:    a.Type,
		Message: a.Message,
	}
	return &aliasModel, nil
}
