// Copyright (c) Huawei Technologies Co., Ltd. 2022-2022. All rights reserved.

// 应用包查询方法
package build

import (
	"fleetmanager/api/common/query"
	"fleetmanager/api/errors"
	"fleetmanager/api/model/build"
	"fleetmanager/api/params"
	"fleetmanager/api/service/constants"
	"fleetmanager/api/service/base"
	"fleetmanager/api/user"
	"fleetmanager/client"
	"fleetmanager/db/dao"
	"fleetmanager/db/dbm"
	"strings"

	"github.com/beego/beego/v2/server/web/context"
)

func (s *Service) List(ctx *context.Context) (build.List, *errors.CodedError) {
	var list build.List
	projectId := ctx.Input.Param(params.ProjectId)

	q := build.QueryRequest{}

	q.Name = ctx.Input.Query("name")
	q.State = ctx.Input.Query("state")
	q.CreationTime = ctx.Input.Query("creation_time")

	offset, err := query.CheckOffset(ctx)
	if err != nil {
		return list, s.ErrorMsg(errors.InvalidParameterValue, "invalid offset", err)
	}
	limit, err := query.CheckLimit(ctx)
	if err != nil {
		return list, s.ErrorMsg(errors.InvalidParameterValue, "invalid limit", err)
	}
	if q.CreationTime != "" && !strings.Contains(q.CreationTime, "/") {
		return list, s.ErrorMsg(errors.InvalidParameterValue, "creation_time should be startTimeStamp/endTimeStamp", nil)
	}

	if q.State != "" && q.State != constants.BuildStateFailed && q.State != constants.BuildStateReady &&
		q.State != constants.BuildStateInitialized && q.State != constants.BuildImageInitialized {
		return list, s.ErrorMsg(errors.InvalidParameterValue, "invalid state", nil)
	}

	ds, total, err := dao.QueryBuildByCondition(projectId, q, offset, limit)
	if err != nil {
		return list, s.ErrorMsg(errors.DBError, "query db error", err)
	}
	list.TotalCount = total

	for _, d := range ds {
		b := buildBuildModel(&d)
		list.Builds = append(list.Builds, b)
	}
	list.Count = len(ds)

	return list, nil
}

// @Title Show
// @Description  Show build detailInfo
// @Author wangnannan 2022-05-07 09:43:36 ${time}
// @Param ctx
// @Return build.Build
// @Return error
func (s *Service) Show(ctx *context.Context) (BuildMsg, *errors.CodedError) {
	var b dao.Build
	var f []dao.Fleet
	var fm []build.FleetMsg
	var bd BuildMsg
	buildId := ctx.Input.Param(params.BuildId)
	projectId := ctx.Input.Param(params.ProjectId)
	err := dbm.Ormer.QueryTable(dao.BuildTable).Filter("Id", buildId).Filter("ProjectId", projectId).One(&b)
	if err != nil {
		return bd, s.ErrorMsg(errors.DBError, "get build from db error", err)
	}
	_, err = dbm.Ormer.QueryTable(dao.FleetTable).Filter("BuildId", buildId).
		Filter("ProjectId", projectId).Filter("Terminated", false).
		Filter("State__in", dao.FleetStateActive, dao.FleetStateCreating, dao.FleetStateDeleting, dao.FleetStateError).
		All(&f)
	if err != nil {
		return bd, s.ErrorMsg(errors.DBError, "get fleet created by build from db error", err)
	}

	fm = buildFleetModel(f)
	bm := buildFullBuild(&b)
	bd.Build = bm
	bd.FleetList.Fleet = fm
	bd.FleetList.Count = len(fm)

	return bd, nil
}

func (s *Service) Bucket(ctx *context.Context) (build.BucketList, *errors.CodedError) {
	var bucketList build.BucketList
	projectId := ctx.Input.Param(params.ProjectId)
	region := ctx.Input.Query("region")
	if region == "" {
		u, err := dao.GetUserResConfByProjectId(projectId)
		if err != nil {
			return bucketList, s.ErrorMsg(errors.InvalidParameterValue, "get region by projectId error", nil)
		}
		if u == nil || u.Region == "" {
			return bucketList, s.ErrorMsg(errors.InvalidParameterValue, "error region value", nil)
		}
		region = u.Region
	}
	// get ResDomianInfo by projectId
	resDomainInfo, resError := base.GetResDomainInfo(projectId, region)
	if resError != nil {
		return bucketList, resError
	}

	// get resDomain ProjectId
	obsClient, obserror := client.GetAgencyObsClient(region, resDomainInfo.AgencyName, resDomainInfo.ResDomainId)
	if obserror != nil {
		return bucketList, s.ErrorMsg(errors.ServerInternalError, "get obsClient error", obserror)
	}

	bucketList.BucketName, obserror = client.ListOBSBucket(obsClient)
	if obserror != nil {
		return bucketList, s.ErrorMsg(errors.ServerInternalError, "list bucket error", obserror)
	}
	bucketList.Count = len(bucketList.BucketName)

	return bucketList, nil
}

func (s *Service) Image(ctx *context.Context) (build.ImageList, *errors.CodedError) {
	var imageList build.ImageList
	projectId := ctx.Input.Param(params.ProjectId)
	region := ctx.Input.Query("region")
	if region == "" {
		u, err := dao.GetUserResConfByProjectId(projectId)
		if err != nil {
			return imageList, s.ErrorMsg(errors.InvalidParameterValue, "get region by projectId error", nil)
		}
		if u == nil || u.Region == "" {
			return imageList, s.ErrorMsg(errors.InvalidParameterValue, "error region value", nil)
		}
		region = u.Region
	}
	// get ResDomianInfo by projectId
	resDomainInfo, resError := base.GetResDomainInfo(projectId, region)
	if resError != nil {
		return imageList, resError
	}
	// get resDomain ProjectId
	imsClient, imserror := client.GetAgencyIMSClient(region, resDomainInfo.ResProjectId, resDomainInfo.AgencyName, resDomainInfo.ResDomainId)
	if imserror != nil {
		return imageList, s.ErrorMsg(errors.ServerInternalError, "get imsClient error", imserror)
	}

	imageList, imserror = client.ListIMSImage(imsClient)
	if imserror != nil {
		return imageList, s.ErrorMsg(errors.ServerInternalError, "list image error", imserror)
	}

	imageList, err := imageFilter(imageList, projectId, region)
	if err != nil {
		return imageList, s.ErrorMsg(errors.ServerInternalError, "check private image used in build error", err)
	}

	return imageList, nil
}

func (s *Service) Vpc(ctx *context.Context) (build.VpcList, *errors.CodedError) {
	var vpcList build.VpcList
	projectId := ctx.Input.Param(params.ProjectId)
	region := ctx.Input.Query("region")
	if region == "" {
		u, err := dao.GetUserResConfByProjectId(projectId)
		if err != nil {
			return vpcList, s.ErrorMsg(errors.InvalidParameterValue, "get region by projectId error", nil)
		}
		if u == nil || u.Region == "" {
			return vpcList, s.ErrorMsg(errors.InvalidParameterValue, "error region value", nil)
		}
		region = u.Region
	}
	// get ResDomianInfo by projectId
	resDomainInfo, resError := base.GetResDomainInfo(projectId, region)
	if resError != nil {
		return vpcList, resError
	}

	// get resDomain ProjectId
	vpcClient, vpcerror := client.GetAgencyVpcClient(region, resDomainInfo.ResProjectId, resDomainInfo.AgencyName, resDomainInfo.ResDomainId)
	if vpcerror != nil {
		return vpcList, s.ErrorMsg(errors.ServerInternalError, "get imsClient error", vpcerror)
	}

	vpcList, vpcerror = client.ListVpcList(vpcClient)
	if vpcerror != nil {
		return vpcList, s.ErrorMsg(errors.ServerInternalError, "list image error", vpcerror)
	}

	return vpcList, nil
}

func (s *Service) Subnet(ctx *context.Context) (build.SubnetList, *errors.CodedError) {
	var subnetList build.SubnetList
	projectId := ctx.Input.Param(params.ProjectId)
	vpcId := ctx.Input.Query("vpc_id")
	region := ctx.Input.Query("region")
	if vpcId == "" {
		return subnetList, s.ErrorMsg(errors.InvalidParameterValue, "error vpcId or region", nil)
	}
	if region == "" {
		u, err := dao.GetUserResConfByProjectId(projectId)
		if err != nil {
			return subnetList, s.ErrorMsg(errors.InvalidParameterValue, "get region by projectId error", nil)
		}
		if u == nil || u.Region == "" {
			return subnetList, s.ErrorMsg(errors.InvalidParameterValue, "error region value", nil)
		}
		region = u.Region
	}
	// get ResDomianInfo by projectId
	resDomainInfo, resError := base.GetResDomainInfo(projectId, region)
	if resError != nil {
		return subnetList, resError
	}

	// get resDomain ProjectId
	vpcClient, vpcerror := client.GetAgencyVpcClient(region, resDomainInfo.ResProjectId, resDomainInfo.AgencyName, resDomainInfo.ResDomainId)
	if vpcerror != nil {
		return subnetList, s.ErrorMsg(errors.ServerInternalError, "get imsClient error", vpcerror)
	}

	subnetList, vpcerror = client.ListSubnetList(vpcClient, vpcId)
	if vpcerror != nil {
		return subnetList, s.ErrorMsg(errors.ServerInternalError, "list image error", vpcerror)
	}

	return subnetList, nil
}

// @Title GetUploadCredentials
// @Description  Get Credentials for obs upload
// @Author wangnannan 2022-05-07 09:43:48 ${time}
// @Param ctx
// @Return build.UploadCredentials
// @Return error
func (s *Service) GetUploadCredentials(ctx *context.Context) (build.UploadCredentials, error) {
	regionId := ctx.Input.Query(params.QueryRegionId)
	bucketKey := ctx.Input.Query(params.QueryBucketKey)
	projectId := ctx.Input.Param(params.ProjectId)

	crid, err := s.GetObsUploadCredentialsByRegion(regionId, bucketKey, projectId)
	if err != nil {
		return crid, err
	}

	return crid, nil
}

// @Title GetObsUploadCredentialsByRegion
// @Description  Get resUsers info  and create bucket for this build
// @Author wangnannan 2022-05-07 09:44:00 ${time}
// @Param regionId
// @Param bucketKey
// @Param projectId
// @Return build.UploadCredentials
// @Return *errors.CodedError
func (s *Service) GetObsUploadCredentialsByRegion(regionId string, bucketKey string,
	projectId string) (build.UploadCredentials, *errors.CodedError) {

	var cred build.UploadCredentials

	// create bucket
	bucketName, err := s.CreateBucketInObs(bucketKey, regionId, projectId)
	if err != nil {
		return cred, err
	}
	cred.BucketName = bucketName
	cred.BucketKey = bucketKey
	cred.RegionId = regionId

	// Get User Domain Tmp token
	userSecurityInfo, err := s.GetSecurityInfoForOriginUser(regionId, projectId)
	if err != nil {
		return cred, err
	}

	cred.AccessKeyId = userSecurityInfo.AK
	cred.SecretAccessKey = userSecurityInfo.SK
	cred.SecurityToken = userSecurityInfo.Token
	return cred, nil
}

// @Title GetSecurityInfoForResUser
// @Description  get AK/SK/securityToken for ResUser
// @Author wangnannan 2022-05-07 09:44:13 ${time}
// @Param region
// @Param projectId
// @Return ak
// @Return sk
// @Return token
// @Return e
func (s *Service) GetSecurityInfoForResUser(region string, projectId string) (
	*user.UserSecurityInfo, *errors.CodedError) {

	// get ResUserInfo by UserProjectId
	resDomainInfo, resError := base.GetResDomainInfo(projectId, region)
	if resError != nil {
		return nil, s.ErrorMsg(errors.ServerInternalError, "Get resDomian info failed!", resError)
	}

	userSecurityInfo, err := client.GetObsSecurityInfo(region, resDomainInfo.AgencyName, resDomainInfo.ResDomainId)
	if err != nil {
		return nil, s.ErrorMsg(errors.ServerInternalError, "Get res ak/sk/token failed!", resError)
	}

	return userSecurityInfo, nil
}

// @Title GetSecurityInfoForOriginUser
// @Description  get AK/SK/securityToken
// @Author wangnannan 2022-05-07 09:44:26 ${time}
// @Param region
// @Param projectId
// @Return ak
// @Return sk
// @Return token
// @Return e
func (s *Service) GetSecurityInfoForOriginUser(region string, projectId string) (
	userSecurityInfo *user.UserSecurityInfo, e *errors.CodedError) {

	// get ResUserInfo by UserProjectId
	originUserDomainId, originUserAgencyName, resError := base.GetOriginDomainInfo(projectId, region)
	if resError != nil {
		return nil, s.ErrorMsg(errors.ServerInternalError, "Get OriginUser domainID and agency info failed!", resError)
	}

	userSecurityInfo, err := client.GetObsSecurityInfo(region, originUserAgencyName, originUserDomainId)
	if err != nil {
		return nil, s.ErrorMsg(errors.ServerInternalError, "Get OriginUser ak/sk/token failed!", err)
	}

	return userSecurityInfo, nil
}

// @Title CreateBucketInObs
// @Description  Create bucket in obs for a new build
// @Author wangnannan 2022-05-07 09:44:38 ${time}
// @Param bucketKey
// @Param region
// @Param projectId
// @Return bucketName
// @Return e
func (s *Service) CreateBucketInObs(bucketKey string, region string,
	projectId string) (bucketName string, e *errors.CodedError) {

	s.Logger.Info("BuildId %s start to create bucket in OBS !", bucketKey)

	// get ResDomianInfo by projectId
	resDomainInfo, resError := base.GetResDomainInfo(projectId, region)
	if resError != nil {
		return "", e
	}

	// use resourceProjectId instead of bucketName
	bucketName = region + resDomainInfo.ResProjectId
	// can't get token when exit build relate to the same bucketKey and bucketName
	e = s.CheckBucketKeyExist(bucketKey, bucketName)
	if e != nil {
		return "", e
	}

	// get AK/SK/token for ResUser
	userSecurityInfo, err := s.GetSecurityInfoForResUser(region, projectId)
	if err != nil {
		return "", err
	}

	// get resDomain ProjectId
	obsClient, obserror := client.GetObsClient(userSecurityInfo, region)
	if obserror != nil {
		return bucketName, s.ErrorMsg(errors.ServerInternalError, "Create bucket Get Obsclient failed!", obserror)
	}

	// if bucketName already exist  onley need to set ACL
	exist, obserror := client.HeadBucket(obsClient, bucketName)
	if exist {
		s.Logger.Info("Bucket already exist.BucketName = %s", bucketName)
	} else {
		// create a new bucket for this user
		if obserror != nil {
			return "", s.ErrorMsg(errors.ServerInternalError,
				"Create bucket when HeadBucket from OBS failed!", obserror)
		}

		obserror = client.CreateObsBucket(obsClient, bucketName)
		if err != nil {
			return "", s.ErrorMsg(errors.ServerInternalError,
				"Create bucket from OBS failed!", obserror)
		}
	}

	// update ACL
	obserror = client.UpdateObsBucketACL(obsClient, bucketName,
		resDomainInfo.ResDomainId, resDomainInfo.OriginUserDomain)
	if obserror != nil {
		return bucketName, s.ErrorMsg(errors.ServerInternalError, "Set bucket ACL Failed!", obserror)
	}

	obsClient.Close()
	s.Logger.Info("BuildId %s   create bucket in OBS success!", bucketKey)
	return bucketName, nil

}

// @Title ErrorMsg
// @Description
// @Author wangnannan 2022-05-07 09:44:52 ${time}
// @Param code
// @Param msg
// @Param err
// @Return *errors.CodedError
func (s *Service) ErrorMsg(code errors.ErrCode, msg string, err error) *errors.CodedError {
	errMsg := msg
	if err != nil {
		errMsg += err.Error()
		s.Logger.Error(msg, err)
	} else {
		s.Logger.Error(msg)
	}

	e := errors.NewErrorF(code, errMsg)

	return e
}

// @Title CheckBucketKeyExist
// @Description  bucketName and bucketKey already exist can get the token
// @Author wangnannan 2022-05-07 10:08:27 ${time}
// @Param bucketKey
// @Param bucketName
// @Return *errors.CodedError
func (s *Service) CheckBucketKeyExist(bucketKey string, bucketName string) *errors.CodedError {

	exist := dao.CheckBuildByBucket(bucketName, bucketKey)
	if exist == true {
		return s.ErrorMsg(errors.ServerInternalError,
			"Duplicate bucketKey ! ",
			nil)
	}

	return nil
}

func imageFilter(imageList build.ImageList, projectId string, region string) (build.ImageList, error) {
	var res build.ImageList
	var b []dao.Build
	_, err := dbm.Ormer.QueryTable(dao.BuildTable).Filter("ProjectId", projectId).Filter("ImageRegion", region).All(&b)
	if err != nil {
		return res, err
	}
	set := make(map[string]bool)
	for _, bd := range b {
		set[bd.ImageId] = true
	}
	for _, image := range imageList.ImageList {
		exists := set[image.Id]
		if exists {
			continue
		}
		res.ImageList = append(res.ImageList, image)
	}
	res.Count = len(res.ImageList)
	return res, nil
}

type BuildMsg struct {
	Build     build.FullBuild `json:"build"`
	FleetList build.FleetList `json:"fleet_list"`
}
