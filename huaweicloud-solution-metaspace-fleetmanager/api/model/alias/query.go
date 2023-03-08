// Copyright (c) Huawei Technologies Co., Ltd. 2022-2022. All rights reserved.

// alias查询结构体定义
package alias

type List struct {
	TotalCount int                 `json:"total_count"`
	Count      int                 `json:"count"`
	Alias      []ShowAliasResponse `json:"aliases"`
}

type ShowAliasResponse struct {
	AliasId          string               `json:"alias_id"`
	Name             string               `json:"name"`
	Description      string               `json:"description"`
	CreationTime     string               `json:"creation_time"`
	UpdateTime		 string				  `json:"update_time"`
	AssociatedFleets []AssociatedFleetRsp `json:"associated_fleets"`
	Type            string                `json:"type"`
	Message          string               `json:"message,omitempty" validate:"omitempty,min=0,max=1024"`
}

// AssociatedFleetRs: 别名关联fleet响应
type AssociatedFleetRsp struct {
	Name               string  `json:"name"`
	FleetId            string  `json:"fleet_id"`
	Weight             float32 `json:"weight"`
	State              string  `json:"state"`
	InstanceCount      int     `json:"instance_count"`
	ProcessCount       int     `json:"process_count"`
	ServerSessionCount int     `json:"server_session_count"`
	MaxServerSessionNum int		`json:"max_server_session_count"`
}

type MessageRsp struct {
	Message string `json:"message"`
}
