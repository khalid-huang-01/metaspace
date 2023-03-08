// Copyright (c) Huawei Technologies Co., Ltd. 2022-2022. All rights reserved.

// 应用包创建方法
package build

import (
	"fleetmanager/api/errors"
	"fleetmanager/api/model/build"
	"fleetmanager/api/params"
	"fleetmanager/api/service/base"
	"fleetmanager/api/service/constants"
	"fleetmanager/client"
	"fleetmanager/db/dao"
	"fleetmanager/db/dbm"
	"fleetmanager/setting"
	"fleetmanager/workflow"
	"fleetmanager/workflow/directer"
	"fleetmanager/worknode"
	"github.com/beego/beego/v2/server/web/context"
)

// @Title Create
// @Description  create build function. Neet to create bucket first.
// @Author wangnannan 2022-05-07 09:13:44 ${time}
// @Param ctx
// @Param r
// @Return *build.Build
// @Return *errors.CodedError
func (s *Service) Create(ctx *context.Context, r build.CreateRequest) (*build.Build, *errors.CodedError) {
	s.createReq = &r
	projectId := ctx.Input.Param(params.ProjectId)
	regionStr := r.Region

	vpcId := r.VpcId
	subnetId := r.SubnetId
	enterpriseProjectId := r.EnterpriseProject
	if enterpriseProjectId == "" {
		enterpriseProjectId = setting.EnterpriseProject
	}
	if r.OperatingSystem == "" {
		r.OperatingSystem = setting.ImageRef
	}

	// check build Number
	if e := s.CheckBuildNumber(projectId); e != nil {
		return nil, e
	}

	// check build name
	if e := s.CheckBuildByName(r.Name, r.Version, projectId); e != nil {
		return nil, e
	}

	// check build bucket
	e := s.CheckBucket(r.StorageLocation.BucketName, projectId, regionStr)
	if e != nil {
		return nil, e
	}

	fileSize, e := s.CheckSize(r.StorageLocation.BucketName, r.StorageLocation.BucketKey,
		projectId, regionStr)
	if e != nil {
		return nil, e
	}

	b, err := dao.InsertBuild(r, regionStr, projectId, fileSize)
	if err != nil {
		return nil, errors.NewErrorF(errors.ServerInternalError, "insert build error")
	}
	bd := buildBuildModel(b)

	s.Build = b

	if e := s.startCreateBuildImageWorkflow(regionStr, r.OperatingSystem, vpcId, subnetId, enterpriseProjectId); e != nil {
		if e := s.updateStateError(); e != nil {
			s.Logger.Error("update build state to error failed: %v", e)
		}
		return &bd, e
	}

	return &bd, nil
}

func (s *Service) CreateByImage(ctx *context.Context, r build.CreateByImageRequest) (*build.Build, *errors.CodedError) {
	s.createByImageReq = &r
	projectId := ctx.Input.Param(params.ProjectId)
	regionStr := r.Region

	// check build Number
	if e := s.CheckBuildNumber(projectId); e != nil {
		return nil, e
	}

	// check build name
	if e := s.CheckBuildByName(r.Name, r.Version, projectId); e != nil {
		return nil, e
	}

	osVersion, e := s.GetOperatingSystem(r.ImageId, projectId)
	if e != nil {
		return nil, e
	}
	if osVersion == "" {
		s.Logger.Info("can not get the operating system by the private imageId %s", r.ImageId)
	}

	b, err := dao.InsertBuildByImage(r, regionStr, projectId, osVersion)
	bd := buildBuildModel(b)
	if err != nil {
		return &bd, errors.NewErrorF(errors.ServerInternalError, "insert build error")
	}

	s.Build = b

	return &bd, nil
}

// @Title CheckBucket
// @Description  check if name exit
// @Author wangnannan 2022-05-07 09:14:20 ${time}
// @Param bucketName
// @Param projectId
// @Param region
// @Return e
func (s *Service) CheckBucket(bucketName string, projectId string, region string) (e *errors.CodedError) {

	s.Logger.Info("Check bucket= %s .", bucketName)

	// get ResDomianInfo by projectId
	resDomainInfo, resError := base.GetResDomainInfo(projectId, region)
	if resError != nil {
		return resError
	}

	// get resDomain ProjectId
	obsClient, obsErr := client.GetAgencyObsClient(region, resDomainInfo.AgencyName, resDomainInfo.ResDomainId)
	if obsErr != nil {
		return s.ErrorMsg(errors.ServerInternalError,
			"Create build when Check bucket failed!Can not get client", obsErr)
	}

	exist, obsErr := client.HeadBucket(obsClient, bucketName)
	if exist {
		s.Logger.Info("Bucket already exist.BucketName =%s ", bucketName)
	} else {
		if obsErr != nil {
			return s.ErrorMsg(errors.ServerInternalError, "Create build when HeadBucket from OBS failed!", obsErr)
		} else {
			// if Bucket Not Exist return Error
			return s.ErrorMsg(errors.BucketNotExist, "Create build Bucket Not Exist!", nil)
		}
	}

	// change ACL no write to delete
	err := client.RecycleObsBucketACL(obsClient, bucketName, resDomainInfo.ResDomainId)
	if err != nil {
		s.Logger.Info("Set Bucket ACL failed")
	}

	obsClient.Close()
	s.Logger.Info("Check bucketName= %s exist success.", bucketName)

	return nil
}

func (s *Service) CheckSize(bucketName string, bucketKey string,
	projectId string, region string) (size int64, e *errors.CodedError) {
	s.Logger.Info("Check bucket= %s .", bucketName)

	// get ResDomianInfo by projectId
	resDomainInfo, resError := base.GetResDomainInfo(projectId, region)
	if resError != nil {
		return 0, resError
	}

	// get resDomain ProjectId
	obsClient, obsErr := client.GetAgencyObsClient(region, resDomainInfo.AgencyName, resDomainInfo.ResDomainId)
	if obsErr != nil {
		return 0, s.ErrorMsg(errors.ServerInternalError,
			"Create build when Check bucket failed!Can not get client", obsErr)
	}

	size, obsErr = client.CheckObjectSizeInOBS(obsClient, bucketName, bucketKey)
	if obsErr != nil {
		return 0, s.ErrorMsg(errors.ServerInternalError, "Check build size failed", obsErr)
	}
	return size, nil
}

// @Title CheckBuildByName
// @Description  check if name exit
// @Author wangnannan 2022-05-07 09:14:40 ${time}
// @Param name
// @Param bucketName
// @Param bucketKey
// @Return e
func (s *Service) CheckBuildByName(name string, version string, projectId string) (e *errors.CodedError) {
	exist := dao.CheckBuildByName(name, version, projectId)
	if exist == true {
		s.Logger.Error("Build already exit! BuildName is %s, Version is %s", name, version)
		e := errors.NewErrorF(errors.BuildIsAlreadyExist, "Build already exit! BuildName is %s, Version is %s", name, version)
		return e
	}
	return nil
}

func (s *Service) CheckBuildByBucket(bucketName string, bucketKey string) (e *errors.CodedError) {
	exist := dao.CheckBuildByBucket(bucketName, bucketKey)
	if exist == true {
		s.Logger.Error("BucketKey is already exist!")
		e := errors.NewErrorF(errors.DuplicateBucketKey,
			"DuplicateBuildLocation! BucketName:"+bucketName+" BucketKey:"+bucketKey)
		return e
	}

	return nil
}

// @Title CheckBuildNumber
// @Description  check build Number,if exist buildnumber>100 reject
// @Author wangnannan 2022-05-07 09:15:37 ${time}
// @Param projectid
// @Return e
func (s *Service) CheckBuildNumber(projectId string) (e *errors.CodedError) {

	number, err := dao.GetBuildCount(projectId)
	if err != nil {
		s.Logger.Error("Get build exist number failed!", err)
		e = errors.NewErrorF(errors.ServerInternalError, "Get build exist number failed!")
		return e
	}

	if number >= params.MaxBuildNumber {
		s.Logger.Error("Exist build number exceeds the maximum value!", err)
		e = errors.NewErrorF(errors.BuildNumExceedMaxSize, "Exist build number exceeds the maximum value!")
		return e
	}

	return nil
}

// @Title CheckOperatingSystem
// @Description  TODO:need to check support system  check OperateingSystem if legal
// @Author wangnannan 2022-05-07 09:15:59 ${time}
// @Return e
func (s *Service) CheckOperatingSystem() (e *errors.CodedError) {

	request := *s.createReq
	op := request.OperatingSystem

	s.Logger.Error("Operating system %s  is not supported!", op, e)
	e = errors.NewErrorF(errors.OperateSystemNoSupport, "Operating system %s  is not supported!", op)
	return e
}

func (s *Service) GetOperatingSystem(imageId string, projectId string) (string, *errors.CodedError) {
	u, err := dao.GetUserResConfByProjectId(projectId)
	if err != nil {
		return "", s.ErrorMsg(errors.InvalidParameterValue, "get region by projectId error", nil)
	}
	resDomainInfo, resError := base.GetResDomainInfo(projectId, u.Region)
	if resError != nil {
		return "", resError
	}
	imsClient, imserror := client.GetAgencyIMSClient(u.Region, resDomainInfo.ResProjectId, resDomainInfo.AgencyName, resDomainInfo.ResDomainId)
	if imserror != nil {
		return "", s.ErrorMsg(errors.ServerInternalError, "get imsClient error", imserror)
	}

	operatingSystem, imserror := client.GetIMSOperatingSystem(imsClient, imageId)
	if imserror != nil {
		return "", s.ErrorMsg(errors.ServerInternalError, "list image error", imserror)
	}
	return operatingSystem, nil
}

// 更新应用包创建状态至失败
func (s *Service) updateStateError() *errors.CodedError {
	b := &dao.Build{
		Id:    s.Build.Id,
		State: constants.BuildStateFailed,
	}

	_, err := dbm.Ormer.Update(b, "State")

	if err != nil {
		return s.ErrorMsg(errors.BuildUpdateFailed,
			"update build state to error failed", nil)
	}
	return nil
}

func (s *Service) startCreateBuildImageWorkflow(regionStr string, imageRef string,
	vpcId string, subnetId string, enterpriseProjectId string) (e *errors.CodedError) {

	parameter := map[string]interface{}{
		directer.WfKeyRegion:              regionStr,
		directer.WfKeyBuild:               s.Build,
		directer.WfKeyBuildVpcId:          vpcId,
		directer.WfKeyBuildSubnetId:       subnetId,
		directer.WfKeyOriginProjectId:     s.Build.ProjectId,
		directer.WfKeyImageRef:            imageRef,
		directer.WfKeyBandwidthName:       s.Build.Id,
		directer.WfKeyBuildBandwidth:      setting.DefaultBandWidth,
		directer.WfKeyBandwidthType:       setting.Config.Get(setting.EipType + "." + regionStr).ToString(""),
		directer.WfKeyEnterpriseProjectId: enterpriseProjectId,
	}

	wf, err := workflow.CreateWorkflow("./conf/workflow/create_build_image_workflow.json",
		parameter,
		s.Build.Id,
		s.Build.ProjectId,
		s.Logger,
		worknode.WorkNodeId)

	if err != nil {
		s.Logger.Error("create workflow in create build error: %v", err)
		return errors.NewError(errors.ServerInternalError)
	}

	wf.Run()
	return nil
}
