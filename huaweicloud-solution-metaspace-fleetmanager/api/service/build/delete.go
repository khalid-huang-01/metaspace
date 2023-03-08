// Copyright (c) Huawei Technologies Co., Ltd. 2022-2022. All rights reserved.

// 应用包删除方法
package build

import (
	"fleetmanager/api/errors"
	"fleetmanager/api/params"
	"fleetmanager/api/service/base"
	"fleetmanager/client"
	"fleetmanager/db/dao"
	"github.com/beego/beego/v2/server/web/context"
)

// @Title Delete
// @Description
// @Author wangnannan 2022-05-07 09:20:38 ${time}
// @Param ctx
// @Return *errors.CodedError
func (s *Service) Delete(ctx *context.Context) *errors.CodedError {

	projectId := ctx.Input.Param(params.ProjectId)
	buildId := ctx.Input.Param(params.BuildId)

	s.Logger.Info("buildId=%s   projectId=%s  delete begin!", buildId, projectId)

	tmpBuild, e := dao.GetBuildById(buildId, projectId)
	if e != nil {
		return s.ErrorMsg(errors.ServerInternalError,
			"Delete build Failed!Search DB failed!", e)
	}

	// check if build is related to any fleet
	err := s.CheckBuildInUsedSystem(buildId, projectId)
	if err != nil {
		return err
	}

	// if build created by IMS-image directly, just delete the record from db
	if tmpBuild.StorageBucketName != "" && tmpBuild.StorageKey != "" {

		// Delete build in IMS
		err = s.DeleteBuildInIMS(buildId, projectId)
		if err != nil {
			return err
		}
	}

	// Delete buildId record from db
	err = s.DeleteByBuild(buildId, projectId)
	if err != nil {
		return err
	}

	s.Logger.Info("buildId=%s   projectId=%s  delete success!", buildId, projectId)
	return nil
}

// @Title CheckBuildInUsedSystem
// @Description  if buildId is relate to a  fleet ,you must delete fleet first.
// @Author wangnannan 2022-05-07 09:20:53 ${time}
// @Param buildId
// @Return e
func (s *Service) CheckBuildInUsedSystem(buildId string, projectId string) (e *errors.CodedError) {

	res, err := dao.CheckBuildInUsedSystem(buildId, projectId)
	if err != nil {
		return s.ErrorMsg(errors.ServerInternalError,
			"Check build relate fleet Failed!Search DB failed!", err)
	}
	if res == false {
		return s.ErrorMsg(errors.BuildIsInUseNotSupportDelete,
			"This build is in used! Can not Delete!", err)
	}
	return nil
}

// @Title DeleteByBuild
// @Description  Delete by buildId
// @Author wangnannan 2022-05-07 09:21:06 ${time}
// @Param buildId
// @Param projectId
// @Return e
func (s *Service) DeleteByBuild(buildId string, projectId string) (e *errors.CodedError) {

	err := dao.DeleteByBuild(buildId, projectId)
	if err != nil {
		return s.ErrorMsg(errors.ServerInternalError, "Build Delete error!", err)
	}

	s.Logger.Info("BuildId %s delete from db success!", buildId)

	return nil
}

func (s *Service) DeleteBuildInIMS(buildId string, projectId string) (e *errors.CodedError) {

	s.Logger.Info("BuildId %s start to delete build in IMS!", buildId)

	// get build info failed delete directly
	tmpBuild, err := dao.GetBuildById(buildId, projectId)
	if err != nil {
		s.Logger.Warn("BuildId %s Get Build info failed!", buildId)
	}

	if tmpBuild.ImageId == "" {
		s.Logger.Warn("This build %s has no image in IMS", buildId)
		return nil
	}

	// get ResDomianInfo by projectId
	resDomainInfo, resError := base.GetResDomainInfo(projectId, tmpBuild.ImageRegion)
	if resError != nil {
		return resError
	}

	imsClient, imserror := client.GetAgencyIMSClient(tmpBuild.ImageRegion, resDomainInfo.ResProjectId,
		resDomainInfo.AgencyName, resDomainInfo.ResDomainId)
	if imserror != nil {
		return s.ErrorMsg(errors.ServerInternalError, "Delete build Get imsclient error!",
			imserror)
	}

	err = client.DeleteBuildInIMS(imsClient, tmpBuild.ImageId)
	if err != nil {
		return s.ErrorMsg(errors.ServerInternalError, "Delete build in IMS failed!", err)
	}

	return nil
}

// @Title DeleteBuildInObs
// @Description  Delete build and  bucket in obs
// @Author wangnannan 2022-05-07 09:21:19 ${time}
// @Param buildId
// @Param projectId
// @Return e
func (s *Service) DeleteBuildInObs(buildId string, projectId string) (e *errors.CodedError) {

	s.Logger.Info("BuildId %s start to delete build in OBS !", buildId)

	// get build info failed delete directly
	tmpBuild, err := dao.GetBuildById(buildId, projectId)
	if err != nil {
		s.Logger.Warn("BuildId %s Get Build info failed!", buildId)
	}

	if tmpBuild.StorageBucketName == "" && tmpBuild.StorageKey == "" {
		s.Logger.Warn("This Build %s is not in OBS!", buildId)
		return nil
	}

	// get ResDomianInfo by projectId
	resDomainInfo, resError := base.GetResDomainInfo(projectId, tmpBuild.StorageRegion)
	if resError != nil {
		return e
	}

	// get use resUser ProjectId to delete bucket
	obsClient, obserror := client.GetAgencyObsClient(tmpBuild.StorageRegion,
		resDomainInfo.AgencyName, resDomainInfo.ResDomainId)
	if obserror != nil {
		return s.ErrorMsg(errors.ServerInternalError, "Delete build Get obsclient error!",
			obserror)
	}

	exist, obsErr := client.HeadBucket(obsClient, tmpBuild.StorageBucketName)
	if exist {
		s.Logger.Info("Bucket already exist.BucketName = %s ", tmpBuild.StorageBucketName)
		if obsErr == nil {
			err = client.DeleteObsObject(obsClient, tmpBuild.StorageKey, tmpBuild.StorageBucketName)
			if err != nil {
				return s.ErrorMsg(errors.ServerInternalError, "Delete build from OBS failed!", err)
			}

			// try to delete bucket TODO: bucket need to delete when user is quit
			client.DeleteObsBucket(obsClient, tmpBuild.StorageBucketName)

		} else {
			return s.ErrorMsg(errors.ServerInternalError, "Delete buildBucket from OBS failed!", obsErr)
		}
	} else {
		// if Bucket Not Exist return success
		if obserror != nil {
			return s.ErrorMsg(errors.ServerInternalError,
				"Delete buildBucket when HeadBucket from OBS failed!", obserror)
		}
	}

	obsClient.Close()
	s.Logger.Info("BuildId %s  object and bucket delete from obs success!", buildId)
	return nil

}
