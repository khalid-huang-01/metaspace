package lts

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

type CreateAccessConfigReq struct {
	FleetId     string    `json:"fleet_id" validate:"required,min=1,max=64" reg_error_info:"Incorrect format"`
	LtsConfig   LtsConfig `json:"lts_config" validate:"dive"`
	Description string    `json:"description" validate:"min=0,max=128" reg_error_info:"Incorrect format"`
}

type CreateAccessConfigReqToAASS struct {
	CreateAccessConfigReq
	EnterpriseProjectId string `json:"enterprise_project_id" validate:"min=0,max=64"`
}
type LtsConfig struct {
	LtsConfitName string   `json:"lts_access_config_name" validate:"required,min=1,max=64" reg_error_info:"Incorrect format"`
	LogGroupId    string   `json:"log_group_id" validate:"required,min=1,max=64" reg_error_info:"Incorrect format"`
	LogGroupPath  []string `json:"lts_log_path" validate:"required,dive,checkPath" reg_error_info:"Incorrect format"`
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

type UpdateAccessConfigToDB struct {
	AccessConfigId string `json:"access_config_id" validate:"required,min=1,max=64"`
	Description    string `json:"description" validate:"min=0,max=128"`
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

type DeleteHostGroupReq struct {
	HostGroupIdList []string
}

type ListAccessConfig struct {
	Total            int            `json:"total"`
	Count            int            `json:"count"`
	AccessConfigList []AccessConfig `json:"access_config"`
}
type AccessConfig struct {
	Id                  string   `json:"id"`
	AccessConfigName    string   `json:"access_config_name"`
	AccessConfigId      string   `json:"access_config_id"`
	HostGroupIDList     []string `json:"host_group_id_list"`
	HostGroupName       string   `json:"host_group_name"`
	LogGroupId          string   `json:"log_group_id"`
	LogGroupName        string   `json:"log_group_name"`
	LogStreamName       string   `json:"log_stream_name"`
	LogStreamId         string   `json:"log_stream_id"`
	LogStreamLink       string   `json:"log_stream_link"`
	LogConfigPath       []string `json:"log_config_path"`
	ObsTransferPath     string   `json:"obs_transfer_path"`
	ObsTransferPathLink string   `json:"obs_transfer_path_link"`
	CreateTime          string   `json:"create_time"`
	Description         string   `json:"description"`
}

type LogTransferReq struct {
	TransferInfo TransferInfo `json:"transfer_info" validate:"dive"`
	LogStreamId  string       `json:"log_stream_id"`
	LogGroupId   string       `json:"log_group_id"`
}

type TransferResp struct {
	LogGroupId     string       `json:"log_group_id"`
	LogGroupName   string       `json:"log_group_name"`
	LogStreamId    string       `json:"log_stream_id"`
	LogStreamName  string       `json:"log_stream_name"`
	LogTransferId  string       `json:"log_transfer_id"`
	TransferDetail TransferInfo `json:"transfer_info"`
}
type TransferInfo struct {
	ObsPeriodUnit   string `json:"obs_period_unit" validate:"min=1,max=5"`
	ObsBucketName   string `json:"obs_bucket_name" validate:"min=1,max=64"`
	ObsPeriod       int    `json:"obs_period" validate:"gte=1,lte=30"`
	ObsTransferPath string `json:"obs_transfer_path" validate:"required,checkPrefix"`
}

type CreateTransferTemplate struct {
	LogGroupId     string               `json:"log_group_id"`
	LogGroupName   string               `json:"log_group_name"`
	LogStreamId    string               `json:"log_stream_id"`
	LogStreamName  string               `json:"log_stream_name"`
	LogTransferId  string               `json:"log_transfer_id"`
	TransferDetail TransferInfoTemplate `json:"transfer_info"`
}

type TransferInfoTemplate struct {
	ObsBucketName   string `json:"obs_bucket_name" validate:"min=1,max=64"`
	ObsPeriod       string `json:"obs_period" validate:"gte=1,lte=10"`
	ObsTransferPath string `json:"obs_transfer_path" validate:"required,checkPrefix"`
}

type ListTransfersInfo struct {
	Total         int                `json:"total"`
	Count         int                `json:"count"`
	ListTransfers []ListTransferResp `json:"log_transfer_list"`
}

type ListTransferResp struct {
	LogGroupId     string       `json:"log_group_id"`
	LogStreamId    string       `json:"log_stream_id"`
	LogTransferId  string       `json:"log_transfer_id"`
	TransferDetail TransferInfo `json:"transfer_info"`
}
