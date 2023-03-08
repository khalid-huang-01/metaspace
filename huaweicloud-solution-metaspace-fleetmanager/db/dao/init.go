// Copyright (c) Huawei Technologies Co., Ltd. 2022-2022. All rights reserved.

// 数据表初始化
package dao

import "github.com/beego/beego/v2/client/orm"

// Init 数据库表初始化
func Init() {
	orm.RegisterModel(new(Fleet))
	orm.RegisterModel(new(InboundPermission))
	orm.RegisterModel(new(RuntimeConfiguration))
	orm.RegisterModel(new(FleetEvent))
	orm.RegisterModel(new(Build))
	orm.RegisterModel(new(BuildImage))
	orm.RegisterModel(new(ScalingGroup))
	orm.RegisterModel(new(ScalingPolicy))
	orm.RegisterModel(new(FleetVpcCidr))
	orm.RegisterModel(new(ResDomain))
	orm.RegisterModel(new(ResAgency))
	orm.RegisterModel(new(UserAgency))
	orm.RegisterModel(new(ResUser))
	orm.RegisterModel(new(ResProject))
	orm.RegisterModel(new(ResKeypair))
	orm.RegisterModel(new(Workflow))
	orm.RegisterModel(new(WorkNode))
	orm.RegisterModel(new(FleetServerSession))
	orm.RegisterModel(new(Alias))
	orm.RegisterModel(new(User))
	orm.RegisterModel(new(UserResConf))
}
