// Copyright (c) Huawei Technologies Co., Ltd. 2022-2022. All rights reserved.

// 配置管理与初始化
package configmanager

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"runtime"
	"sync"
	"time"

	"codehub-g.huawei.com/videocloud/mediaprocesscenter/auxproxy/config"
	"codehub-g.huawei.com/videocloud/mediaprocesscenter/auxproxy/pkg/apis"
	"codehub-g.huawei.com/videocloud/mediaprocesscenter/auxproxy/pkg/utils/clients"
	"codehub-g.huawei.com/videocloud/mediaprocesscenter/auxproxy/pkg/utils/log"
)

const (
	publicIPType  = "public"
	privateIPType = "private"
)

type Config struct {
	FleetID        string
	ScalingGroupID string
	InstanceID     string
	PublicIP       string
	PrivateIP      string
	InstanceConfig apis.InstanceConfiguration
}

type ConfigManager struct {
	Interval     time.Duration
	GatewayAddr  string
	AuxProxyPort int

	ConfigGetChan chan bool
	Config        Config

	StopChan chan int
}

type Meta struct {
	// 应用下载使用的metadata
	BuildID              string `json:"build_id"`
	Bucket               string `json:"bucket"`
	Object               string `json:"object"`
	Region               string `json:"region"`
	GlobalServiceAddress string `json:"global_service_address"`
	Ak                   string `json:"ak"`
	Sk                   string `json:"sk"`
	// 弹缩使用的metadata
	FleetID        string `json:"fleet_id"`
	GatewayAddress string `json:"gateway_address"`
	ScalingGroupID string `json:"scaling_group_id"`
}

type Metadata struct {
	InstanceType string `json:"instance_type"`
	Meta         Meta   `json:"meta"`
}

var ConfMgr *ConfigManager

// InitConfigManager init config manager
func InitConfigManager(interval time.Duration) (retErr error) {
	once := sync.Once{}
	once.Do(func() {
		// get fleet id
		fleetID, gwAddr, sgID, err := getFleetIDAndGWAddrAndScalingGroupID()
		if err != nil {
			retErr = fmt.Errorf("[config manager] failed to get fleet id, gateway address, sg id for %v", err)
			return
		}
		if config.Opts.GatewayAddr != "" {
			gwAddr = config.Opts.GatewayAddr
		}

		if config.Opts.CmdFleetId != "" {
			fleetID = config.Opts.CmdFleetId
		}

		if config.Opts.ScalingGroupId != "" {
			sgID = config.Opts.ScalingGroupId
		}

		instanceID, publicIP, privateIP, err := parseParam()
		if err != nil {
			retErr = err
			return
		}

		ConfMgr = &ConfigManager{
			Interval:      interval,
			GatewayAddr:   gwAddr,
			AuxProxyPort:  config.HttpsPort,
			ConfigGetChan: make(chan bool),
			Config: Config{
				FleetID:        fleetID,
				ScalingGroupID: sgID,
				InstanceID:     instanceID,
				PublicIP:       publicIP,
				PrivateIP:      privateIP,
				InstanceConfig: apis.InstanceConfiguration{},
			},
			StopChan: make(chan int),
		}
	})

	return
}

func parseParam() (string, string, string, error) {
	// get instance id
	instanceID, err := getInstanceID()
	if err != nil {
		return "", "", "", fmt.Errorf("[config manager] failed to get instance id for %v", err)
	}

	// get public ip
	publicIP, err := getPublicIPOrPrivateIP(publicIPType)
	if err != nil {
		return "", "", "", fmt.Errorf("[config manager] failed to get public ip for %v", err)
	}

	// get private ip
	privateIP, err := getPublicIPOrPrivateIP(privateIPType)
	if err != nil {
		return "", "", "", fmt.Errorf("[config manager] failed to get private ip for %v", err)
	}
	return instanceID, publicIP, privateIP, nil
}

// Stop stops config work
func (c *ConfigManager) Stop() {
	c.StopChan <- 1
}

// Work let config manager work
func (c *ConfigManager) Work() {
	go c.work()
}

func (c *ConfigManager) work() {
	ticker := time.NewTicker(c.Interval)

	c.fetchConfiguration()
	c.ConfigGetChan <- true

	for {
		select {
		case <-ticker.C:
			c.fetchConfiguration()
		case <-c.StopChan:
			log.RunLogger.Infof("[config manager] get config stopped")
			return
		}
	}
}

func (c *ConfigManager) fetchConfiguration() {
	log.RunLogger.Infof("[config manager] time to check configuration (every %v)", c.Interval)
	configuration, err := clients.GWClient.FetchConfiguration(ConfMgr.Config.ScalingGroupID)
	if err != nil {
		log.RunLogger.Errorf("fetch configuration from gateway failed, err %v", err)
		return
	}
	c.Config.InstanceConfig = *configuration
	log.RunLogger.Infof("[config manager] has config %+v", c)
}

func getFleetIDAndGWAddrAndScalingGroupID() (string, string, string, error) {
	if runtime.GOOS == "windows" {
		return getFleetIDAndGWAddrAndScalingGroupIDFromFile()
	}

	cli := clients.NewHttpsClientWithoutCerts()

	req, err := clients.NewRequest("GET",
		fmt.Sprintf("%s/openstack/latest/user_data", config.Opts.CloudPlatformAddr),
		map[string][]string{}, nil)
	if err != nil {
		return "", "", "", err
	}

	code, buf, _, err := clients.DoRequest(cli, req)
	if err != nil || code != http.StatusOK {
		return "", "", "", fmt.Errorf("code %d or err %v", code, err)
	}

	var fleetIDStruct struct {
		FleetID        string `json:"fleet_id"`
		GatewayAddress string `json:"gateway_address"`
		ScalingGroupID string `json:"scaling_group_id"`
	}

	err = json.Unmarshal(buf, &fleetIDStruct)
	if err != nil {
		return "", "", "", err
	}

	return fleetIDStruct.FleetID, fleetIDStruct.GatewayAddress, fleetIDStruct.ScalingGroupID, nil
}

func getFleetIDAndGWAddrAndScalingGroupIDFromFile() (string, string, string, error) {
	buf, err := ioutil.ReadFile("C:/config.txt")
	if err != nil {
		log.RunLogger.Errorf("[config manager] getFleetIDAndGWAddrAndScalingGroupIDFromFile err:%v", err)
		return "", "", "", err
	}

	var fleetIDStruct struct {
		FleetID        string `json:"fleet_id"`
		GatewayAddress string `json:"gateway_address"`
		ScalingGroupID string `json:"scaling_group_id"`
	}

	err = json.Unmarshal(buf, &fleetIDStruct)
	if err != nil {
		return "", "", "", err
	}

	return fleetIDStruct.FleetID, fleetIDStruct.GatewayAddress, fleetIDStruct.ScalingGroupID, nil
}

func GetMetaDataConfig() (*Meta, error) {
	cli := clients.NewHttpsClientWithoutCerts()

	req, err := clients.NewRequest("GET",
		fmt.Sprintf("%s/openstack/latest/meta_data.json", config.Opts.CloudPlatformAddr),
		map[string][]string{}, nil)
	if err != nil {
		return nil, err
	}

	var code int
	var buf []byte
	var reqErr error
	reqOK := false
	// 获取metadata参数，失败每隔2s重试一次，最多30次
	for i := 0; i < 30; i++ {
		code, buf, _, reqErr = clients.DoRequest(cli, req)
		if err != nil || code != http.StatusOK {
			time.Sleep(2 * time.Second)
			continue
		}
		reqOK = true
	}

	if !reqOK {
		return nil, fmt.Errorf("code %d or err %v, resp:%v", code, reqErr, string(buf))
	}

	metadata := &Metadata{}
	log.RunLogger.Errorf("[config manager] GetMetaDataConfig string:%v, buf:%v", string(buf), buf)
	err = json.Unmarshal(buf, &metadata)
	if err != nil || metadata == nil {
		return nil, err
	}

	meta := metadata.Meta
	log.RunLogger.Errorf("[config manager] GetMetaDataConfig success, meta:%+v", meta)
	return &meta, nil
}

func getInstanceID() (string, error) {
	cli := clients.NewHttpsClientWithoutCerts()

	req, err := clients.NewRequest("GET",
		fmt.Sprintf("%s/openstack/latest/meta_data.json", config.Opts.CloudPlatformAddr),
		map[string][]string{}, nil)
	if err != nil {
		return "", err
	}

	code, buf, _, err := clients.DoRequest(cli, req)
	if err != nil || code != http.StatusOK {
		return "", fmt.Errorf("code %d or err %v", code, err)
	}

	var instanceIDStruct struct {
		InstanceID string `json:"uuid"`
	}

	err = json.Unmarshal(buf, &instanceIDStruct)
	if err != nil {
		return "", err
	}

	return instanceIDStruct.InstanceID, nil
}

func getPublicIPOrPrivateIP(ipType string) (string, error) {
	cli := clients.NewHttpsClientWithoutCerts()

	var url string

	switch ipType {
	case publicIPType:
		url = fmt.Sprintf("%s/latest/meta-data/%s", config.Opts.CloudPlatformAddr, "public-ipv4")
	case privateIPType:
		url = fmt.Sprintf("%s/latest/meta-data/%s", config.Opts.CloudPlatformAddr, "local-ipv4")
	default:
		return "", fmt.Errorf("error ip type")
	}

	req, err := clients.NewRequest("GET", url, map[string][]string{}, nil)
	if err != nil {
		return "", err
	}

	code, buf, _, err := clients.DoRequest(cli, req)
	if err != nil || code != http.StatusOK {
		return "", fmt.Errorf("code %d or err %v", code, err)
	}

	return string(buf), nil
}
