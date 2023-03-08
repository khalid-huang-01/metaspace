package cloudresource

import (
	"fmt"
	"strings"
	"time"

	"github.com/huaweicloud/huaweicloud-sdk-go-v3/core/auth/basic"
	lts "github.com/huaweicloud/huaweicloud-sdk-go-v3/services/lts/v2"
	ltsmodel "github.com/huaweicloud/huaweicloud-sdk-go-v3/services/lts/v2/model"
	"scase.io/application-auto-scaling-service/pkg/api/model"
	"scase.io/application-auto-scaling-service/pkg/db"
	"scase.io/application-auto-scaling-service/pkg/setting"
	"scase.io/application-auto-scaling-service/pkg/utils/logger"
)

type LtsResourceController struct {
	ltsClient *lts.LtsClient

	expireAt  time.Time
	projectId string
}

// 创建日志组
func (c *LtsResourceController) CreateLogGroup(tLogger *logger.FMLogger, param model.CreateLogGroup) (string, error) {
	request := &ltsmodel.CreateLogGroupRequest{}
	request.Body = &ltsmodel.CreateLogGroupParams{
		TtlInDays:    int32(param.TTLInDay),
		LogGroupName: param.LogGroupName,
	}
	response, err := c.ltsClient.CreateLogGroup(request)
	if err != nil {
		tLogger.Error("CreateLogGroup err %s ,request body: %+v", err.Error(), *request.Body)
		return "", err
	}

	return *response.LogGroupId, nil
}

// List 所有日志组
func (c *LtsResourceController) ListLogGroups(tLogger *logger.FMLogger) (*model.ListLogGroups, error) {
	request := &ltsmodel.ListLogGroupsRequest{}
	response, err := c.ltsClient.ListLogGroups(request)
	if err != nil {
		tLogger.Error("ListLogGroups err %s ,request body: %+v", err.Error(), *request)
		return nil, err
	}
	listResp := model.ListLogGroups{}
	var total int
	for _, resp := range *response.LogGroups {
		listResp.LogGroups = append(listResp.LogGroups, model.LogGroups{
			CreationTime: time.UnixMilli(resp.CreationTime).Format("2006-01-02T15:04:05Z"),
			LogGroupName: resp.LogGroupName,
			LogGroupID:   resp.LogGroupId,
			TTLInDays:    int(resp.TtlInDays),
		})
		total++
	}
	listResp.Total = total
	return &listResp, nil
}

// 创建主机组操作
func (c *LtsResourceController) CreateLTSHostGroup(tLogger *logger.FMLogger, params model.CreateHostGroupReq) (string, error) {
	createLtsHostGroupReq := &ltsmodel.CreateHostGroupRequest{
		Body: &ltsmodel.CreateHostGroupRequestBody{
			HostGroupName: params.HostGroupName,
			HostGroupType: ltsmodel.GetCreateHostGroupRequestBodyHostGroupTypeEnum().LINUX,
		},
	}
	createResp, err := c.ltsClient.CreateHostGroup(createLtsHostGroupReq)
	if err != nil {
		tLogger.Error("CreateLtsHostGroup err %s ,request body: %+v", err.Error(), *createLtsHostGroupReq.Body)
		return "", err
	}
	return *createResp.HostGroupId, nil
}

// 创建日志流操作
func (c *LtsResourceController) CreateLTSLogStream(tLogger *logger.FMLogger, params model.CreateLogStreamReq) (string, error) {
	createLtsLogStreamReq := &ltsmodel.CreateLogStreamRequest{}
	createLtsLogStreamReq.LogGroupId = params.LogGroupId
	createLtsLogStreamReq.Body = &ltsmodel.CreateLogStreamParams{
		LogStreamName: params.LogStreamName,
	}
	createResp, err := c.ltsClient.CreateLogStream(createLtsLogStreamReq)
	if err != nil {
		tLogger.Error("CreateLogStreamReq err %s ,request body: %+v", err.Error(), *createLtsLogStreamReq.Body)
		return "", err
	}
	return *createResp.LogStreamId, nil
}

// list日志流
func (c *LtsResourceController) ListLtsLogStream(tLogger *logger.FMLogger, logGroupId string) (*model.ListLogStreams, error) {
	request := &ltsmodel.ListLogStreamRequest{}
	request.LogGroupId = logGroupId
	response, err := c.ltsClient.ListLogStream(request)
	if err != nil {
		tLogger.Error("ListLtsLogSteam err %s ,request body: %+v", err.Error(), &request.LogGroupId)
		return nil, err
	}
	var total int
	listLogStream := model.ListLogStreams{}
	for _, resp := range *response.LogStreams {
		listLogStream.LogStreams = append(listLogStream.LogStreams, model.LogStreams{
			LogStreamName: resp.LogStreamName,
			LogStreamID:   resp.LogStreamId,
		})
		total++
	}
	listLogStream.Total = total
	tLogger.Info("list log stream success")
	return &listLogStream, nil
}

// 配置日志接入
func (c *LtsResourceController) CreateAccessConfig(tLogger *logger.FMLogger,
	params model.CreateAccessConfigReqToCloud) (*model.CreateAccessConfigResp, error) {

	request := &ltsmodel.CreateAccessConfigRequest{}
	var listHostGroupIdListHostGroupInfo = params.HostGroupIDList
	hostGroupInfobody := &ltsmodel.AccessConfigHostGroupIdListCreate{
		HostGroupIdList: listHostGroupIdListHostGroupInfo,
	}
	logInfobody := &ltsmodel.AccessConfigBaseLogInfoCreate{
		LogGroupId:  params.LogGroupId,
		LogStreamId: params.LogStreamId,
	}
	modeSingle := ltsmodel.GetAccessConfigFormatSingleCreateModeEnum().WILDCARD
	valueSingle := "YYYY-MM-DD hh:mm:ss"
	singleFormat := &ltsmodel.AccessConfigFormatSingleCreate{
		Mode:  &modeSingle,
		Value: &valueSingle,
	}
	formatAccessConfigDetail := &ltsmodel.AccessConfigFormatCreate{
		Single: singleFormat,
	}
	var listPathsAccessConfigDetail = params.LogConfigPath
	accessConfigDetailbody := &ltsmodel.AccessConfigDeatilCreate{
		Paths:  listPathsAccessConfigDetail,
		Format: formatAccessConfigDetail,
	}
	request.Body = &ltsmodel.CreateAccessConfigRequestBody{
		HostGroupInfo:      hostGroupInfobody,
		LogInfo:            logInfobody,
		AccessConfigDetail: accessConfigDetailbody,
		AccessConfigType:   ltsmodel.GetCreateAccessConfigRequestBodyAccessConfigTypeEnum().AGENT,
		AccessConfigName:   params.AccessConfigName,
	}
	createResp, err := c.ltsClient.CreateAccessConfig(request)
	if err != nil {
		tLogger.Error("CreateLtsLogStream request body: %+v", *request.Body)
		return nil, err
	}
	resp := model.CreateAccessConfigResp{
		AccessConfigName: *createResp.AccessConfigName,
		AccessConfigId:   *createResp.AccessConfigId,
		LogGroupName:     *createResp.LogInfo.LogGroupName,
	}

	return &resp, nil
}

// list 日志接入
func (c *LtsResourceController) ListAccessConfig(tLogger *logger.FMLogger) (*model.ListAccessConfig, error) {
	request := &ltsmodel.ListAccessConfigRequest{}
	request.Body = &ltsmodel.GetAccessConfigListRequestBody{
		LogStreamNameList:    nil,
		LogGroupNameList:     nil,
		HostGroupNameList:    nil,
		AccessConfigNameList: nil,
	}
	response, err := c.ltsClient.ListAccessConfig(request)
	if err != nil {
		tLogger.Error("ListAccessConfig err %s ,request body: %+v", err.Error(), &request)
		return nil, err
	}
	accessConfigList := *&model.ListAccessConfig{}
	for _, resp := range *response.Result {
		if resp.HostGroupInfo == nil {
			resp.HostGroupInfo = &ltsmodel.AccessConfigHostGroupIdList{}
		}
		accessConfigList.AccessConfigList = append(accessConfigList.AccessConfigList,
			model.AccessConfig{
				AccessConfigName: *resp.AccessConfigName,
				AccessConfigId:   *resp.AccessConfigId,
				HostGroupIDList:  resp.HostGroupInfo.HostGroupIdList,
				LogGroupId:       *resp.LogInfo.LogGroupId,
				LogStreamName:    *resp.LogInfo.LogStreamName,
				LogStreamId:      *resp.LogInfo.LogStreamId,
				LogConfigPath:    *resp.AccessConfigDetail.Paths,
				CreateTime:       string(*resp.CreateTime),
			})
	}
	accessConfigList.Total = len(accessConfigList.AccessConfigList)
	return &accessConfigList, nil
}

// 删除日志接入
func (c *LtsResourceController) DeleteAccessConfig(tLogger *logger.FMLogger, param string) error {
	request := &ltsmodel.DeleteAccessConfigRequest{}
	var listAccessConfigIdListbody = []string{
		param,
	}
	request.Body = &ltsmodel.DeleteAccessConfigRequestBody{
		AccessConfigIdList: listAccessConfigIdListbody,
	}
	_, err := c.ltsClient.DeleteAccessConfig(request)
	if err != nil {
		tLogger.Error("delete access config err:%+v", err)
		return err
	}
	return nil
}

// 插入主机组
func (c *LtsResourceController) InsertLtsHostGroup(tLogger *logger.FMLogger,
	params model.UpdateHostGroupReq) error {

	updateLtsHostGroupReq := &ltsmodel.UpdateHostGroupRequest{
		Body: &ltsmodel.UpdateHostGroupRequestBody{
			HostGroupId: params.HostGroupId,
			HostIdList:  &params.HostIdList,
		},
	}
	createResp, err := c.ltsClient.UpdateHostGroup(updateLtsHostGroupReq)
	if err != nil {
		tLogger.Error("CreateLtsHostgGroup err %s ,request body: %+v", err.Error(), *updateLtsHostGroupReq.Body)
		return err
	}
	tLogger.Info("update host group %s success", *createResp.HostGroupId)
	return nil
}

// list所有ICAgent正常运行的主机
func (c *LtsResourceController) HostIdListFilter(tLogger *logger.FMLogger, instanceList []string) ([]string, error) {
	hostStatusFilter := ltsmodel.GetGetHostListFilterHostStatusEnum().RUNNING
	filterbody := &ltsmodel.GetHostListFilter{
		HostStatus: &hostStatusFilter,
	}
	request := &ltsmodel.ListHostRequest{
		Body: &ltsmodel.GetHostListRequestBody{
			Filter:     filterbody,
			HostIdList: instanceList,
		},
	}
	filterResp, err := c.ltsClient.ListHost(request)
	if err != nil {
		tLogger.Error("CreateLtsHostgGroup err %s ,request body: %+v", err.Error(), *request.Body)
		return nil, err
	}
	instanceIdList := []string{}
	for _, instance := range *filterResp.Result {
		instanceIdList = append(instanceIdList, *instance.HostId)
	}
	tLogger.Info("CreateLtsHostgGroup success")
	return instanceIdList, nil
}

// 删除日志流
func (c *LtsResourceController) DeleteLtsLogStream(tLogger *logger.FMLogger, param model.DeleteLogStreamReq) error {
	request := &ltsmodel.DeleteLogStreamRequest{}
	request.LogGroupId = param.LogGroupId
	request.LogStreamId = param.LogStreamId
	_, err := c.ltsClient.DeleteLogStream(request)
	if err != nil {
		tLogger.Error("DeleteLtsLogStream err %s ,request body: %+v", err.Error(), *&request.LogStreamId)
		return err
	}
	tLogger.Info("Delete lts log stream %s success", request.LogStreamId)
	return nil
}

// 删除主机组
func (c *LtsResourceController) DeleteLtsHostGroup(tLogger *logger.FMLogger, param model.DeleteHostGroupReq) error {
	request := &ltsmodel.DeleteHostGroupRequest{}
	var listHostGroupIdListbody = param.HostGroupIdList
	request.Body = &ltsmodel.DeleteHostGroupRequestBody{
		HostGroupIdList: listHostGroupIdListbody,
	}
	_, err := c.ltsClient.DeleteHostGroup(request)
	if err != nil {
		tLogger.Error("DeleteLtsHostGroup err %s ,request body: %+v", err.Error(), *request.Body)
		return err
	}
	tLogger.Info("Delete lts log group %s success", request.Body.HostGroupIdList)
	return nil
}

// 日志转储
func (c *LtsResourceController) LogTransfer(tLogger *logger.FMLogger, param *model.LogTransferReq) (*model.CreateTransferResp, error) {
	request := &ltsmodel.CreateTransferRequest{}
	logTransferDetailLogTransferInfo := &ltsmodel.TransferDetail{
		ObsPeriod:        convertPeriodEnum(param.TransferInfo.ObsPeriod),
		ObsPeriodUnit:    getPeriodUnit(param.TransferInfo.ObsPeriod),
		ObsBucketName:    param.TransferInfo.ObsBucketName,
		ObsEpsId:         &param.EnterpriseProjectId,
		ObsDirPreFixName: &param.TransferInfo.ObsTransferPath,
	}
	logTransferInfobody := &ltsmodel.CreateTransferRequestBodyLogTransferInfo{
		LogTransferType:   "OBS",
		LogTransferMode:   ltsmodel.GetCreateTransferRequestBodyLogTransferInfoLogTransferModeEnum().CYCLE,
		LogStorageFormat:  ltsmodel.GetCreateTransferRequestBodyLogTransferInfoLogStorageFormatEnum().RAW,
		LogTransferStatus: ltsmodel.GetCreateTransferRequestBodyLogTransferInfoLogTransferStatusEnum().ENABLE,
		LogTransferDetail: logTransferDetailLogTransferInfo,
	}
	var listLogStreamsbody = []ltsmodel.CreateTransferRequestBodyLogStreams{
		{
			LogStreamId: param.LogStreamId,
		},
	}
	request.Body = &ltsmodel.CreateTransferRequestBody{
		LogTransferInfo: logTransferInfobody,
		LogStreams:      listLogStreamsbody,
		LogGroupId:      param.LogGroupId,
	}
	resp, err := c.ltsClient.CreateTransfer(request)
	if err != nil {
		tLogger.Error("LogTransfer err %s ,request body: %+v", err.Error(), *request.Body)
		return nil, err
	}
	createResp := convertTransferResp(resp)
	return &createResp, nil
}

// 删除日志转储
func (c *LtsResourceController) DeleteTransfer(tLogger *logger.FMLogger, transferId string) error {
	request := &ltsmodel.DeleteTransferRequest{}
	request.LogTransferId = transferId
	_, err := c.ltsClient.DeleteTransfer(request)
	if err != nil {
		tLogger.Error("LogTransfer err %s ,request body: %+v", err.Error(), request)
		return err
	}
	return nil
}

// 修改日志转储
func (c *LtsResourceController) UpdateTransfer(tLogger *logger.FMLogger, param *model.UpdateTransfer) (*model.CreateTransferResp, error) {
	request := &ltsmodel.UpdateTransferRequest{}
	logTransferDetailLogTransferInfo := &ltsmodel.TransferDetail{
		ObsPeriod:       convertPeriodEnum(param.TransferInfo.ObsPeriod),
		ObsPeriodUnit:   getPeriodUnit(param.TransferInfo.ObsPeriod),
		ObsBucketName:   param.TransferInfo.ObsBucketName,
		ObsTransferPath: &param.TransferInfo.ObsTransferPath,
	}
	logTransferInfobody := &ltsmodel.UpdateTransferRequestBodyLogTransferInfo{
		LogStorageFormat:  ltsmodel.GetUpdateTransferRequestBodyLogTransferInfoLogStorageFormatEnum().RAW,
		LogTransferStatus: ltsmodel.GetUpdateTransferRequestBodyLogTransferInfoLogTransferStatusEnum().ENABLE,
		LogTransferDetail: logTransferDetailLogTransferInfo,
	}
	request.Body = &ltsmodel.UpdateTransferRequestBody{
		LogTransferInfo: logTransferInfobody,
		LogTransferId:   param.TransferId,
	}
	response, err := c.ltsClient.UpdateTransfer(request)
	if err != nil {
		tLogger.Error("UpdateTransfer err %s ,request body: %+v", err.Error(), request)
		return nil, err
	}
	resp := convertTransferResp((*ltsmodel.CreateTransferResponse)(response))
	return &resp, nil
}

func (c *LtsResourceController) ListTransfer(tLogger *logger.FMLogger) error {
	request := &ltsmodel.ListTransfersRequest{}
	response, err := c.ltsClient.ListTransfers(request)
	if err != nil {
		return err
	}
	fmt.Println(response)
	return nil
}

func convertPeriodEnum(num int32) ltsmodel.TransferDetailObsPeriod {
	switch num {
	case 1:
		return ltsmodel.GetTransferDetailObsPeriodEnum().E_1
	case 2:
		return ltsmodel.GetTransferDetailObsPeriodEnum().E_2
	case 3:
		return ltsmodel.GetTransferDetailObsPeriodEnum().E_3
	case 5:
		return ltsmodel.GetTransferDetailObsPeriodEnum().E_5
	case 6:
		return ltsmodel.GetTransferDetailObsPeriodEnum().E_6
	case 12:
		return ltsmodel.GetTransferDetailObsPeriodEnum().E_12
	case 30:
		return ltsmodel.GetTransferDetailObsPeriodEnum().E_30
	default:
		return ltsmodel.GetTransferDetailObsPeriodEnum().E_2
	}
}

func getPeriodUnit(num int32) string {
	switch num {
	case 2:
		return "min"
	case 5:
		return "min"
	case 30:
		return "min"
	case 1:
		return "hour"
	case 3:
		return "hour"
	case 6:
		return "hour"
	case 12:
		return "hour"
	default:
		return "min"
	}
}

func getPeriodAndUnit(num string) string {
	switch num {
	case "2min":
		return "min"
	case "5min":
		return "min"
	case "30min":
		return "min"
	case "1hour":
		return "hour"
	case "3hour":
		return "hour"
	case "6hour":
		return "hour"
	case "12hour":
		return "hour"
	default:
		return "min"
	}
}

func convertPeriodEnumToInt(period ltsmodel.TransferDetailObsPeriod) int32 {
	switch period {
	case ltsmodel.GetTransferDetailObsPeriodEnum().E_1:
		return 1
	case ltsmodel.GetTransferDetailObsPeriodEnum().E_2:
		return 2
	case ltsmodel.GetTransferDetailObsPeriodEnum().E_3:
		return 3
	case ltsmodel.GetTransferDetailObsPeriodEnum().E_5:
		return 5
	case ltsmodel.GetTransferDetailObsPeriodEnum().E_6:
		return 6
	case ltsmodel.GetTransferDetailObsPeriodEnum().E_12:
		return 12
	case ltsmodel.GetTransferDetailObsPeriodEnum().E_30:
		return 30
	default:
		return 2
	}
}

func convertTransferResp(resp *ltsmodel.CreateTransferResponse) model.CreateTransferResp {
	var logstreams, logstreamNames []string
	for _, logstream := range *resp.LogStreams {
		logstreams = append(logstreams, logstream.LogStreamId)
		logstreamNames = append(logstreamNames, logstream.LogStreamName)
	}
	return model.CreateTransferResp{
		LogGroupId:    *resp.LogGroupId,
		LogGroupName:  *resp.LogGroupName,
		LogTransferId: *resp.LogTransferId,
		LogStreamId:   strings.Join(logstreams, ""),
		LogStreamName: strings.Join(logstreamNames, ""),
		TransferDetail: model.TransferInfo{
			ObsPeriodUnit:   resp.LogTransferInfo.LogTransferDetail.ObsPeriodUnit,
			ObsPeriod:       convertPeriodEnumToInt(resp.LogTransferInfo.LogTransferDetail.ObsPeriod),
			ObsBucketName:   resp.LogTransferInfo.LogTransferDetail.ObsBucketName,
			ObsTransferPath: *resp.LogTransferInfo.LogTransferDetail.ObsTransferPath,
		},
	}
}

// get LTS controller of the resource tenant
func GetLTSController(projectId string) (*LtsResourceController, error) {
	agencyInfo, err := db.GetAgencyInfo(projectId)
	if err != nil {
		return nil, err
	}
	controller, err := newLTSController(agencyInfo.ProjectId, agencyInfo.AgencyName, agencyInfo.DomainId)
	if err != nil {
		return nil, err
	}
	return controller, nil
}

// 新建controller，注册lts client
func newLTSController(projectId string, agencyName string, resDomainId string) (*LtsResourceController, error) {
	credInfo, err := opIamCli.getAgencyCredentialInfo(resDomainId, agencyName)
	if err != nil {
		return nil, err
	}
	cred := basic.NewCredentialsBuilder().
		WithAk(credInfo.Access).
		WithSk(credInfo.Secret).
		WithSecurityToken(credInfo.SecurityToken).
		WithProjectId(projectId).
		Build()

	// 使用资源账号永久AK/SK创VM：①规避功能开启；②当前环境为规避环境；③当前请求的projectID为规避的资源账号projectID
	if setting.AvoidingAgencyEnable && setting.AvoidingAgencyRegion == setting.CloudClientRegion &&
		setting.ResUserProjectId == projectId {
		cred = basic.NewCredentialsBuilder().
			WithAk(string(setting.ResUserAk)).
			WithSk(string(setting.ResUserSk)).
			WithProjectId(setting.ResUserProjectId).
			Build()
	}

	return &LtsResourceController{
		ltsClient: newLtsClient(cred, projectId),
		expireAt:  time.Now().Add(agencyValidDuration),
		projectId: projectId,
	}, nil
}

func newLtsClient(cred basic.Credentials, regionId string) *lts.LtsClient {
	return lts.NewLtsClient(
		lts.LtsClientBuilder().
			WithEndpoint(setting.CloudClientLTSEndpoint).
			WithCredential(cred).
			WithHttpConfig(httpConfig).
			Build())
}
