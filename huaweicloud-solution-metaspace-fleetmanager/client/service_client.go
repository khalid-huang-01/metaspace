// Copyright (c) Huawei Technologies Co., Ltd. 2022-2022. All rights reserved.

// 客户端服务定义
package client

import (
	"bytes"
	"fleetmanager/api/errors"
	"fleetmanager/api/model/build"
	"fleetmanager/api/user"
	"fleetmanager/logger"
	"fleetmanager/setting"
	"fmt"
	"github.com/huaweicloud/huaweicloud-sdk-go-obs/obs"
	"github.com/huaweicloud/huaweicloud-sdk-go-v3/core/auth/basic"
	"github.com/huaweicloud/huaweicloud-sdk-go-v3/core/auth/global"
	"github.com/huaweicloud/huaweicloud-sdk-go-v3/core/config"
	"github.com/huaweicloud/huaweicloud-sdk-go-v3/core/httphandler"
	"github.com/huaweicloud/huaweicloud-sdk-go-v3/core/sdkerr"
	ecs "github.com/huaweicloud/huaweicloud-sdk-go-v3/services/ecs/v2"
	ecsmodel "github.com/huaweicloud/huaweicloud-sdk-go-v3/services/ecs/v2/model"
	eip "github.com/huaweicloud/huaweicloud-sdk-go-v3/services/eip/v2"
	eipmodel "github.com/huaweicloud/huaweicloud-sdk-go-v3/services/eip/v2/model"
	iam "github.com/huaweicloud/huaweicloud-sdk-go-v3/services/iam/v3"
	iammodel "github.com/huaweicloud/huaweicloud-sdk-go-v3/services/iam/v3/model"
	ims "github.com/huaweicloud/huaweicloud-sdk-go-v3/services/ims/v2"
	imsmodel "github.com/huaweicloud/huaweicloud-sdk-go-v3/services/ims/v2/model"
	vpc "github.com/huaweicloud/huaweicloud-sdk-go-v3/services/vpc/v2"
	vpcmodel "github.com/huaweicloud/huaweicloud-sdk-go-v3/services/vpc/v2/model"
	"net"
	"net/http"
	"strings"
	"time"
)

const (
	HttpCodeNotFound         = 404
	DefaultCallIntervalTimes = 5
	HttpCodeConfilct         = 409
)

// RequestHandler 请求handler
func RequestHandler(request http.Request) {
	logger.R.Info("Method:%+v, URL:%+v", request.Method, request.URL)
}

// ResponseHandler 响应handler
func ResponseHandler(response http.Response) {
	logger.R.Info("StatusCode:%+v, Status:%+v, ReqId:%s", response.StatusCode, response.Status,
		response.Header.Get("X-Request-Id"))
}

func getRegionEndpoint(regionId string, serviceType string) string {
	k := serviceType + "." + regionId
	endpoint := setting.Config.Get(k).ToString("")
	return endpoint
}

// @Title getObsEndpoint
// @Description
// @Author wangnannan 2022-05-07 11:22:47 ${time}
// @Param regionId
// @Return string
func getObsEndpoint(regionId string) string {
	endpoint := setting.Config.Get(setting.OBSEndpoint + "." + regionId).ToString("")
	return endpoint
}

func GetDefaultCredentials() (string, string, string) {
	return string(setting.ServiceAK), string(setting.ServiceSK), setting.ServiceDomainId
}

// GetAgencyVpcClient 获取Agency vpc client
func GetAgencyVpcClient(regionId string, projectId string, agencyName string,
	resDomainId string) (*vpc.VpcClient, error) {
	ak, sk, domainId := GetDefaultCredentials()
	iamClient := GetIamClient(ak, sk, regionId, domainId)
	tmpAk, tmpSk, tmpToken, err := GetSecurityToken(iamClient, &resDomainId, &agencyName)
	if err != nil {
		return nil, err
	}
	return GetTmpVpcClient(tmpAk, tmpSk, []byte(tmpToken), regionId, projectId), nil
}

func GetAgencyEcsClient(regionId string, projectId string, agencyName string,
	resDomainId string) (*ecs.EcsClient, error) {
	ak, sk, domainId := GetDefaultCredentials()
	iamClient := GetIamClient(ak, sk, regionId, domainId)
	tmpAk, tmpSk, tmpToken, err := GetSecurityToken(iamClient, &resDomainId, &agencyName)
	if err != nil {
		return nil, err
	}
	return GetTmpECSClient(tmpAk, tmpSk, []byte(tmpToken), regionId, projectId), nil
}

// GetAgencyEipClient 获取Agency eip client
func GetAgencyEipClient(regionId string, projectId string, agencyName string,
	resDomainId string) (*eip.EipClient, error) {
	ak, sk, domainId := GetDefaultCredentials()
	iamClient := GetIamClient(ak, sk, regionId, domainId)
	tmpAk, tmpSk, tmpToken, err := GetSecurityToken(iamClient, &resDomainId, &agencyName)
	if err != nil {
		return nil, err
	}
	return GetTmpEipClient(tmpAk, tmpSk, []byte(tmpToken), regionId, projectId), nil
}

// GetAgencyIMSClient 获取Agency IMS client
func GetAgencyIMSClient(regionId string, projectId string, agencyName string,
	resDomainId string) (*ims.ImsClient, error) {
	ak, sk, domainId := GetDefaultCredentials()
	iamClient := GetIamClient(ak, sk, regionId, domainId)
	tmpAk, tmpSk, tmpToken, err := GetSecurityToken(iamClient, &resDomainId, &agencyName)
	if err != nil {
		return nil, err
	}
	return GetTmpIMSClient(tmpAk, tmpSk, []byte(tmpToken), regionId, projectId), nil
}

// GetImageId 通过临时ECS设定的操作系统 查询公共镜像id
func GetImageId(imsClient *ims.ImsClient, imageRef string) (string, error) {
	imsRequest := &imsmodel.GlanceListImagesRequest{}
	nameRequest := imageRef
	imsRequest.Name = &nameRequest
	imsResponse, err := imsClient.GlanceListImages(imsRequest)
	if err != nil {
		return "", err
	}
	if len(*imsResponse.Images) == 0 {
		return "", fmt.Errorf("this imageRef is not supported %s", nameRequest)
	}
	imageId := (*imsResponse.Images)[0].Id
	return imageId, nil
}

// GetSecurityToken 获取安全token
func GetSecurityToken(iamClient *iam.IamClient, resDomainId *string,
	agencyName *string) (string, string, string, error) {
	req := &iammodel.CreateTemporaryAccessKeyByAgencyRequest{
		Body: &iammodel.CreateTemporaryAccessKeyByAgencyRequestBody{
			Auth: &iammodel.AgencyAuth{
				Identity: &iammodel.AgencyAuthIdentity{
					Methods: []iammodel.AgencyAuthIdentityMethods{
						iammodel.GetAgencyAuthIdentityMethodsEnum().ASSUME_ROLE,
					},
					AssumeRole: &iammodel.IdentityAssumerole{
						AgencyName: *agencyName,
						DomainId:   resDomainId,
					},
				},
			},
		},
	}

	rsp, err := iamClient.CreateTemporaryAccessKeyByAgency(req)
	if err != nil {
		return "", "", "", err
	}
	return rsp.Credential.Access, rsp.Credential.Secret, rsp.Credential.Securitytoken, nil
}

// GetIamClient 获取Iam client
func GetIamClient(ak string, sk string, regionId string, domainId string) *iam.IamClient {
	client := iam.NewIamClient(
		iam.IamClientBuilder().WithEndpoint(getRegionEndpoint(regionId, setting.IAMEndpoint)).WithCredential(
			global.NewCredentialsBuilder().
				WithAk(ak).
				WithSk(sk).
				WithDomainId(domainId).
				Build()).
			WithHttpConfig(config.DefaultHttpConfig().
				WithIgnoreSSLVerification(true).
				WithHttpHandler(httphandler.
					NewHttpHandler().
					AddRequestHandler(RequestHandler).
					AddResponseHandler(ResponseHandler))).
			Build())
	return client
}

// GetTmpVpcClient 获取临时vpc client
func GetTmpVpcClient(ak string, sk string, tmpToken []byte, regionId string, projectId string) *vpc.VpcClient {
	client := vpc.NewVpcClient(
		vpc.VpcClientBuilder().
			WithEndpoint(getRegionEndpoint(regionId, setting.VPCEndpoint)).
			WithCredential(
				basic.NewCredentialsBuilder().
					WithAk(ak).
					WithSk(sk).
					WithSecurityToken(string(tmpToken)).
					WithProjectId(projectId).
					Build()).
			WithHttpConfig(config.DefaultHttpConfig().
				WithIgnoreSSLVerification(true).
				WithHttpHandler(httphandler.
					NewHttpHandler().
					AddRequestHandler(RequestHandler).
					AddResponseHandler(ResponseHandler))).
			Build())
	return client
}

// GetTmpEipClient 获取临时的EipClient
func GetTmpEipClient(ak string, sk string, tmpToken []byte, regionId string, projectId string) *eip.EipClient {
	client := eip.NewEipClient(
		eip.EipClientBuilder().
			WithEndpoint(getRegionEndpoint(regionId, setting.VPCEndpoint)).
			WithCredential(
				basic.NewCredentialsBuilder().
					WithAk(ak).
					WithSk(sk).
					WithSecurityToken(string(tmpToken)).
					WithProjectId(projectId).
					Build()).
			WithHttpConfig(config.DefaultHttpConfig().
				WithIgnoreSSLVerification(true).
				WithHttpHandler(httphandler.
					NewHttpHandler().
					AddRequestHandler(RequestHandler).
					AddResponseHandler(ResponseHandler))).
			Build())

	return client
}

// GetTmpIMSClient 获取临时的IMS Client
func GetTmpIMSClient(ak string, sk string, tmpToken []byte, regionId string, projectId string) *ims.ImsClient {
	client := ims.NewImsClient(
		ims.ImsClientBuilder().
			WithEndpoint(getRegionEndpoint(regionId, setting.IMSEndpoint)).
			WithCredential(
				basic.NewCredentialsBuilder().
					WithAk(ak).
					WithSk(sk).WithSecurityToken(string(tmpToken)).
					WithProjectId(projectId).
					Build()).
			WithHttpConfig(config.DefaultHttpConfig().
				WithIgnoreSSLVerification(true).
				WithHttpHandler(httphandler.
					NewHttpHandler().
					AddRequestHandler(RequestHandler).
					AddResponseHandler(ResponseHandler))).
			Build())
	return client
}

// GetTmpECSClient 获取临时的ECS Client
func GetTmpECSClient(ak string, sk string, tmpToken []byte, regionId string, projectId string) *ecs.EcsClient {
	client := ecs.NewEcsClient(
		ecs.EcsClientBuilder().
			WithEndpoint(getRegionEndpoint(regionId, setting.ECSEndpoint)).
			WithCredential(
				basic.NewCredentialsBuilder().
					WithAk(ak).
					WithSk(sk).WithSecurityToken(string(tmpToken)).
					WithProjectId(projectId).
					Build()).
			WithHttpConfig(config.DefaultHttpConfig().
				WithIgnoreSSLVerification(true).
				WithHttpHandler(httphandler.
					NewHttpHandler().
					AddRequestHandler(RequestHandler).
					AddResponseHandler(ResponseHandler))).
			Build())
	return client
}

// @Title GetAgencyObsClient
// @Description
// @Author wangnannan 2022-05-07 10:22:23 ${time}
// @Param regionId
// @Param agencyName
// @Param resDomainId
// @Return *obs.ObsClient
// @Return error
func GetAgencyObsClient(regionId string, agencyName string, resDomainId string) (*obs.ObsClient, error) {
	userSecurityInfo, err := GetObsSecurityInfo(regionId, agencyName, resDomainId)
	if err != nil {
		return nil, err
	}
	return GetObsClient(userSecurityInfo, regionId)
}

// GetOriginObsClient 获取管理账号obsClient以下载服务组件
func GetOriginObsClient(regionId string) (*obs.ObsClient, error) {
	ak, sk, _ := GetDefaultCredentials()
	obsClient, err := obs.New(ak, sk, getRegionEndpoint(regionId, setting.OBSEndpoint))
	if err != nil {
		return nil, err
	}
	return obsClient, nil
}

// @Title GetObsSecurityInfo
// @Description
// @Author wangnannan 2022-05-07 10:22:31 ${time}
// @Param regionId
// @Param agencyName
// @Param resDomainId
// @Return *user.UserSecurityInfo
// @Return error
func GetObsSecurityInfo(regionId string, agencyName string, resDomainId string) (*user.UserSecurityInfo, error) {
	var userSecurityInfo user.UserSecurityInfo
	ak, sk, domainId := GetDefaultCredentials()
	iamClient := GetIamClient(ak, sk, regionId, domainId)
	tmpAk, tmpSk, token, err := GetSecurityToken(iamClient, &resDomainId, &agencyName)
	if err != nil {
		return &userSecurityInfo, err
	}

	userSecurityInfo.AK = tmpAk
	userSecurityInfo.SK = tmpSk
	userSecurityInfo.Token = token

	return &userSecurityInfo, nil
}

// @Title GetObsClient
// @Description
// @Author wangnannan 2022-05-07 10:22:37 ${time}
// @Param userSecurityInfo
// @Param regionId
// @Return *obs.ObsClient
// @Return error
func GetObsClient(userSecurityInfo *user.UserSecurityInfo, regionId string) (*obs.ObsClient, error) {
	obsEndpoint := getObsEndpoint(regionId)
	var obsClient, err = obs.New(userSecurityInfo.AK, userSecurityInfo.SK, obsEndpoint,
		obs.WithSecurityToken(userSecurityInfo.Token))
	if err != nil {
		return nil, err
	}

	return obsClient, nil
}

// @Title DeleteObsObject
// @Description
// @Author wangnannan 2022-05-07 10:22:43 ${time}
// @Param obsClient
// @Param storageKey
// @Param storageBucketName
// @Return error
func DeleteObsObject(obsClient *obs.ObsClient, storageKey string, storageBucketName string) error {
	input := &obs.DeleteObjectInput{}
	input.Bucket = storageBucketName
	input.Key = storageKey
	_, err := obsClient.DeleteObject(input)
	return err

}

// @Title UpdateObsBucketACL
// @Description  update ACL from resUser to originUser bucket is belongs to resUser
// @Author wangnannan 2022-05-07 10:22:49 ${time}
// @Param obsClient
// @Param bucketName
// @Param resUserDomainID
// @Param originUserId
// @Return error
func UpdateObsBucketACL(obsClient *obs.ObsClient, bucketName string,
	resUserDomainID string, originUserId string) error {
	input := &obs.SetBucketAclInput{}
	input.Bucket = bucketName
	input.Owner.ID = resUserDomainID
	var grants [2]obs.Grant

	grants[0].Grantee.Type = obs.GranteeUser
	grants[0].Grantee.ID = originUserId
	grants[0].Permission = obs.PermissionWrite

	// need set ACL for myself
	grants[1].Grantee.Type = obs.GranteeUser
	grants[1].Grantee.ID = resUserDomainID
	grants[1].Permission = obs.PermissionFullControl

	input.Grants = grants[0:OBS_GRANTNUM]
	_, err := obsClient.SetBucketAcl(input)
	if err != nil {
		return err
	}

	return nil
}

// @Title RecycleObsBucketACL
// @Description  update ACL from resUser to originUser bucket is belongs to resUser
// @Author wangnannan 2022-05-07 10:23:04 ${time}
// @Param obsClient
// @Param bucketName
// @Param resUserDomainID
// @Return error
func RecycleObsBucketACL(obsClient *obs.ObsClient, bucketName string, resUserDomainID string) error {
	input := &obs.SetBucketAclInput{}
	input.Bucket = bucketName
	input.Owner.ID = resUserDomainID
	var grants [1]obs.Grant

	// need set ACL for myself
	grants[0].Grantee.Type = obs.GranteeUser
	grants[0].Grantee.ID = resUserDomainID
	grants[0].Permission = obs.PermissionFullControl

	input.Grants = grants[0:1]
	_, err := obsClient.SetBucketAcl(input)
	if err != nil {
		return err
	}

	return nil
}

// @Title CreateObsBucket
// @Description  Create a new obs bucket for a new build
// @Author wangnannan 2022-05-07 10:23:18 ${time}
// @Param obsClient
// @Param bucketName
// @Return error
func CreateObsBucket(obsClient *obs.ObsClient, bucketName string) error {
	input := &obs.CreateBucketInput{}

	// use buildId as  buildId
	input.Bucket = bucketName
	output, err := obsClient.CreateBucket(input)
	if err != nil {
		return err
	}

	fmt.Println(output)

	return err
}

// @Title HeadBucket
// @Description  Create a new obs bucket for a new build
// @Author wangnannan 2022-05-07 10:23:28 ${time}
// @Param obsClient
// @Param bucketName
// @Return bool
// @Return error
func HeadBucket(obsClient *obs.ObsClient, bucketName string) (bool, error) {
	_, obserror := obsClient.HeadBucket(bucketName)
	if obserror == nil {
		// Bucket exists
		return true, nil
	} else {
		if obsError, ok := obserror.(obs.ObsError); ok {
			if obsError.StatusCode == OBS_HEADBUCKET_404 {
				// Bucket does not exists
				return false, nil
			} else {
				return false, obserror
			}
		} else {
			return false, obserror
		}
	}

}

func CheckObjectSizeInOBS(obsClient *obs.ObsClient, bucketName string, bucketKey string) (int64, error) {
	input := obs.GetObjectMetadataInput{}
	input.Bucket = bucketName
	input.Key = bucketKey
	res, obserror := obsClient.GetObjectMetadata(&input)
	if obserror != nil {
		return 0, obserror
	}
	return res.ContentLength, obserror
}

// @Title DeleteObsBucket
// @Description
// @Author wangnannan 2022-05-07 10:24:11 ${time}
// @Param obsClient
// @Param storageBucketName
// @Return error
func DeleteObsBucket(obsClient *obs.ObsClient, storageBucketName string) error {
	_, err := obsClient.DeleteBucket(storageBucketName)
	obsClient.Close()
	return err

}

// CreateVpc 创建vpc
func CreateVpc(vpcClient VpcV2, cidr *string, name *string, enterpriseProject *string) (string, error) {
	req := &vpcmodel.CreateVpcRequest{
		Body: &vpcmodel.CreateVpcRequestBody{
			Vpc: &vpcmodel.CreateVpcOption{
				Cidr:                cidr,
				Name:                name,
				EnterpriseProjectId: enterpriseProject,
			},
		},
	}

	rsp, err := vpcClient.CreateVpc(req)
	if err != nil {
		return "", err
	}

	return rsp.Vpc.Id, nil
}

// GetSubnet 获取子网
func GetSubnet(vpcClient *vpc.VpcClient, name *string, vpcId string) (string, string, error) {
	req := &vpcmodel.ListSubnetsRequest{
		VpcId: &vpcId,
	}
	rsp, err := vpcClient.ListSubnets(req)
	if err != nil {
		return "", "", err
	}

	for _, s := range *rsp.Subnets {
		if s.Name == *name {
			return s.Id, s.VpcId, nil
		}
	}

	return "", "", nil
}

// GetSubnetById 根据子网ID获取子网信息
func GetSubnetById(vpcClient *vpc.VpcClient, subnetId string) (*vpcmodel.Subnet, error) {
	reqGetSubnet := &vpcmodel.ShowSubnetRequest{
		SubnetId: subnetId,
	}
	subnet, err := vpcClient.ShowSubnet(reqGetSubnet)
	if err != nil {
		return nil, err
	}
	return subnet.Subnet, nil
}

// 检查随机的子网是否与VPC内已有子网冲突
func CheckSubnet(vpcClient *vpc.VpcClient, cidr *string, vpcId string) (bool, error) {
	reqListSubnet := &vpcmodel.ListSubnetsRequest{
		VpcId: &vpcId,
	}
	reqShowVpc := &vpcmodel.ShowVpcRequest{
		VpcId: vpcId,
	}
	rsp, err1 := vpcClient.ListSubnets(reqListSubnet)
	vpc, err2 := vpcClient.ShowVpc(reqShowVpc)
	if err1 != nil || err2 != nil {
		return false, fmt.Errorf("list subnets or show vpc error: err1: %+v or err2: %+v", err1, err2)
	}
	_, IPNetRandom, err1 := net.ParseCIDR(*cidr)
	_, vpcCidr, err2 := net.ParseCIDR(vpc.Vpc.Cidr)
	if err1 != nil || err2 != nil {
		return false, fmt.Errorf("parse cidr: %s or %s error, err1: %+v or err2: %+v", vpc.Vpc.Cidr, *cidr, err1, err2)
	}
	// 检查子网是否在vpc网段范围内
	if !ContainsCidr(vpcCidr, IPNetRandom) {
		return false, nil
	}
	for _, s := range *rsp.Subnets {
		_, IPNetA, err := net.ParseCIDR(s.Cidr)
		if err != nil {
			return false, err
		}
		// 检查随机的子网是否与VPC内已有子网冲突
		if ContainsCidr(IPNetRandom, IPNetA) || ContainsCidr(IPNetA, IPNetRandom) {
			return true, nil
		}
	}
	return false, nil
}

// 判断子网a是否包含子网b
func ContainsCidr(a, b *net.IPNet) bool {
	onesA, _ := a.Mask.Size()
	onesB, _ := b.Mask.Size()
	return onesA <= onesB && a.Contains(b.IP)
}

// GetVpc 获取vpc
func GetVpc(vpcClient *vpc.VpcClient, name *string) (*vpcmodel.Vpc, error) {
	req := &vpcmodel.ListVpcsRequest{}
	rsp, err := vpcClient.ListVpcs(req)
	if err != nil {
		return nil, err
	}

	for _, v := range *rsp.Vpcs {
		if v.Name == *name {
			return &v, nil
		}
	}

	return nil, nil
}

// GetVpc 获取vpc
func GetVpcById(vpcClient *vpc.VpcClient, id *string) (*vpcmodel.Vpc, error) {
	req := &vpcmodel.ShowVpcRequest{
		VpcId: *id,
	}
	rsp, err := vpcClient.ShowVpc(req)
	if err != nil {
		return nil, err
	}

	return rsp.Vpc, nil
}

func CheckVpc(vpc *vpcmodel.Vpc, vpcCidr *string) (bool, error) {
	_, IPNetVpcA, err := net.ParseCIDR(vpc.Cidr)
	if err != nil {
		return false, err
	}
	_, IPNetVpcB, err := net.ParseCIDR(*vpcCidr)
	if err != nil {
		return false, err
	}
	return ContainsCidr(IPNetVpcA, IPNetVpcB), nil
}

// DeleteVpc 删除vpc
func DeleteVpc(vpcClient *vpc.VpcClient, vpcId *string) error {
	req := &vpcmodel.DeleteVpcRequest{
		VpcId: *vpcId,
	}
	_, err := vpcClient.DeleteVpc(req)
	if err != nil {
		// 如果已经删除了, 就不需要继续删了
		newErr, ok := err.(*sdkerr.ServiceResponseError)
		if ok && (newErr.StatusCode == HttpCodeNotFound) {
			err = nil
		}
	}
	return err
}

// DeleteSubnet 删除子网
func DeleteSubnet(vpcClient *vpc.VpcClient, subnetId *string, vpcId *string) error {
	req := &vpcmodel.DeleteSubnetRequest{
		VpcId:    *vpcId,
		SubnetId: *subnetId,
	}
	_, err := vpcClient.DeleteSubnet(req)
	if err != nil {
		// 如果已经删除了, 就不需要继续删了
		newErr, ok := err.(*sdkerr.ServiceResponseError)
		if ok && (newErr.StatusCode == HttpCodeNotFound) {
			err = nil
		}
	}
	return err
}

// WaitVpcDeleted 等待vpc删除完成
func WaitVpcDeleted(vpcClient *vpc.VpcClient, vpcId *string) error {
	req := &vpcmodel.ListVpcsRequest{}
	for {
		// TODO(wangjun): timeout
		rsp, err := vpcClient.ListVpcs(req)
		if err != nil {
			return err
		}

		exists := false
		for _, v := range *rsp.Vpcs {
			if v.Id == *vpcId {
				exists = true
				break
			}
		}

		if !exists {
			return nil
		}

		time.Sleep(DefaultCallIntervalTimes * time.Second)
	}
}

// WaitVpcReady 等待vpc创建完成
func WaitVpcReady(vpcClient *vpc.VpcClient, vpcId *string) error {
	req := &vpcmodel.ShowVpcRequest{
		VpcId: *vpcId,
	}

	for {
		// TODO(wangjun): timeout
		rsp, err := vpcClient.ShowVpc(req)
		if err != nil {
			return err
		}

		if rsp.Vpc.Status == vpcmodel.GetVpcStatusEnum().CREATING {
			time.Sleep(DefaultCallIntervalTimes * time.Second)
			continue
		}

		if rsp.Vpc.Status == vpcmodel.GetVpcStatusEnum().ERROR {
			return errors.NewError("Vpc status error")
		}

		// 创建已成功
		return nil
	}
}

// CreateSubnet 创建子网
func CreateSubnet(vpcClient *vpc.VpcClient, vpcId *string, gatewayIp *string, cidr *string, name *string,
	primaryDns *string, secondaryDns *string, dnsList *[]string) (string, error) {
	req := &vpcmodel.CreateSubnetRequest{
		Body: &vpcmodel.CreateSubnetRequestBody{
			Subnet: &vpcmodel.CreateSubnetOption{
				Cidr:         *cidr,
				Name:         *name,
				VpcId:        *vpcId,
				GatewayIp:    *gatewayIp,
				PrimaryDns:   primaryDns,
				SecondaryDns: secondaryDns,
				DnsList:      dnsList,
			},
		},
	}

	rsp, err := vpcClient.CreateSubnet(req)
	if err != nil {
		return "", err
	}

	return rsp.Subnet.Id, nil
}

// WaitSubnetReady 等待子网创建完成
func WaitSubnetReady(vpcClient *vpc.VpcClient, subnetId *string) error {
	req := &vpcmodel.ShowSubnetRequest{
		SubnetId: *subnetId,
	}

	for {
		// TODO(wangjun): timeout
		rsp, err := vpcClient.ShowSubnet(req)
		if err != nil {
			return err
		}

		if rsp.Subnet.Status == vpcmodel.GetSubnetStatusEnum().UNKNOWN {
			time.Sleep(DefaultCallIntervalTimes * time.Second)
			continue
		}

		if rsp.Subnet.Status == vpcmodel.GetSubnetStatusEnum().ERROR {
			return errors.NewError("Subnet status error")
		}

		// 创建已成功
		return nil
	}
}

// WaitSubnetDeleted 等待子网删除
func WaitSubnetDeleted(vpcClient *vpc.VpcClient, subnetId *string) error {
	req := &vpcmodel.ListSubnetsRequest{}

	for {
		// TODO(wangjun): timeout
		rsp, err := vpcClient.ListSubnets(req)
		if err != nil {
			return err
		}

		exists := false
		for _, s := range *rsp.Subnets {
			if s.Id == *subnetId {
				exists = true
				break
			}
		}
		if !exists {
			return nil
		}

		time.Sleep(DefaultCallIntervalTimes * time.Second)
	}
}

// CreateBandwidth 创建带宽
func CreateBandwidth(eipClient *eip.EipClient, name *string, size *int32, bandwidthType *string) (string,
	error) {
	var req *eipmodel.CreateSharedBandwidthRequest
	if *bandwidthType == "" {
		req = &eipmodel.CreateSharedBandwidthRequest{
			Body: &eipmodel.CreateSharedBandwidhRequestBody{
				Bandwidth: &eipmodel.CreateSharedBandwidthOption{
					Name: *name,
					Size: *size,
				},
			},
		}
	} else {
		req = &eipmodel.CreateSharedBandwidthRequest{
			Body: &eipmodel.CreateSharedBandwidhRequestBody{
				Bandwidth: &eipmodel.CreateSharedBandwidthOption{
					Name:          *name,
					Size:          *size,
					BandwidthType: bandwidthType,
				},
			},
		}
	}

	rsp, err := eipClient.CreateSharedBandwidth(req)
	if err != nil {
		return "", err
	}

	return *rsp.Bandwidth.Id, nil
}

// CreateEip 创建EIP
func CreateEip(eipClient *eip.EipClient, name *string, size *int32, bandwidthType *string) (string,
	error) {
	req := &eipmodel.CreatePublicipRequest{}
	bandWidthBody := &eipmodel.CreatePublicipBandwidthOption{
		Name:      name,
		Size:      size,
		ShareType: eipmodel.GetCreatePublicipBandwidthOptionShareTypeEnum().PER,
	}
	publicIpBody := &eipmodel.CreatePublicipOption{
		Type: *bandwidthType,
	}
	req.Body = &eipmodel.CreatePublicipRequestBody{
		Publicip:  publicIpBody,
		Bandwidth: bandWidthBody,
	}

	rsp, err := eipClient.CreatePublicip(req)
	if err != nil {
		return "", err
	}
	return *rsp.Publicip.Id, nil
}

// DeleteEip 删除带宽
func DeleteEip(eipClient *eip.EipClient, eipId string) error {
	request := &eipmodel.DeletePublicipRequest{
		PublicipId: eipId,
	}
	_, err := eipClient.DeletePublicip(request)
	if err != nil {
		return err
	}
	return nil
}

// CreateSecurityGroup 创建安全组
func CreateSecurityGroup(vpcClient VpcV2, name *string, enterpriseProject *string) (string,
	error) {
	req := &vpcmodel.CreateSecurityGroupRequest{
		Body: &vpcmodel.CreateSecurityGroupRequestBody{
			SecurityGroup: &vpcmodel.CreateSecurityGroupOption{
				Name:                *name,
				EnterpriseProjectId: enterpriseProject,
			},
		},
	}

	rsp, err := vpcClient.CreateSecurityGroup(req)
	if err != nil {
		return "", err
	}

	return rsp.SecurityGroup.Id, nil
}

// GetSecurityGroup 获取安全组
func GetSecurityGroup(vpcClient *vpc.VpcClient, name *string) (string, error) {
	req := &vpcmodel.ListSecurityGroupsRequest{}
	rsp, err := vpcClient.ListSecurityGroups(req)
	if err != nil {
		return "", err
	}

	for _, s := range *rsp.SecurityGroups {
		if s.Name == *name {
			return s.Id, nil
		}
	}

	return "", nil
}

// DeleteSecurityGroup 删除安全组
func DeleteSecurityGroup(vpcClient *vpc.VpcClient, securityGroupId *string) error {
	req := &vpcmodel.DeleteSecurityGroupRequest{
		SecurityGroupId: *securityGroupId,
	}

	_, err := vpcClient.DeleteSecurityGroup(req)
	if err != nil {
		// 如果已经删除了, 就不需要继续删了
		newErr, ok := err.(*sdkerr.ServiceResponseError)
		if ok && (newErr.StatusCode == HttpCodeNotFound) {
			err = nil
		}
	}

	return err
}

// CreateSecurityGroupRule 创建安全组规则, 支持重入
func CreateSecurityGroupRule(vpcClient *vpc.VpcClient, securityGroupId *string, direction *string, ethertype *string,
	protocol *string, portRangeMin *int32, portRangeMax *int32, remoteIpPrefix *string) (string, error) {
	req := &vpcmodel.CreateSecurityGroupRuleRequest{
		Body: &vpcmodel.CreateSecurityGroupRuleRequestBody{
			SecurityGroupRule: &vpcmodel.CreateSecurityGroupRuleOption{
				SecurityGroupId: *securityGroupId,
				Direction:       *direction,
				Ethertype:       ethertype,
				Protocol:        protocol,
				PortRangeMin:    portRangeMin,
				PortRangeMax:    portRangeMax,
				RemoteIpPrefix:  remoteIpPrefix,
			},
		},
	}

	rsp, err := vpcClient.CreateSecurityGroupRule(req)
	if err != nil {
		// 如果创建冲突, 说明已经创建过
		newErr, ok := err.(*sdkerr.ServiceResponseError)
		if ok && (newErr.StatusCode == HttpCodeConfilct) {
			// 查询安全组规则
			return GetSecurityGroupRule(vpcClient, securityGroupId, direction, ethertype, protocol, portRangeMin,
				portRangeMax, remoteIpPrefix)
		}

		return "", err
	}

	return rsp.SecurityGroupRule.Id, nil
}

// GetSecurityGroupRules 获取所有的安全组规则
func GetSecurityGroupRules(vpcClient VpcV2, securityGroupId *string) (*[]vpcmodel.SecurityGroupRule, error) {
	// 查询所有的规则
	req := &vpcmodel.ListSecurityGroupRulesRequest{
		SecurityGroupId: securityGroupId,
	}

	rsp, err := vpcClient.ListSecurityGroupRules(req)
	if err != nil {
		return nil, err
	}

	return rsp.SecurityGroupRules, nil
}

// GetSecurityGroupRule 获取安全组规则
func GetSecurityGroupRule(vpcClient *vpc.VpcClient, securityGroupId *string, direction *string, ethertype *string,
	protocol *string, portRangeMin *int32, portRangeMax *int32, remoteIpPrefix *string) (string, error) {
	// 查询所有的规则
	req := &vpcmodel.ListSecurityGroupRulesRequest{
		SecurityGroupId: securityGroupId,
	}
	rsp, err := vpcClient.ListSecurityGroupRules(req)
	if err != nil {
		return "", err
	}

	// 匹配需要查找的规则ID
	for _, r := range *rsp.SecurityGroupRules {
		if strings.EqualFold(r.Protocol, *protocol) &&
			strings.EqualFold(r.Direction, *direction) &&
			strings.EqualFold(r.Ethertype, *ethertype) &&
			r.PortRangeMin == *portRangeMin &&
			r.PortRangeMax == *portRangeMax &&
			r.RemoteIpPrefix == *remoteIpPrefix {
			return r.Id, nil
		}
	}

	return "", errors.NewError(errors.SecurityGroupRuleNotFound)
}

// DeleteSecurityGroupRule 删除安全组规则
func DeleteSecurityGroupRule(vpcClient *vpc.VpcClient, id string) error {
	req := &vpcmodel.DeleteSecurityGroupRuleRequest{
		SecurityGroupRuleId: id,
	}
	_, err := vpcClient.DeleteSecurityGroupRule(req)
	if err != nil {
		// 如果已经删除了, 就不需要继续删了
		newErr, ok := err.(*sdkerr.ServiceResponseError)
		if ok && (newErr.StatusCode == HttpCodeNotFound) {
			err = nil
		}
	}

	return err
}

// CreateImageEcs 创建ECS以打包镜像
func CreateImageEcs(ecsClient *ecs.EcsClient, vpcId string, subNetId string, name string,
	userData string, imageId string, eipId string, ltsAgency string) (string, error) {
	ecsRequest := &ecsmodel.CreateServersRequest{}
	volumeSize := int32(setting.ImageDiskSize)
	rootVolumeServer := &ecsmodel.PrePaidServerRootVolume{
		Volumetype: ecsmodel.GetPrePaidServerRootVolumeVolumetypeEnum().SATA,
		Size:       &volumeSize,
	}
	var listNicsServer = []ecsmodel.PrePaidServerNic{
		{
			SubnetId: subNetId,
		},
	}
	var preEip = ecsmodel.PrePaidServerPublicip{
		Id: &eipId,
	}
	mp := map[string]string{
		"agency_name": ltsAgency,
	}
	serverBody := &ecsmodel.PrePaidServer{
		ImageRef:   imageId,
		Metadata:   mp,
		FlavorRef:  setting.ImageFlavor,
		Name:       name + "image",
		Vpcid:      vpcId,
		Nics:       listNicsServer,
		RootVolume: rootVolumeServer,
		UserData:   &userData,
		Publicip:   &preEip,
	}
	ecsRequest.Body = &ecsmodel.CreateServersRequestBody{
		Server: serverBody,
	}
	ecsResponse, err := ecsClient.CreateServers(ecsRequest)
	if err != nil {
		return "", err
	}
	err = WaitEcsReady(ecsClient, *ecsResponse.JobId)
	if err != nil {
		return "", err
	}
	ecsId := (*ecsResponse.ServerIds)[0]
	return ecsId, nil
}

func WaitEcsReady(ecsClient *ecs.EcsClient, jobId string) error {
	req := &ecsmodel.ShowJobRequest{
		JobId: jobId,
	}
	for {
		response, err := ecsClient.ShowJob(req)
		if err != nil {
			return err
		}
		if *response.Status == ecsmodel.GetShowJobResponseStatusEnum().SUCCESS {
			break
		}
		time.Sleep(5 * time.Second)
		continue
	}
	return nil
}

// WaitImageEcsShutDown 等待镜像ECS就绪 （执行完环境构建脚本并关机）
func WaitImageEcsShutDown(ecsClient *ecs.EcsClient, ecsId string) error {
	request := &ecsmodel.ShowServerRequest{}
	request.ServerId = ecsId
	for {
		response, err := ecsClient.ShowServer(request)
		if err != nil {
			return err
		}
		if response.Server.Status == "SHUTOFF" {
			return nil
		}

		time.Sleep(10 * time.Second)
		continue
	}
}

// DeleteEcs 删除ECS资源
func DeleteEcs(ecsClient *ecs.EcsClient, ecsId string) (string, error) {
	request := &ecsmodel.DeleteServersRequest{}
	var listServersbody = []ecsmodel.ServerId{
		{
			Id: ecsId,
		},
	}
	request.Body = &ecsmodel.DeleteServersRequestBody{
		Servers: listServersbody,
	}
	response, err := ecsClient.DeleteServers(request)
	if err != nil {
		return "", err
	}
	return *response.JobId, nil
}

func WaitEcsDeleted(ecsClient *ecs.EcsClient, jobId string) error {
	request := &ecsmodel.ShowJobRequest{}
	request.JobId = jobId
	for {
		response, err := ecsClient.ShowJob(request)

		if *response.Status == ecsmodel.GetShowJobResponseStatusEnum().SUCCESS {
			return nil
		}

		if *response.Status == ecsmodel.GetShowJobResponseStatusEnum().FAIL {
			return err
		}

		time.Sleep(10 * time.Second)
		continue
	}
}

// WaitImsReady 等待镜像创建完成
func WaitImsReady(imsClient *ims.ImsClient, jobId string) (string, error) {
	request := &imsmodel.ShowJobRequest{}
	request.JobId = jobId
	for {
		response, err := imsClient.ShowJob(request)
		if err != nil {
			return "", err
		}
		if *response.Status == imsmodel.GetShowJobResponseStatusEnum().SUCCESS {
			return *response.Entities.ImageId, nil
		}

		if *response.Status == imsmodel.GetShowJobResponseStatusEnum().FAIL {
			return "", errors.NewError("create build failed")
		}

		time.Sleep(10 * time.Second)
		continue
	}
}

// DeleteBuildInIMS 删除应用包镜像
func DeleteBuildInIMS(imsClient *ims.ImsClient, imageId string) error {
	req := &imsmodel.GlanceDeleteImageRequest{
		ImageId: imageId,
	}
	_, err := imsClient.GlanceDeleteImage(req)
	if err != nil {
		newErr, ok := err.(*sdkerr.ServiceResponseError)
		if ok && (newErr.StatusCode == HttpCodeNotFound) {
			logger.R.Info("This image is not in IMS")
			err = nil
		}
	}

	return err
}

// UploadFileToOBS 上传文件至OBS
func UploadFileToOBS(obsClient *obs.ObsClient, file []byte, bucketName string, bucketKey string) (string, error) {
	reader := bytes.NewReader(file)
	finalKey := "/" + bucketName + "/" + bucketKey
	input := &obs.PutObjectInput{}
	input.Bucket = bucketName
	input.Key = bucketKey
	input.Body = reader
	_, err := obsClient.PutObject(input)
	if err != nil {
		return "", err
	}
	return finalKey, nil
}

func ListOBSBucket(obsClient *obs.ObsClient) ([]string, error) {
	var list []string
	output, err := obsClient.ListBuckets(nil)
	if err != nil {
		return list, err
	}
	for _, val := range output.Buckets {
		list = append(list, val.Name)
	}
	return list, nil
}

func ListIMSImage(imsClient *ims.ImsClient) (build.ImageList, error) {
	var res build.ImageList
	var imageList []build.Image
	req := &imsmodel.ListImagesRequest{}
	imageType := imsmodel.GetListImagesRequestImagetypeEnum().PRIVATE
	req.Imagetype = &imageType
	output, err := imsClient.ListImages(req)
	if err != nil {
		return res, err
	}
	for _, val := range *(output.Images) {
		tmpImage := build.Image{
			Name:    val.Name,
			Id:      val.Id,
			Version: *val.OsVersion,
		}
		imageList = append(imageList, tmpImage)
	}
	res.Count = len(imageList)
	res.ImageList = imageList
	return res, nil
}

func GetIMSOperatingSystem(imsClient *ims.ImsClient, imageId string) (string, error) {
	req := &imsmodel.ListImagesRequest{
		Id: &imageId,
	}
	rep, err := imsClient.ListImages(req)
	if err != nil {
		return "", err
	}
	imsList := *rep.Images
	if len(imsList) == 0 {
		return "", nil
	}
	return *imsList[0].OsVersion, nil
}

func ListVpcList(vpcClient *vpc.VpcClient) (build.VpcList, error) {
	var res build.VpcList
	var vpcList []build.Vpc
	req := &vpcmodel.ListVpcsRequest{}
	output, err := vpcClient.ListVpcs(req)
	if err != nil {
		return res, err
	}
	for _, val := range *(output.Vpcs) {
		tmpVpc := build.Vpc{
			Name: val.Name,
			Id:   val.Id,
		}
		vpcList = append(vpcList, tmpVpc)
	}
	res.Count = len(vpcList)
	res.VpcList = vpcList
	return res, nil
}

func ListSubnetList(vpcClient *vpc.VpcClient, vpcId string) (build.SubnetList, error) {
	var res build.SubnetList
	var subnetList []build.Subnet
	req := &vpcmodel.ListSubnetsRequest{}
	id := vpcId
	req.VpcId = &id
	output, err := vpcClient.ListSubnets(req)
	if err != nil {
		return res, err
	}
	for _, val := range *(output.Subnets) {
		tmpSubnet := build.Subnet{
			Name: val.Name,
			Id:   val.Id,
		}
		subnetList = append(subnetList, tmpSubnet)
	}
	res.Count = len(subnetList)
	res.VpcList = subnetList
	return res, nil
}
