// Copyright (c) Huawei Technologies Co., Ltd. 2022-2022. All rights reserved.

// 应用包更新方法
package build

import (
	"encoding/json"
	"fleetmanager/api/errors"
	"fleetmanager/api/model/build"
	"fleetmanager/api/params"
	"fleetmanager/db/dao"
	"fleetmanager/db/dbm"
	"github.com/beego/beego/v2/server/web/context"
)

// @Title Update
// @Description  Update build info
// @Author wangnannan 2022-05-07 10:10:10 ${time}
// @Param ctx
// @Param r
// @Return error
func (s *Service) Update(ctx *context.Context, r build.UpdateRequest) (dao.Build, error) {

	b := dao.Build{}
	buildId := ctx.Input.Param(params.BuildId)
	projectId := ctx.Input.Param(params.ProjectId)
	err := dbm.Ormer.QueryTable(dao.BuildTable).Filter("Id", buildId).Filter("ProjectId", projectId).One(&b)
	if err != nil {
		return b, s.ErrorMsg(errors.BuildNotExists, "get build from db failed", err)
	}

	if b.Name != r.Name || b.Version != r.Version {
		if e := s.CheckBuildByName(r.Name, r.Version, projectId); e != nil {
			return b, e
		}
	}
	err = json.Unmarshal(ctx.Input.RequestBody, &b)
	if err != nil {
		return b, s.ErrorMsg(errors.BuildUpdateFailed, "Build update failed", err)
	}

	_, err = dbm.Ormer.Update(&b, "Name", "Version", "Description")
	if err != nil {
		return b, s.ErrorMsg(errors.BuildUpdateFailed,
			"Build update failed", err)
	}
	return b, nil

}
