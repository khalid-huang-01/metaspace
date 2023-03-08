package build

import (
	"fleetmanager/api/service/constants"
	"fleetmanager/client"
	"fleetmanager/db/dao"
	"fleetmanager/db/dbm"
	"fleetmanager/setting"
	"fleetmanager/workflow/components"
	"fleetmanager/workflow/directer"
	"fleetmanager/workflow/meta"
	"github.com/google/uuid"
	ims "github.com/huaweicloud/huaweicloud-sdk-go-v3/services/ims/v2"
	imsmodel "github.com/huaweicloud/huaweicloud-sdk-go-v3/services/ims/v2/model"
	"time"
)

type CreateBuildImageTask struct {
	components.BaseTask
}

// Execute 执行镜像创建操作
func (t *CreateBuildImageTask) Execute(ctx *directer.ExecuteContext) (output interface{}, err error) {
	defer func() { t.ExecNext(output, err) }()

	ecsId := t.Directer.GetContext().Get(directer.WfKeyBuildECSId).ToString("")
	regionId := t.Directer.GetContext().Get(directer.WfKeyBuildRegion).ToString("")
	buildName := t.Directer.GetContext().Get(directer.WfKeyBuildName).ToString("")
	buildVersion := t.Directer.GetContext().Get(directer.WfKeyBuildVersion).ToString("")
	buildId := t.Directer.GetContext().Get(directer.WfKeyBuildUuid).ToString("")
	agencyName := t.Directer.GetContext().Get(directer.WfKeyResourceAgencyName).ToString("")
	resourceDomainId := t.Directer.GetContext().Get(directer.WfKeyResourceDomainId).ToString("")
	resourceProjectId := t.Directer.GetContext().Get(directer.WfKeyResourceProjectId).ToString("")
	originProjectId := t.Directer.GetContext().Get(directer.WfKeyOriginProjectId).ToString("")
	err = buildBegin(buildId)
	if err != nil {
		return nil, err
	}

	imsClient, err := client.GetAgencyIMSClient(regionId, resourceProjectId, agencyName, resourceDomainId)
	if err != nil {
		return nil, err
	}

	imsJobId, err := CreateImage(imsClient, buildName+buildVersion, ecsId)
	if err != nil {
		return nil, err
	}

	imsId, err := client.WaitImsReady(imsClient, imsJobId)
	if err != nil {
		return nil, err
	}
	err = buildReady(buildId, imsId, originProjectId, regionId)
	if err != nil {
		return nil, err
	}
	return nil, nil
}

// NewCreateBuildImageTask 新建镜像打包任务
func NewCreateBuildImageTask(meta meta.TaskMeta, directer directer.Directer, step int) components.Task {
	t := &CreateBuildImageTask{
		components.NewBaseTask(meta, directer, step),
	}

	return t
}

func CreateImage(imsClient *ims.ImsClient, imageName string, ecsId string) (string, error) {
	request := &imsmodel.CreateImageRequest{}
	instanceIdCreateImageRequestBody := ecsId
	request.Body = &imsmodel.CreateImageRequestBody{
		Name:                imageName,
		InstanceId:          &instanceIdCreateImageRequestBody,
		EnterpriseProjectId: &setting.EnterpriseProject,
	}
	response, err := imsClient.CreateImage(request)
	if err != nil {
		return "", err
	}
	return *response.JobId, nil
}

func buildBegin(buildId string) error {

	build := &dao.Build{
		Id:    buildId,
		State: constants.BuildImageInitialized,
	}

	_, err := dbm.Ormer.Update(build, "State", "ImageId")
	if err != nil {
		return err
	}
	return nil
}

func buildReady(buildId string, imsId string, projectId string, regionId string) error {

	build := &dao.Build{
		Id:      buildId,
		State:   constants.BuildStateReady,
		ImageId: imsId,
	}

	_, err := dbm.Ormer.Update(build, "State", "ImageId")
	if err != nil {
		return err
	}

	u, _ := uuid.NewUUID()
	tmpId := u.String()
	buildImage := &dao.BuildImage{
		Id:            tmpId,
		ProjectId:     projectId,
		CreateTime:    time.Now().UTC(),
		ImageId:       imsId,
		ImageRegionId: regionId,
		BuildId:       buildId,
	}
	_, err = dbm.Ormer.Insert(buildImage)
	if err != nil {
		return err
	}

	return nil
}
