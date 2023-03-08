// Copyright (c) Huawei Technologies Co., Ltd. 2022-2022. All rights reserved.

// 应用包创建方法
package build

import (
	"fleetmanager/api/errors"
	"fleetmanager/api/model/build"
	"fleetmanager/api/params"
	"fleetmanager/api/service/base"
	"fleetmanager/client"
	"github.com/beego/beego/v2/server/web/context"
	"github.com/google/uuid"
	"io/ioutil"
	"mime/multipart"
	"strings"
)

func (s *Service) Upload(ctx *context.Context, file multipart.File, information *multipart.FileHeader,
	bucketName string, region string) (build.UploadResponse, *errors.CodedError) {
	b := build.UploadResponse{}
	projectId := ctx.Input.Param(params.ProjectId)

	if information.Size >= params.MaxBuildSize {
		s.Logger.Error("build file too large")
		return b, s.ErrorMsg(errors.InvalidParameterValue, "build file too large", nil)
	}

	data, err := ioutil.ReadAll(file)

	u, _ := uuid.NewUUID()
	tmpId := u.String()

	fileObj := strings.Split(information.Filename, ".")
	fileName := fileObj[0]
	fileExt := fileObj[1]
	if fileExt != "zip" && fileExt != "rar" {
		s.Logger.Error("File type error: %v", err)
		return b, s.ErrorMsg(errors.InvalidParameterValue, ".zip and .rar file are supported", nil)
	}

	// check build Number
	if e := s.CheckBuildNumber(projectId); e != nil {
		return b, e
	}

	b.BucketName = bucketName
	b.StorageRegion = region
	b.BucketKey = fileName + tmpId + "." + fileExt

	// check build bucket
	if e := s.CheckBucket(bucketName, projectId, region); e != nil {
		return b, e
	}

	if e := s.UploadFile(data, b, projectId, region); e != nil {
		return b, e
	}

	return b, nil
}

func (s *Service) UploadFile(file []byte, b build.UploadResponse, projectId string,
	region string) (e *errors.CodedError) {

	s.Logger.Info("Upload build to obs %s .", b.BucketKey)

	// get ResDomianInfo by projectId
	resDomainInfo, resError := base.GetResDomainInfo(projectId, region)
	if resError != nil {
		return resError
	}

	// get resDomain ProjectId
	obsClient, obserror := client.GetAgencyObsClient(region, resDomainInfo.AgencyName, resDomainInfo.ResDomainId)
	if obserror != nil {
		return s.ErrorMsg(errors.ServerInternalError,
			"Create build when Check bucket failed!Can not get client", obserror)
	}

	_, obserror = client.UploadFileToOBS(obsClient, file, b.BucketName, b.BucketKey)
	if obserror != nil {
		return s.ErrorMsg(errors.ServerInternalError,
			"Upload build to obs error", obserror)
	}

	// change ACL no write to delete
	err := client.RecycleObsBucketACL(obsClient, b.BucketName, resDomainInfo.ResDomainId)
	if err != nil {
		return nil
	}

	obsClient.Close()
	s.Logger.Info("Upload file success.")

	return nil
}
