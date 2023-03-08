// Copyright (c) Huawei Technologies Co., Ltd. 2022-2022. All rights reserved.

// 应用包通用方法
package build

import (
	"fleetmanager/api/model/build"
	"fleetmanager/api/service/constants"
	"fleetmanager/db/dao"
)

// @Title buildBuildModel
// @Description
// @Author wangnannan 2022-05-07 09:10:52 ${time}
// @Param bd
// @Return build.Build
func buildBuildModel(bd *dao.Build) build.Build {
	b := build.Build{
		BuildId:         bd.Id,
		Name:            bd.Name,
		Description:     bd.Description,
		State:           bd.State,
		CreationTime:    bd.CreationTime.Format(constants.TimeFormatLayout),
		Version:         bd.Version,
		Size:            bd.Size,
		OperatingSystem: bd.OperatingSystem,
	}

	return b
}

func buildFleetModel(f []dao.Fleet) []build.FleetMsg {
	var res []build.FleetMsg
	for _, val := range f {
		fm := build.FleetMsg{
			FleetId:      val.Id,
			FleetName:    val.Name,
			FleetState:   val.State,
			CreationTime: val.CreationTime.Format(constants.TimeFormatLayout),
		}
		res = append(res, fm)
	}
	return res
}

func buildFullBuild(bd *dao.Build) build.FullBuild {
	b := build.FullBuild{
		Id:                bd.Id,
		Name:              bd.Name,
		Description:       bd.Description,
		State:             bd.State,
		CreationTime:      bd.CreationTime.Format(constants.TimeFormatLayout),
		UpdateTime:        bd.UpdateTime.Format(constants.TimeFormatLayout),
		Version:           bd.Version,
		Size:              bd.Size,
		ProjectId:         bd.ProjectId,
		OperatingSystem:   bd.OperatingSystem,
		ImageId:           bd.ImageId,
		ImageRegion:       bd.ImageRegion,
		StorageBucketName: bd.StorageBucketName,
		StorageKey:        bd.StorageKey,
		StorageRegion:     bd.StorageRegion,
	}
	return b
}
