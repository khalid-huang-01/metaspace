// Copyright (c) Huawei Technologies Co., Ltd. 2022-2022. All rights reserved.

// 应用包数据表定义
package dao

import (
	"fleetmanager/api/model/build"
	"fleetmanager/api/params"
	"fleetmanager/api/service/constants"
	"fleetmanager/db/dbm"
	"github.com/beego/beego/v2/client/orm"
	"github.com/google/uuid"
	"strconv"
	"strings"
	"time"
)

// Build
// @Description:
type Build struct {
	Id                string    `orm:"column(id);size(64);pk" json:"id"`
	ProjectId         string    `orm:"column(project_id);size(64)" json:"project_id"`
	Name              string    `orm:"column(name);size(50)" json:"name"`
	Description       string    `orm:"column(description);size(100)" json:"description"`
	State             string    `orm:"column(state);size(32)" json:"state"`
	CreationTime      time.Time `orm:"column(creation_time);type(datetime);auto_now_add" json:"creation_time"`
	UpdateTime        time.Time `orm:"column(update_time);type(datetime);auto_now" json:"update_time"`
	TerminationTime   time.Time `orm:"column(termination_time);auto_now" json:"termination_time"`
	ImageId           string    `orm:"column(image_id);size(64)" json:"image_id"`
	ImageRegion       string    `orm:"column(image_region);size(50)" json:"image_region"`
	StorageBucketName string    `orm:"column(bucket_name);size(100)" json:"storage_bucket_name"`
	StorageKey        string    `orm:"column(key);size(100)" json:"storage_key"`
	StorageRegion     string    `orm:"column(storage_region);size(50)" json:"storage_region"`
	OperatingSystem   string    `orm:"column(operating_system);size(50)" json:"operating_system"`
	Version           string    `orm:"column(version);size(50)" json:"version"`
	Size              int64     `orm:"column(size);size(32)" json:"size"`
}

// @Title GetBuildById
// @Description  Get build by build ID from DB
// @Author wangnannan 2022-05-07 10:18:23 ${time}
// @Param id
// @Return *Build
// @Return error
func GetBuildById(id string, projectId string) (*Build, error) {
	var b Build
	err := dbm.Ormer.QueryTable(BuildTable).Filter("Id", id).Filter("ProjectId", projectId).One(&b)
	if err != nil {
		return nil, err
	}

	return &b, nil
}

// @Title CheckBuildByName
// @Description  Get build by build name from DB
// @Author wangnannan 2022-05-07 10:18:33 ${time}
// @Param name
// @Return bool
func CheckBuildByName(name string, version string, projectId string) bool {

	existName := dbm.Ormer.QueryTable(BuildTable).Filter("Name", name).
		Filter("Version", version).Filter("ProjectId", projectId).Exist()

	return existName
}

// @Title InsertBuild
// @Description  insert build to DB
// @Author wangnannan 2022-05-07 10:18:43 ${time}
// @Param r
// @Param region
// @Param project
// @Return *build.Build
// @Return error
func InsertBuild(r build.CreateRequest, region string, project string, size int64) (*Build, error) {

	to, err := dbm.Ormer.Begin()
	if err != nil {
		return nil, err
	}

	u, _ := uuid.NewUUID()
	buildId := u.String()
	bd := &Build{
		Id:                buildId,
		ProjectId:         project,
		Name:              r.Name,
		Description:       r.Description,
		State:             constants.BuildStateInitialized,
		Version:           r.Version,
		StorageBucketName: r.StorageLocation.BucketName,
		StorageKey:        r.StorageLocation.BucketKey,
		ImageRegion:       r.Region,
		CreationTime:      time.Now().UTC(),
		UpdateTime:        time.Now().UTC(),
		StorageRegion:     region,
		OperatingSystem:   r.OperatingSystem,
		Size:              size,
	}

	_, err = to.Insert(bd)
	if err != nil {
		_ = to.Rollback()
		return nil, err
	}

	err = to.Commit()
	if err != nil {
		_ = to.Rollback()
		return nil, err
	}

	return bd, nil
}

func InsertBuildByImage(r build.CreateByImageRequest, region string, project string, osVersion string) (*Build, error) {
	to, err := dbm.Ormer.Begin()
	if err != nil {
		return nil, err
	}

	u, _ := uuid.NewUUID()
	buildId := u.String()
	bd := &Build{
		Id:              buildId,
		ProjectId:       project,
		Name:            r.Name,
		Description:     r.Description,
		State:           constants.BuildStateReady,
		Version:         r.Version,
		ImageId:         r.ImageId,
		ImageRegion:     r.Region,
		CreationTime:    time.Now().UTC(),
		UpdateTime:      time.Now().UTC(),
		StorageRegion:   region,
		OperatingSystem: osVersion,
	}

	_, err = to.Insert(bd)
	if err != nil {
		_ = to.Rollback()
		return nil, err
	}

	err = to.Commit()
	if err != nil {
		_ = to.Rollback()
		return nil, err
	}

	err = InsertBuildImage(r, project, region, buildId)
	if err != nil {
		return nil, err
	}

	return bd, nil
}

func InsertBuildImage(r build.CreateByImageRequest, projectId string, region string, buildId string) error {
	to, err := dbm.Ormer.Begin()
	if err != nil {
		return err
	}

	uu, _ := uuid.NewUUID()
	buildImageId := uu.String()
	bdi := &BuildImage{
		Id:            buildImageId,
		ProjectId:     projectId,
		ImageId:       r.ImageId,
		BuildId:       buildId,
		ImageRegionId: region,
		CreateTime:    time.Now().UTC(),
	}

	_, err = to.Insert(bdi)
	if err != nil {
		_ = to.Rollback()
		return err
	}

	err = to.Commit()
	if err != nil {
		_ = to.Rollback()
		return err
	}

	return nil
}

// @Title DeleteByBuild
// @Description  delete build in DB
// @Author wangnannan 2022-05-07 10:18:55 ${time}
// @Param buildId
// @Param projectId
// @Return error
func DeleteByBuild(buildId string, projectId string) error {

	bd := &Build{
		Id:        buildId,
		ProjectId: projectId,
		State:     constants.BuildDeleted,
	}

	_, err := dbm.Ormer.Update(bd, "State")
	if err != nil {
		return err
	}
	return nil
}

// @Title CheckBuildInUsedSystem
// @Description  check if build is relate to fleet
// @Author wangnannan 2022-05-07 10:19:08 ${time}
// @Param buildId
// @Return error
func CheckBuildInUsedSystem(buildId string, projectId string) (bool, error) {

	to, err := dbm.Ormer.Begin()
	if err != nil {
		return false, err
	}

	fleetNum, err := to.QueryTable(FleetTable).Filter("BuildId", buildId).Filter("ProjectId", projectId).
		Filter("State__in", FleetStateActive, FleetStateCreating, FleetStateDeleting, FleetStateError).Count()
	if err != nil {
		return false, err
	} else {
		if fleetNum != 0 {
			return false, nil
		}
	}

	return true, nil
}

// @Title buildBuildModel
// @Description
// @Author wangnannan 2022-05-07 10:19:18 ${time}
// @Param bd
// @Return build.Build
func buildBuildModel(bd *Build) build.Build {
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

// @Title CheckBuildByBucket
// @Description  Get build by build name from DB
// @Author wangnannan 2022-05-07 10:19:24 ${time}
// @Param bucketName
// @Param bucketKey
// @Return bool
func CheckBuildByBucket(bucketName string, bucketKey string) bool {

	existName := dbm.Ormer.QueryTable(BuildTable).Filter("StorageBucketName", bucketName).
		Filter("StorageKey", bucketKey).
		Filter("State__in", constants.BuildImageInitialized, constants.BuildStateReady,
			constants.BuildImageInitialized).Exist()

	return existName
}

func GetBuildImage(buildId string, regionId string, projectId string) (string, error) {
	var ds []BuildImage
	_, err := dbm.Ormer.QueryTable(BuildImageTable).Filter("ProjectId", projectId).Filter(
		"BuildId", buildId).Filter("ImageRegionId", regionId).All(&ds)
	if err != nil {
		return "", err
	}

	// 该位置传入nil不会报错 len(nil)=0
	if len(ds) > 1 {
		// 记录告警日志
	}

	if len(ds) != 0 {
		return ds[0].ImageId, nil
	} else {
		return "", nil
	}
}

// @Title GetBuildCount
// @Description  Exist build count
// @Author wangnannan 2022-05-07 10:19:34 ${time}
// @Param projectid
// @Return int64
// @Return error
func GetBuildCount(projectid string) (int64, error) {
	number, err := dbm.Ormer.QueryTable(BuildTable).Filter("ProjectId", projectid).Count()
	if err != nil {
		return params.DefaultNumber, err
	}

	return number, nil
}

// BuildImage
// @Description:
type BuildImage struct {
	Id            string    `orm:"column(id);size(128);pk"`
	ProjectId     string    `orm:"column(project_id);size(64)"`
	CreateTime    time.Time `orm:"column(create_time);type(datetime);auto_now_add"`
	ImageId       string    `orm:"column(image_id);size(50)"`
	ImageRegionId string    `orm:"column(image_region_id);size(50)"`
	BuildId       string    `orm:"column(build_id);size(128)"`
}

func QueryBuildByCondition(projectId string, q build.QueryRequest, offset int, limit int) ([]Build, int64, error) {
	var list []Build

	qs := dbm.Ormer.QueryTable(BuildTable)

	condition := orm.NewCondition()
	condition = condition.And("project_id", projectId)
	if q.Name != "" {
		condition = condition.And("name__contains", q.Name)
	}
	if q.State != "" {
		condition = condition.And("state", q.State)
	}
	if q.CreationTime != "" {
		queryTime := strings.Split(q.CreationTime, "/")
		var startTime, endTime int64
		var err error
		if queryTime[0] == "" {
			startTime = 0
		} else {
			startTime, err = strconv.ParseInt(queryTime[0], 10, 64)
			if err != nil {
				return nil, 0, err
			}
		}
		if queryTime[1] == "" {
			endTime = time.Now().Unix()
		} else {
			endTime, err = strconv.ParseInt(queryTime[1], 10, 64)
			if err != nil {
				return nil, 0, err
			}
		}
		condition = condition.And("creation_time__gte", time.Unix(startTime, 0).UTC().Format(constants.TimeFormatLayout))
		condition = condition.And("creation_time__lte", time.Unix(endTime, 0).UTC().Format(constants.TimeFormatLayout))
	}
	condition = condition.AndNot("state__in", constants.BuildDeleted)
	total, err := qs.SetCond(condition).All(&list)
	if err != nil {
		return nil, 0, err
	}

	_, err = qs.SetCond(condition).OrderBy("-creation_time").Offset((offset) * limit).Limit(limit).All(&list)
	if err != nil {
		return nil, 0, err
	}
	return list, total, nil
}
