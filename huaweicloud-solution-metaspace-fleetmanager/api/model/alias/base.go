// Copyright (c) Huawei Technologies Co., Ltd. 2022-2022. All rights reserved.

// alias基本结构体定义
package alias

type AssociatedFleet struct {
	FleetId string  `json:"fleet_id,omitempty" validate:"omitempty,min=0,max=64"`
	Weight  float32 `json:"weight" validate:"min=0,max=1"`
}

type Alias struct {
	AliasId          string            `json:"alias_id"`
	Name             string            `json:"name"`
	Description      string            `json:"description"`
	CreationTime     string            `json:"creation_time"`
	AssociatedFleets []AssociatedFleet `json:"associated_fleets"`
	Type            string            `json:"type" validate:"required,oneof=ACTIVE DEACTIVE TERMINATED"`
	Message          string            `json:"message,omitempty" validate:"omitempty,min=0,max=1024"`
}
