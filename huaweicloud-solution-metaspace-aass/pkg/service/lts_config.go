package service

import (
	"fmt"
	"strings"
	"time"

	"github.com/beego/beego/v2/adapter/toolbox"
	"github.com/google/uuid"
	"scase.io/application-auto-scaling-service/pkg/api/errors"
	"scase.io/application-auto-scaling-service/pkg/api/model"
	"scase.io/application-auto-scaling-service/pkg/cloudresource"
	"scase.io/application-auto-scaling-service/pkg/common"
	"scase.io/application-auto-scaling-service/pkg/db"

	"scase.io/application-auto-scaling-service/pkg/utils/logger"
)

const formatDateTime = "2006-01-02T15:04:05Z"

// 创建日志组
func CreateLtsLogGroup(tLogger *logger.FMLogger, projectId string,
	config *model.CreateLogGroup) (*model.CreateLogGroupResp, *errors.ErrorResp) {
	rc, err := cloudresource.GetLTSController(projectId)
	if err != nil {
		tLogger.Error("Get resource controller of project[%s] err: %+v", projectId, err)
		return nil, errors.NewErrorRespWithMessage(errors.LtsHostGroupError, err.Error())
	}
	logGroupId, err := rc.CreateLogGroup(tLogger, *config)
	if err != nil {
		tLogger.Error("Create LTS Log Group Error:%+v", err)
		return nil, errors.NewErrorRespWithMessage(errors.LtsHostGroupError, err.Error())
	}
	resp := model.CreateLogGroupResp{
		LogGroupName: config.LogGroupName,
		LogGroupId:   logGroupId,
		TTLInDay:     config.TTLInDay,
	}
	return &resp, nil
}

// 查询所有日志组
func ListLtsLogGroup(tLogger *logger.FMLogger, projectId string) (*model.ListLogGroups, *errors.ErrorResp) {
	rc, err := cloudresource.GetLTSController(projectId)
	if err != nil {
		tLogger.Error("Get resource controller of project[%s] err: %+v", projectId, err)
		return nil, errors.NewErrorRespWithMessage(errors.LtsHostGroupError, err.Error())
	}
	listResp, err := rc.ListLogGroups(tLogger)
	if err != nil {
		tLogger.Error("ListLtsLogGroup err:%+v ", err)
		return nil, errors.NewErrorRespWithMessage(errors.LtsHostGroupError, err.Error())
	}
	return listResp, nil
}

// 创建主机组
func CreateLTSHostGroup(tLogger *logger.FMLogger, projectId string,
	config *model.CreateAccessConfig) (string, error) {
	rc, err := cloudresource.GetLTSController(projectId)
	if err != nil {
		tLogger.Error("Get resource controller of project[%s] err: %+v", projectId, err)
		return "", err
	}
	req := model.CreateHostGroupReq{
		HostGroupName: config.LtsConfig.LtsHostGroupName,
	}
	hostGroupId, err := rc.CreateLTSHostGroup(tLogger, req)
	if err != nil {
		tLogger.Error("Create LTS Host Group Error:%+v", err)
		return "", err
	}
	return hostGroupId, nil
}

// 创建日志流
func CreateLTSLogStream(tLogger *logger.FMLogger, projectId string,
	config *model.CreateAccessConfig) (string, error) {
	rc, err := cloudresource.GetLTSController(projectId)
	if err != nil {
		tLogger.Error("Get resource controller of project[%s] err: %+v", projectId, err)
		return "", err
	}
	req := model.CreateLogStreamReq{
		LogStreamName:       config.LtsConfig.LtsLogStreamName,
		LogGroupId:          config.LtsConfig.LogGroupId,
		EnterpriseProjectId: config.EnterpriseProjectId,
	}
	logstream, err := rc.CreateLTSLogStream(tLogger, req)
	if err != nil {
		tLogger.Error("Create LTS log stream Error:%+v", err)
		return "", err
	}
	tLogger.Info(logstream)
	return logstream, nil
}

// 查询日志流
func ListLogStreams(tLogger *logger.FMLogger, projectId string, logGroupId string) (*model.ListLogStreams, error) {
	rc, err := cloudresource.GetLTSController(projectId)
	if err != nil {
		tLogger.Error("Get resource controller of project[%s] err: %+v", projectId, err)
		return nil, err
	}
	listLogStream, err := rc.ListLtsLogStream(tLogger, logGroupId)
	if err != nil {
		tLogger.Error("list log stream err:%s", err.Error())
		return nil, err
	}
	return listLogStream, nil
}

// 创建日志接入
func CreateLTSAccessConfig(tLogger *logger.FMLogger, projectId string, fleetId string,
	config *model.CreateAccessConfig) (*model.CreateAccessConfigResp, error) {

	rc, err := cloudresource.GetLTSController(projectId)
	if err != nil {
		tLogger.Error("Get resource controller of project[%s] err: %+v", projectId, err)
		return nil, err
	}
	if err := checkAccessConfigName(fleetId, config); err != nil {
		tLogger.Error("config name exist: %+v", err)
		return nil, err
	}
	// 新建主机组
	hostGroupId, err := CreateLTSHostGroup(tLogger, projectId, config)
	if err != nil {
		tLogger.Error("create host group err: %+v", err)
		return nil, err
	}
	// 新建日志流
	logStreamId, err := CreateLTSLogStream(tLogger, projectId, config)
	// 出错则回滚，把新建的主机组删除
	if err != nil {
		tLogger.Error("create log stream err: %+v", err)
		deleteHostReq := model.DeleteHostGroupReq{
			HostGroupIdList: []string{hostGroupId},
		}
		rc.DeleteLtsHostGroup(tLogger, deleteHostReq)
		return nil, err
	}
	// 整合参数
	createACreq := model.CreateAccessConfigReqToCloud{
		LogGroupId:          config.LtsConfig.LogGroupId,
		LogGroupName:        config.LtsConfig.LogGroupName,
		LogStreamId:         logStreamId,
		LogStreamName:       config.LtsConfig.LtsLogStreamName,
		HostGroupIDList:     []string{hostGroupId},
		HostGroupName:       config.LtsConfig.LtsHostGroupName,
		LogConfigPath:       config.LtsConfig.LtsLogPath,
		AccessConfigName:    config.LtsConfig.LtsAccessConfigName,
		EnterpriseProjectId: config.EnterpriseProjectId,
	}
	ltsAccessConfigResp, err := rc.CreateAccessConfig(tLogger, createACreq)
	if err != nil {
		if err := rollBackLogStreamAndHostGroup(rc, tLogger, createACreq); err != nil {
			return nil, err
		}
		return nil, err
	}
	ltsConfig, err := insertConfig(&createACreq, ltsAccessConfigResp, projectId, fleetId)
	if err != nil {
		tLogger.Error("Insert LTS Access Config to DB Error:%+v", err)
		return nil, err
	}
	createResp := model.BuildCreateAccessConfigResp(*ltsConfig)
	tLogger.Info("lts access config create success :%s", ltsAccessConfigResp.AccessConfigId)
	return &createResp, nil
}

// 写入数据库
func insertConfig(createACreq *model.CreateAccessConfigReqToCloud,
	createACResp *model.CreateAccessConfigResp, projectId string, fleetId string) (*db.LtsConfig, error) {
	u, _ := uuid.NewUUID()
	id := u.String()
	ltsConfig := db.LtsConfig{
		Id:                  id,
		FleetId:             fleetId,
		ProjectId:           projectId,
		LogGroupId:          createACreq.LogGroupId,
		LogGroupName:        createACResp.LogGroupName,
		LogStreamId:         createACreq.LogStreamId,
		LogStreamName:       createACreq.LogStreamName,
		HostGroupID:         strings.Join(createACreq.HostGroupIDList, ","),
		HostGroupName:       createACreq.HostGroupName,
		LogConfigPath:       strings.Join(createACreq.LogConfigPath, ","),
		AccessConfigName:    createACResp.AccessConfigName,
		AccessConfigId:      createACResp.AccessConfigId,
		Description:         createACreq.Description,
		EnterpriseProjectId: createACreq.EnterpriseProjectId,
	}
	if err := db.LtsConfigTable().Insert(&ltsConfig); err != nil {
		return nil, err
	}
	return &ltsConfig, nil
}
func checkAccessConfigName(fleetId string, config *model.CreateAccessConfig) error {
	if config.LtsConfig.LtsHostGroupName == "" {
		config.LtsConfig.LtsHostGroupName = "Fleet_" + fleetId
	}
	if config.LtsConfig.LtsLogStreamName == "" {
		config.LtsConfig.LtsLogStreamName = "Fleet_" + fleetId
	}
	if config.LtsConfig.LtsAccessConfigName == "" {
		config.LtsConfig.LtsAccessConfigName = "Fleet_" + fleetId
	}
	_, err := db.LtsConfigTable().CheckExist(config.FleetId, config.LtsConfig.LtsLogStreamName,
		config.LtsConfig.LtsHostGroupName, config.LtsConfig.LtsAccessConfigName)
	if err != nil {
		return err
	}
	return nil
}

func rollBackLogStreamAndHostGroup(rc *cloudresource.LtsResourceController, tLogger *logger.FMLogger,
	createACreq model.CreateAccessConfigReqToCloud) error {
	deleteHostReq := model.DeleteHostGroupReq{
		HostGroupIdList: createACreq.HostGroupIDList,
	}
	if err := rc.DeleteLtsHostGroup(tLogger, deleteHostReq); err != nil {
		return err
	}
	deleteLogStreamReq := model.DeleteLogStreamReq{
		LogGroupId:  createACreq.LogGroupId,
		LogStreamId: createACreq.LogStreamId,
	}
	if err := rc.DeleteLtsLogStream(tLogger, deleteLogStreamReq); err != nil {
		return err
	}
	return nil
}

// 删除日志接入
func DeleteLTSAccessConfig(tLogger *logger.FMLogger, projectId string, accessconfigId string) error {
	rc, err := cloudresource.GetLTSController(projectId)
	if err != nil {
		tLogger.Error("Get resource controller of project[%s] err: %+v", projectId, err)
		return err
	}
	ltsConfig, err := db.LtsConfigTable().Get(db.Filters{"AccessConfigId": accessconfigId})
	if err != nil {
		tLogger.Error("Get ltsConfig from db err : %s", err.Error())
		return err
	}
	// 删除主机组
	deleteHostGroupReq := model.DeleteHostGroupReq{
		HostGroupIdList: []string{ltsConfig.HostGroupID},
	}
	err = rc.DeleteLtsHostGroup(tLogger, deleteHostGroupReq)
	if err != nil {
		tLogger.Error("Delete host group err : %s", err.Error())
		return err
	}
	// 删除日志流，日志接入自动删除
	deleteLogStreamReq := model.DeleteLogStreamReq{
		LogGroupId:  ltsConfig.LogGroupId,
		LogStreamId: ltsConfig.LogStreamId,
	}
	err = rc.DeleteLtsLogStream(tLogger, deleteLogStreamReq)
	if err != nil {
		tLogger.Error("Delete log stream err : %s", err.Error())
		return err
	}
	// 删除日志接入
	err = rc.DeleteAccessConfig(tLogger, accessconfigId)
	if err != nil {
		tLogger.Error("Delete log access config err : %s", err.Error())
		return err
	}
	// 删除数据库
	err = db.LtsConfigTable().Delete(&ltsConfig)
	if err != nil {
		tLogger.Error("Delete ltsConfig from db err : %s", err.Error())
		return err
	}
	tLogger.Info("Delete lts access config %s success", ltsConfig.AccessConfigName)
	return nil
}

// list 日志接入
func ListLtsAccessConfig(tLogger *logger.FMLogger, projectId string) (*model.ListAccessConfig, *errors.ErrorResp) {
	rc, err := cloudresource.GetLTSController(projectId)
	if err != nil {
		tLogger.Error("Get resource controller of project[%s] err: %+v", projectId, err)
		return nil, errors.NewErrorRespWithMessage(errors.ServerInternalError, err.Error())
	}
	accessConfigList, err := rc.ListAccessConfig(tLogger)
	if err != nil {
		tLogger.Error("list access config err : %s", err.Error())
		return nil, errors.NewErrorRespWithMessage(errors.LtsAccessConfigError, err.Error())
	}
	return accessConfigList, nil
}

func QueryAccessConfig(tLogger *logger.FMLogger, access_config_id string) (*model.CreateAccessConfigResp, *errors.ErrorResp) {
	object, err := db.LtsConfigTable().Get(db.Filters{"AccessConfigId": access_config_id})
	if err != nil {
		return nil, errors.NewErrorRespWithMessage(errors.LtsAccessConfigError, err.Error())
	}
	respMsg := model.BuildCreateAccessConfigResp(object)
	return &respMsg, nil
}

// 从数据库list 日志接入
func ListLtsAccessConfigFromDB(tLogger *logger.FMLogger, limit int, offset int, projectId string) (*model.ListAccessConfig, *errors.ErrorResp) {
	acList, err := db.LtsConfigTable().ListbyProject(offset, limit, projectId)
	if err != nil {
		tLogger.Error("ListLtsAccessConfigFromDB err:%+v", err)
		return nil, errors.NewErrorRespWithMessage(errors.ServerInternalError, err.Error())
	}
	resp := model.ListAccessConfig{}
	for _, ac := range acList {
		resp.AccessConfigList = append(resp.AccessConfigList, model.AccessConfig{
			Id:               ac.Id,
			AccessConfigName: ac.AccessConfigName,
			AccessConfigId:   ac.AccessConfigId,
			HostGroupIDList:  []string{ac.HostGroupID},
			HostGroupName:    ac.HostGroupName,
			LogGroupId:       ac.LogGroupId,
			LogGroupName:     ac.LogGroupName,
			LogStreamName:    ac.LogStreamName,
			LogStreamId:      ac.LogStreamId,
			LogConfigPath:    []string{ac.LogConfigPath},
			ObsTransferPath:  ac.ObsTransferPath,
			CreateTime:       ac.CreateTime.Format(formatDateTime),
			Description:      ac.Description,
		})
	}
	resp.Total, err = db.LtsConfigTable().Total(projectId)
	if err != nil {
		tLogger.Error("list access config err : %s", err.Error())
	}
	resp.Count = len(acList)
	return &resp, nil
}

// 更新数据库日志接入描述
func UpdateAccessConfigToDB(tLogger *logger.FMLogger, projectId string, config model.UpdateAccessConfigToDB) (*model.AccessConfig, *errors.ErrorResp) {
	ltsconfig, err := db.LtsConfigTable().Get(db.Filters{"AccessConfigId": config.AccessConfigId})
	if err != nil {
		tLogger.Error("UpdateAccessConfigToDB err:%+v", err)
		return nil, errors.NewErrorRespWithMessage(errors.ServerInternalError, err.Error())
	}
	ltsconfig.Description = config.Description
	if err := db.LtsConfigTable().Update(&ltsconfig, "Description"); err != nil {
		tLogger.Error("UpdateAccessConfigToDB err:%+v", err)
		return nil, errors.NewErrorRespWithMessage(errors.ServerInternalError, err.Error())
	}
	resp := model.AccessConfig{
		AccessConfigName: ltsconfig.AccessConfigName,
		AccessConfigId:   ltsconfig.AccessConfigId,
		HostGroupIDList:  []string{ltsconfig.HostGroupID},
		LogGroupId:       ltsconfig.LogGroupId,
		LogStreamName:    ltsconfig.LogStreamName,
		LogStreamId:      ltsconfig.LogStreamId,
		LogConfigPath:    []string{ltsconfig.LogConfigPath},
		CreateTime:       ltsconfig.CreateTime.Format(formatDateTime),
		Description:      ltsconfig.Description,
	}
	return &resp, nil
}

// 创建日志转储
func CreateLtsTransfer(tLogger *logger.FMLogger, projectId string, config model.LogTransferReq) (*model.CreateTransferResp, *errors.ErrorResp) {
	rc, err := cloudresource.GetLTSController(projectId)
	if err != nil {
		tLogger.Error("Get resource controller of project[%s] err: %+v", projectId, err)
		return nil, errors.NewErrorRespWithMessage(errors.ServerInternalError, err.Error())
	}
	ltsconfig, err := db.LtsConfigTable().Get(db.Filters{"LogStreamId": config.LogStreamId})
	if err != nil {
		return nil, errors.NewErrorRespWithMessage(errors.LtsLogTransferError, err.Error())
	}
	config.EnterpriseProjectId = ltsconfig.EnterpriseProjectId
	resp, err := rc.LogTransfer(tLogger, &config)
	if err != nil {
		tLogger.Error("LtsTransfer err : %s", err.Error())
		return nil, errors.NewErrorRespWithMessage(errors.LtsLogTransferError, err.Error())
	}
	// 更新数据库字段

	ltsconfig.ObsTransferPath = resp.TransferDetail.ObsTransferPath
	if errUd := db.LtsConfigTable().Update(&ltsconfig, "ObsTransferPath"); errUd != nil {
		return nil, errors.NewErrorRespWithMessage(errors.LtsLogTransferError, errUd.Error())
	}

	if errIt := insertTransferToDB(resp, projectId); errIt != nil {
		tLogger.Error("insert to DB err : %s", errIt.Error())
		return nil, errors.NewErrorRespWithMessage(errors.LtsLogTransferError, errIt.Error())
	}
	return resp, nil
}

// list 日志转储
func ListLtsTransferFromDB(tLogger *logger.FMLogger, limit int, offset int, projectId string,
	logStreamId string) (*model.ListTransfersInfo, *errors.ErrorResp) {
	transferList, err := db.TransferTable().ListbyProjectAndLogStreamId(offset, limit, projectId, logStreamId)
	if err != nil {
		tLogger.Error("ListLtsAccessConfigFromDB err:%+v", err)
		return nil, errors.NewErrorRespWithMessage(errors.ServerInternalError, err.Error())
	}
	resp := model.ListTransfersInfo{}
	for _, transfer := range transferList {
		resp.ListTransfers = append(resp.ListTransfers, model.ListTransferResp{
			LogGroupId:    transfer.LogGroupId,
			LogStreamId:   transfer.LogStreamId,
			LogTransferId: transfer.LogTransferId,
			TransferDetail: model.TransferInfo{
				ObsPeriodUnit:   transfer.ObsPeriodUnit,
				ObsPeriod:       int32(transfer.ObsPeriod),
				ObsBucketName:   transfer.ObsBucketName,
				ObsTransferPath: transfer.ObsTransferPath,
			},
		})
	}
	resp.Total, err = db.TransferTable().Total(projectId)
	if err != nil {
		tLogger.Error("list access config err : %s", err.Error())
	}
	resp.Count = len(transferList)
	return &resp, nil
}

func insertTransferToDB(config *model.CreateTransferResp, projectId string) error {
	u, _ := uuid.NewUUID()
	id := u.String()
	object := db.LogTransfer{
		Id:              id,
		ProjectId:       projectId,
		LogGroupId:      config.LogGroupId,
		LogStreamId:     config.LogStreamId,
		LogTransferId:   config.LogTransferId,
		ObsPeriodUnit:   config.TransferDetail.ObsPeriodUnit,
		ObsPeriod:       int(config.TransferDetail.ObsPeriod),
		ObsTransferPath: config.TransferDetail.ObsTransferPath,
		ObsBucketName:   config.TransferDetail.ObsBucketName,
		CreateTime:      time.Now(),
	}
	if err := db.TransferTable().Insert(&object); err != nil {
		return err
	}
	return nil
}

// 删除日志转储
func DeleteTransfer(tLogger *logger.FMLogger, projectId string, transferId string) *errors.ErrorResp {
	rc, err := cloudresource.GetLTSController(projectId)
	if err != nil {
		tLogger.Error("Get resource controller of project[%s] err: %+v", projectId, err)
		return errors.NewErrorRespWithMessage(errors.ServerInternalError, err.Error())
	}
	if err := rc.DeleteTransfer(tLogger, transferId); err != nil {
		tLogger.Error("Delete log transfer err : %s", err.Error())
		return errors.NewErrorRespWithMessage(errors.LtsLogTransferError, err.Error())
	}
	transferInfo, err := db.TransferTable().Get(db.Filters{"LogTransferId": transferId})
	if err != nil {
		tLogger.Error("Get TransferTable from DB err:%s", err.Error())
		return errors.NewErrorRespWithMessage(errors.LtsLogTransferError, err.Error())
	}
	ltsConfig, err := db.LtsConfigTable().Get(db.Filters{"LogStreamId": transferInfo.LogStreamId})
	if err != nil {
		tLogger.Error("Get ltsConfigTable from DB err:%s", err.Error())
		return errors.NewErrorRespWithMessage(errors.LtsLogTransferError, err.Error())
	}
	ltsConfig.ObsTransferPath = ""
	if err := db.LtsConfigTable().Update(&ltsConfig, "ObsTransferPath"); err != nil {
		tLogger.Error("Update ltsConfigTable from DB err:%s", err.Error())
		return errors.NewErrorRespWithMessage(errors.LtsLogTransferError, err.Error())
	}
	if err := db.TransferTable().DeleteByTransferId(transferId); err != nil {
		tLogger.Error("Delete log transfer from db err : %s", err.Error())
		return errors.NewErrorRespWithMessage(errors.LtsLogTransferError, err.Error())
	}
	return nil
}

// 更新日志转储
func UpdateTransfer(tLogger *logger.FMLogger, projectId string, config model.UpdateTransfer) (*model.CreateTransferResp, *errors.ErrorResp) {
	rc, err := cloudresource.GetLTSController(projectId)
	if err != nil {
		tLogger.Error("Get resource controller of project[%s] err: %+v", projectId, err)
		return nil, errors.NewErrorRespWithMessage(errors.ServerInternalError, err.Error())
	}
	updateResp, err := rc.UpdateTransfer(tLogger, &config)
	if err != nil {
		tLogger.Error("update log transfer err : %s", err.Error())
		return nil, errors.NewErrorRespWithMessage(errors.LtsLogTransferError, err.Error())
	}
	if err := updateTransferToDB(updateResp, updateResp.LogTransferId); err != nil {
		tLogger.Error("update log transfer to DB err : %s", err.Error())
		return nil, errors.NewErrorRespWithMessage(errors.LtsLogTransferError, err.Error())
	}
	return updateResp, nil
}

func updateTransferToDB(config *model.CreateTransferResp, transferId string) error {
	object, err := db.TransferTable().Get(db.Filters{"TransferId": transferId})
	if err != nil {
		return err
	}
	object.ObsBucketName = config.TransferDetail.ObsBucketName
	object.ObsPeriod = int(config.TransferDetail.ObsPeriod)
	object.ObsPeriodUnit = config.TransferDetail.ObsPeriodUnit
	object.ObsTransferPath = config.TransferDetail.ObsTransferPath
	if err := db.TransferTable().Update(&object); err != nil {
		return err
	}
	return nil
}

func QueryTransfer(projectId string, transferId string) (*model.ListTransferResp, *errors.ErrorResp) {
	object, err := db.TransferTable().Get(db.Filters{"LogTransferId": transferId})
	if err != nil {
		return nil, errors.NewErrorRespWithMessage(errors.ServerInternalError, err.Error())
	}
	respMsg := model.BuildTransferResp(object)
	return &respMsg, nil
}

func InitTask() {
	tk := toolbox.NewTask("UpdateHostGroup", "0 */10 * * * *", UpdateHostGroup)
	toolbox.AddTask("UpdateHostGroup", tk)
	toolbox.StartTask()
	defer toolbox.StartTask()
}

// 更新主机组
func UpdateHostGroup() error {
	tLogger := logger.A.WithField("scheduled task", "update host group")
	ltsList, err := db.LtsConfigTable().List()
	var errList []string
	if err != nil {
		tLogger.Error("List LTS Access Config from DB Error:%+v", err)
		return err
	}
	for _, ltsconfig := range ltsList {
		instanceList, err := GetInstanceList(ltsconfig, tLogger)
		if err != nil {
			errList = append(errList, err.ErrMsg)
			tLogger.Error("Get instance list err:%s", err)
			continue
		}
		tLogger.Info("update host group %s success", instanceList)
	}
	if len(errList) > 0 {
		return fmt.Errorf("update host group error:%s", strings.Join(errList, ","))
	}
	return nil
}

// 过滤所有instance id， 并插入主机组
func GetInstanceList(ltsconfig db.LtsConfig, tLogger *logger.FMLogger) ([]string, *errors.ErrorResp) {
	rc, err := cloudresource.GetLTSController(ltsconfig.ProjectId)
	if err != nil {
		tLogger.Error("Get resource controller of project[%s] err: %+v", ltsconfig.ProjectId, err)
		return nil, errors.NewErrorRespWithMessage(errors.ServerInternalError, err.Error())
	}
	if ltsconfig.ASGroupId == "" {
		if err := getASGroupIdFunc(tLogger, &ltsconfig); err != nil {
			return nil, errors.NewErrorRespWithMessage(errors.ServerInternalError, err.ErrMsg)
		}
		db.LtsConfigTable().Update(&ltsconfig, "as_group_id")
	}
	// 根据as group id获得所有实例信息
	req := &model.QueryInstanceParams{
		ScalingGroupId: ltsconfig.ASGroupId,
		Limit:          common.DefaultLimit,
		StartNumber:    common.DefaultStartNumber,
		HealthState:    "",
		LifeCycleState: "",
		ProjectId:      ltsconfig.ProjectId,
	}
	resp, e := GenerateInstancesFromAs(tLogger, req)
	if e != nil {
		tLogger.Error("list scaling instance error: %v", err)
		return nil, errors.NewErrorRespWithMessage(errors.ServerInternalError, e.ErrMsg)
	}
	var instanceList []string
	for _, instance := range resp.Instances {
		instanceList = append(instanceList, instance.InstanceId)
	}
	// 过滤出状态为running的
	HostIdList, err := rc.HostIdListFilter(tLogger, instanceList)
	if err != nil {
		tLogger.Error("filter host id list err:%s", err.Error())
		return nil, errors.NewErrorRespWithMessage(errors.ServerInternalError, err.Error())
	}
	// 插入主机组
	InsertHostGroupreq := model.UpdateHostGroupReq{
		HostGroupId: ltsconfig.HostGroupID,
		HostIdList:  HostIdList,
	}
	err = rc.InsertLtsHostGroup(tLogger, InsertHostGroupreq)
	if err != nil {
		tLogger.Error("Insert to Host group err:%s", err.Error())
		return nil, errors.NewErrorRespWithMessage(errors.ServerInternalError, err.Error())
	}
	return HostIdList, nil
}

func getASGroupIdFunc(tLogger *logger.FMLogger, ltsconfig *db.LtsConfig) *errors.ErrorResp {
	// 获取fleet下的所有instance conf id
	ins_conf_id, err := GetInsConfIdByFleetId(ltsconfig.FleetId, tLogger)
	if err != nil {
		tLogger.Error("get instance configuration id error: %v", err)
		return errors.NewErrorRespWithMessage(errors.ServerInternalError, err.ErrMsg)
	}
	// 通过ins conf id 查询伸缩组id
	ltsconfig.ASGroupId, err = GetScalingGroupByInsConfId(ins_conf_id, tLogger)
	if err != nil {
		tLogger.Error("get scaling group id error: %v", err)
		return errors.NewErrorRespWithMessage(errors.ServerInternalError, err.ErrMsg)
	}
	return nil
}
