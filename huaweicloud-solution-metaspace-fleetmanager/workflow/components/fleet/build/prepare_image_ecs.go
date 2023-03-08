package build

import (
	"encoding/base64"
	"fleetmanager/client"
	"fleetmanager/setting"
	"fleetmanager/workflow/components"
	"fleetmanager/workflow/directer"
	"fleetmanager/workflow/meta"
	"github.com/huaweicloud/huaweicloud-sdk-go-obs/obs"
	ecs "github.com/huaweicloud/huaweicloud-sdk-go-v3/services/ecs/v2"
	ims "github.com/huaweicloud/huaweicloud-sdk-go-v3/services/ims/v2"
	"strings"
)

type PrePareImageEcsTask struct {
	components.BaseTask
}

// 用于构建镜像的临时ECS信息
type TmpEcs struct {
	userData        string
	tmpImageId      string
	regionId        string
	profileRegionId string
	vpcId           string
	subnetId        string
	auxUrl          string
	appBucket       string
	appKey          string
	scriptUrl       string
	tmpResourceName string
	eipId           string
}

// Execute 执行ECS资源准备
func (t *PrePareImageEcsTask) Execute(ctx *directer.ExecuteContext) (output interface{}, err error) {
	defer func() { t.ExecNext(output, err) }()
	tmpEcs := TmpEcs{}
	tmpEcs.regionId = t.Directer.GetContext().Get(directer.WfKeyBuildRegion).ToString("")
	tmpEcs.appBucket = t.Directer.GetContext().Get(directer.WfKeyBuildBucket).ToString("")
	tmpEcs.appKey = t.Directer.GetContext().Get(directer.WfKeyBuildKey).ToString("")
	tmpEcs.auxUrl = setting.AuxProxyPath
	tmpEcs.scriptUrl = setting.ScriptPath
	tmpEcs.profileRegionId = setting.ProfileStorageRegion
	tmpEcs.eipId = t.Directer.GetContext().Get(directer.WfKeyBandwidthId).ToString("")
	tmpEcs.vpcId = t.Directer.GetContext().Get(directer.WfKeyBuildVpcId).ToString("")
	tmpEcs.subnetId = t.Directer.GetContext().Get(directer.WfKeyBuildSubnetId).ToString("")

	var ecsId = t.Directer.GetContext().Get(directer.WfKeyBuildECSId).ToString("")
	version := t.Directer.GetContext().Get(directer.WfKeyBuildVersion).ToString("")
	tmpEcs.tmpResourceName = t.Directer.GetContext().Get(directer.WfKeyBuildName).ToString("") + version
	imageRef := t.Directer.GetContext().Get(directer.WfKeyImageRef).ToString("")
	resourceProjectId := t.Directer.GetContext().Get(directer.WfKeyResourceProjectId).ToString("")
	regionId := t.Directer.GetContext().Get(directer.WfKeyBuildRegion).ToString("")
	agencyName := t.Directer.GetContext().Get(directer.WfKeyResourceAgencyName).ToString("")
	resourceDomainId := t.Directer.GetContext().Get(directer.WfKeyResourceDomainId).ToString("")
	buildName := t.Directer.GetContext().Get(directer.WfKeyBuildName).ToString("")
	ltsAgency := t.Directer.GetContext().Get(directer.WfKeyIamAgencyName).ToString("")

	// 获取创建ECS所需信息
	ecsClient, err := client.GetAgencyEcsClient(regionId, resourceProjectId, agencyName, resourceDomainId)
	if err != nil {
		return nil, err
	}
	if ecsId == "" {
		resourceObsClient, err := client.GetAgencyObsClient(regionId, agencyName, resourceDomainId)
		if err != nil {
			return nil, err
		}

		originObsClient, err := client.GetOriginObsClient(tmpEcs.profileRegionId)

		tmpEcs.userData, err = CreateUserData(originObsClient, resourceObsClient, &tmpEcs, buildName)
		if err != nil {
			return nil, err
		}

		// 获取公共镜像id
		imsClient, err := client.GetAgencyIMSClient(regionId, resourceProjectId, agencyName, resourceDomainId)
		tmpEcs.tmpImageId, err = GetTmpImage(imsClient, imageRef)

		// 创建ECS
		ecsId, err = CreateECS(ecsClient, tmpEcs, ltsAgency)
		if err != nil {
			return nil, err
		}

		if err = t.Directer.GetContext().Set(directer.WfKeyBuildECSId, ecsId); err != nil {
			return nil, err
		}
	}

	// 等待ECS创建完成
	if err := client.WaitImageEcsShutDown(ecsClient, ecsId); err != nil {
		return nil, err
	}

	return nil, nil
}

func (t *PrePareImageEcsTask) Rollback(*directer.ExecuteContext) (output interface{}, err error) {
	defer func() { t.RollbackPrev(output, err) }()

	regionId := t.Directer.GetContext().Get(directer.WfKeyBuildRegion).ToString("")
	ecsId := t.Directer.GetContext().Get(directer.WfKeyBuildECSId).ToString("")
	agencyName := t.Directer.GetContext().Get(directer.WfKeyResourceAgencyName).ToString("")
	resourceDomainId := t.Directer.GetContext().Get(directer.WfKeyResourceDomainId).ToString("")
	resourceProjectId := t.Directer.GetContext().Get(directer.WfKeyResourceProjectId).ToString("")
	ecsClient, err := client.GetAgencyEcsClient(regionId, resourceProjectId, agencyName, resourceDomainId)

	jobId, err := client.DeleteEcs(ecsClient, ecsId)
	if err != nil {
		return nil, err
	}
	err = client.WaitEcsDeleted(ecsClient, jobId)
	if err != nil {
		return nil, err
	}
	return nil, nil
}

// NewPrePareImageEcsTask 新建准备镜像构建ECS任务
func NewPrePareImageEcsTask(meta meta.TaskMeta, directer directer.Directer, step int) components.Task {
	t := &PrePareImageEcsTask{
		components.NewBaseTask(meta, directer, step),
	}

	return t
}

// GetSingleUrl 构建部署脚本、app、auxproxy临时url
func GetSingleUrl(obsClient *obs.ObsClient, url string) (string, error) {
	object := strings.SplitN(url, "/", 2)
	getObjectInput := &obs.CreateSignedUrlInput{}
	getObjectInput.Method = obs.HttpMethodGet
	getObjectInput.Expires = 3600
	getObjectInput.Bucket = object[0]
	getObjectInput.Key = object[1]
	getObjectOutput, err := obsClient.CreateSignedUrl(getObjectInput)
	var tmpUrl string
	if err == nil {
		tmpUrl = getObjectOutput.SignedUrl
	} else {
		return "", nil
	}
	return tmpUrl, nil
}

// CreateUserData 构建注入数据，用于镜像环境构建
func CreateUserData(originObsClient *obs.ObsClient, resourceObsClient *obs.ObsClient,
	ecs *TmpEcs, buildName string) (string, error) {

	tmpScriptUrl, err := GetSingleUrl(originObsClient, ecs.scriptUrl)
	if err != nil {
		return "", nil
	}

	appUrl := ecs.appBucket + "/" + ecs.appKey
	appFile := strings.Split(appUrl, "/")
	tmpAppUrl, err := GetSingleUrl(resourceObsClient, appUrl)
	if err != nil {
		return "", nil
	}

	auxFile := strings.Split(ecs.auxUrl, "/")
	tmpAuxUrl, err := GetSingleUrl(originObsClient, ecs.auxUrl)
	if err != nil {
		return "", nil
	}

	k := setting.LTSIp + "." + ecs.regionId
	ltsIp := setting.Config.Get(k).ToString("")

	key := strings.Split(ecs.scriptUrl, "/")
	fileName := key[len(key)-1]

	var userData = "#!/bin/bash" +
		"\ncd /tmp" +
		"\nwget '" + tmpScriptUrl + "' -O " + fileName +
		"\nbash /tmp/" + fileName + " '" + ecs.regionId + "' '" +
		tmpAppUrl + "' '" + tmpAuxUrl + "' '" + ltsIp + "' '" +
		appFile[len(appFile)-1] + "' '" + auxFile[len(auxFile)-1] + "' '" + buildName +
		"'\nrm -rf /tmp/" + fileName
	byteData := []byte(userData)
	baseData := base64.StdEncoding.EncodeToString(byteData)
	return baseData, nil
}

// GetTmpImage 获取操作系统镜像id 作为镜像打包ECS的操作系统
func GetTmpImage(imsClient *ims.ImsClient, imageRef string) (string, error) {
	imageId, err := client.GetImageId(imsClient, imageRef)
	if err != nil {
		return "", err
	}
	return imageId, nil
}

// CreateECS 创建ECS用于镜像打包
func CreateECS(ecsClient *ecs.EcsClient, ecs TmpEcs, ltsAgency string) (string, error) {
	ecsId, err := client.CreateImageEcs(ecsClient, ecs.vpcId, ecs.subnetId, ecs.tmpResourceName,
		ecs.userData, ecs.tmpImageId, ecs.eipId, ltsAgency)
	if err != nil {
		return "", err
	}
	return ecsId, nil
}
