package model

import (
	"strings"

	"scase.io/application-auto-scaling-service/pkg/db"
)

// Log Group
type CreateLogGroup struct {
	LogGroupName string `json:"log_group_name" validate:"required,min=1,max=64" reg_error_info:"Incorrect format"`
	TTLInDay     int    `json:"ttl_in_day"  validate:"required,gte=1,lte=30" reg_error_info:"Incorrect format"`
}
type CreateLogGroupResp struct {
	LogGroupName string `json:"log_group_name"`
	LogGroupId   string `json:"log_group_id"`
	TTLInDay     int    `json:"ttl_in_day"`
}

type ListLogGroups struct {
	Total     int         `json:"total"`
	LogGroups []LogGroups `json:"log_groups"`
}
type LogGroups struct {
	CreationTime string `json:"creation_time"`
	LogGroupName string `json:"log_group_name"`
	LogGroupID   string `json:"log_group_id"`
	TTLInDays    int    `json:"ttl_in_days"`
}

// Log Stream
type CreateLogStreamReq struct {
	LogGroupId    string `json:"log_group_id"`
	LogStreamName string `json:"log_stream_name"`
	EnterpriseProjectId string `json:"enterprise_project_id"`
}

type ListLogStreams struct {
	Total      int          `json:"total"`
	LogStreams []LogStreams `json:"log_streams"`
}
type LogStreams struct {
	LogStreamName string `json:"log_stream_name"`
	LogStreamID   string `json:"log_stream_id"`
}

type DeleteLogStreamReq struct {
	LogGroupId  string `json:"log_group_id"`
	LogStreamId string `json:"log_stream_id"`
}

// Log Host Group
type CreateHostGroupReq struct {
	HostGroupName string `json:"host_group_name" validate:"required,min=1,max=64"`
}

type DeleteHostGroupReq struct {
	HostGroupIdList []string
}

type UpdateHostGroupReq struct {
	HostGroupId string 
	HostIdList  []string
}

type CreateHostGroupResp struct {
	HostGroupID   string   `json:"host_group_id"`
	HostGroupName string   `json:"host_group_name"`
	HostGroupType string   `json:"host_group_type"`
	HostIDList    []string `json:"host_id_list"`
	CreateTime    int64    `json:"create_time"`
	UpdateTime    int64    `json:"update_time"`
}

// Access Config
type CreateAccessConfigReqToCloud struct {
	AccessConfigName string   `json:"access_config_name"`
	HostGroupIDList  []string `json:"host_group_id_list"`
	HostGroupName    string   `json:"host_group_name"`
	LogGroupId       string   `json:"log_group_id"`
	LogGroupName     string   `json:"log_group_name"`
	LogStreamName    string   `json:"log_stream_name"`
	LogStreamId      string   `json:"log_stream_id"`
	LogConfigPath    []string `json:"log_config_path"`
	EnterpriseProjectId string `json:"enterprise_project_id"`
	Description      string   `json:"description"`
}

type CreateAccessConfigResp struct {
	Id               string   `json:"id"`
	FleetId          string   `json:"fleet_id"`
	LogGroupId       string   `json:"log_group_id"`
	LogGroupName     string   `json:"log_group_name"`
	LogStreamId      string   `json:"log_stream_id"`
	LogStreamName    string   `json:"log_stream_name"`
	HostGroupID      string   `json:"host_group_id"`
	HostGroupName    string   `json:"host_group_name"`
	LogConfigPath    []string `json:"log_config_path"`
	AccessConfigName string   `json:"access_config_name"`
	AccessConfigId   string   `json:"access_config_id"`
	Description      string   `json:"description"`
}

type ListAccessConfig struct {
	Total            int            `json:"total"`
	Count            int            `json:"count"`
	AccessConfigList []AccessConfig `json:"access_config"`
}
type AccessConfig struct {
	Id               string   `json:"id"`
	AccessConfigName string   `json:"access_config_name"`
	AccessConfigId   string   `json:"access_config_id"`
	HostGroupIDList  []string `json:"host_group_id_list"`
	HostGroupName    string   `json:"host_group_name"`
	LogGroupId       string   `json:"log_group_id"`
	LogGroupName     string   `json:"log_group_name"`
	LogStreamName    string   `json:"log_stream_name"`
	LogStreamId      string   `json:"log_stream_id"`
	LogConfigPath    []string `json:"log_config_path"`
	ObsTransferPath  string   `json:"obs_transfer_path"`
	CreateTime       string   `json:"create_time"`
	Description      string   `json:"description"`
}

type CreateAccessConfig struct {
	FleetId     string    `json:"fleet_id" validate:"required,min=1,max=64" reg_error_info:"Incorrect format"`
	LtsConfig   LTSConfig `json:"lts_config" validate:"dive"`
	Description string    `json:"description" validate:"min=0,max=128" reg_error_info:"Incorrect format"`
	EnterpriseProjectId string   `json:"enterprise_project_id" validate:"required,min=0,max=64"`
}

type LTSConfig struct {
	LtsAccessConfigName string   `json:"lts_access_config_name" validate:"required,min=1,max=64"`
	LogGroupId          string   `json:"log_group_id" validate:"required,min=1,max=64"`
	LogGroupName        string   `json:"log_group_name" validate:"required,min=1,max=64"`
	LtsLogPath          []string `json:"lts_log_path" validate:"required,dive,checkPrefix,max=300"`
	LtsLogStreamName    string   `json:"log_stream_name" validate:"required,min=1,max=64"`
	LtsHostGroupName    string   `json:"host_group_name" validate:"required,min=1,max=64"`
}

type FleetListReq struct {
	FleetIdList []string `json:"fleet_id_list"`
}

type UpdateAccessConfig struct {
	AccessConfigId string `json:"access_config_id"`
}
type UpdateAccessConfigToDB struct {
	AccessConfigId string `json:"access_config_id" validate:"required,min=1,max=64"`
	Description    string `json:"description" validate:"required,min=1,max=128"`
}

// Transfer
type LogTransferReq struct {
	TransferInfo TransferInfo `json:"transfer_info"`
	LogStreamId  string       `json:"log_stream_id"`
	LogGroupId   string       `json:"log_group_id"`
	EnterpriseProjectId string	`json:"enterprise_project_id"`
}

type TransferInfo struct {
	ObsPeriodUnit   string `json:"obs_period_unit"`
	ObsBucketName   string `json:"obs_bucket_name"`
	ObsPeriod       int32  `json:"obs_period"`
	ObsTransferPath string `json:"obs_transfer_path"`
}

type CreateTransferResp struct {
	LogGroupId     string       `json:"log_group_id"`
	LogGroupName   string       `json:"log_group_name"`
	LogStreamId    string       `json:"log_stream_id"`
	LogStreamName  string       `json:"log_stream_name"`
	LogTransferId  string       `json:"log_transfer_id"`
	TransferDetail TransferInfo `json:"transfer_info"`
}

type ListTransferResp struct {
	LogGroupId     string       `json:"log_group_id"`
	LogStreamId    string       `json:"log_stream_id"`
	LogTransferId  string       `json:"log_transfer_id"`
	TransferDetail TransferInfo `json:"transfer_info"`
}

type ListTransfersInfo struct {
	Total         int                `json:"total"`
	Count         int                `json:"count"`
	ListTransfers []ListTransferResp `json:"log_transfer_list"`
}

type UpdateTransfer struct {
	TransferId   string       `json:"transfer_id"`
	TransferInfo TransferInfo `json:"transfer_info"`
}

func BuildTransferResp(transfer db.LogTransfer) ListTransferResp {
	return ListTransferResp{
		LogGroupId:    transfer.LogGroupId,
		LogStreamId:   transfer.LogStreamId,
		LogTransferId: transfer.LogTransferId,
		TransferDetail: TransferInfo{
			ObsPeriodUnit:   transfer.ObsPeriodUnit,
			ObsPeriod:       int32(transfer.ObsPeriod),
			ObsBucketName:   transfer.ObsBucketName,
			ObsTransferPath: transfer.ObsTransferPath,
		},
	}
}

func BuildCreateAccessConfigResp(ltsConfig db.LtsConfig) CreateAccessConfigResp {
	return CreateAccessConfigResp{
		Id:               ltsConfig.Id,
		FleetId:          ltsConfig.FleetId,
		AccessConfigName: ltsConfig.AccessConfigName,
		AccessConfigId:   ltsConfig.AccessConfigId,
		LogGroupId:       ltsConfig.LogGroupId,
		LogGroupName:     ltsConfig.LogGroupName,
		LogStreamId:      ltsConfig.LogStreamId,
		LogStreamName:    ltsConfig.LogStreamName,
		LogConfigPath:    strings.Split(ltsConfig.LogConfigPath, ","),
		HostGroupID:      ltsConfig.HostGroupID,
		HostGroupName:    ltsConfig.HostGroupName,
		Description:      ltsConfig.Description,
	}
}
