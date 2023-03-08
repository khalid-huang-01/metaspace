// Copyright (c) Huawei Technologies Co., Ltd. 2022-2022. All rights reserved.

// 弹性伸缩策略相关方法
package metrics

import (
	"fmt"
	"strings"
	"time"

	client "github.com/influxdata/influxdb1-client/v2"

	"codehub-g.huawei.com/videocloud/mediaprocesscenter/application-gateway/config"
	"codehub-g.huawei.com/videocloud/mediaprocesscenter/application-gateway/pkg/common"
	"codehub-g.huawei.com/videocloud/mediaprocesscenter/application-gateway/pkg/distributedlock"
	"codehub-g.huawei.com/videocloud/mediaprocesscenter/application-gateway/pkg/models"
	app_process "codehub-g.huawei.com/videocloud/mediaprocesscenter/application-gateway/pkg/models/appprocess"
	"codehub-g.huawei.com/videocloud/mediaprocesscenter/application-gateway/pkg/utils/log"
	"codehub-g.huawei.com/videocloud/mediaprocesscenter/application-gateway/pkg/utils/security"
)

const (
	// retentionPolicyColumnsName 保存策略的名称列
	retentionPolicyColumnsName = "name"

	// retentionPolicyColumnsDuration 保存策略的时长列
	retentionPolicyColumnsDuration = "duration"

	// retentionPolicyColumnsDefault 保存策略的默认列
	retentionPolicyColumnsDefault = "default"

	// defaultRetentionPolicyName 默认保存策略名称
	defaultRetentionPolicyName = "autogen"

	// defaultRetentionPolicyDuration 默认保存策略时长
	defaultRetentionPolicyDuration = "336h0m0s"
)

type MetricClient struct {
	influxDBAddr     string
	influxDBDatabase string
	influxDBUsername string
	influxDBPassword string

	influxDBClient client.Client

	uploadDuration time.Duration
}

var metricClient *MetricClient

func initInfluxDataBase(influxClient client.Client, database string) error {
	// 查询所配置的InfluxDataBase是否存在
	q := client.Query{Command: fmt.Sprintf(`SHOW DATABASES`)}
	resp, err := influxClient.Query(q)
	if err != nil {
		log.RunLogger.Errorf("It is failed to show databases,err:%s", err.Error())
		return err
	}

	if len(resp.Results) != 0 && len(resp.Results[0].Series) != 0 && len(resp.Results[0].Series[0].Values) != 0 {
		for i := 0; i < len(resp.Results[0].Series[0].Values); i++ {
			name, ok := resp.Results[0].Series[0].Values[i][0].(string)
			if !ok {
				log.RunLogger.Errorf("It is failed to turn interface to string")
				return err
			}
			if name == database {
				// 所配置的InfluxDataBase存在，无需后续操作
				log.RunLogger.Infof("The database(%s) exists", database)
				return err
			}
		}
	}
	// 所配置的InfluxDataBase不存在，创建数据库
	q = client.Query{Command: fmt.Sprintf(`CREATE DATABASE "%s"`, database)}
	_, err = influxClient.Query(q)
	if err != nil {
		log.RunLogger.Errorf("It is failed to create database,err:%s", err.Error())
		return err
	}

	log.RunLogger.Infof("Create the database(%s) successfully", database)

	return nil
}

func initInfluxRetentionPolicies(influxClient client.Client, database string) error {
	q := client.Query{Command: fmt.Sprintf(`SHOW RETENTION POLICIES ON "%s"`, database)}
	resp, err := influxClient.Query(q)
	if err != nil {
		log.RunLogger.Errorf("It is failed to show databases,err:%s", err.Error())
		return err
	}

	if len(resp.Results) != 0 && len(resp.Results[0].Series) != 0 && len(resp.Results[0].Series[0].Values) != 0 {
		err, isContinue := handleSeriesValue(resp, database, influxClient)
		if err != nil {
			return err
		}
		if !isContinue {
			return nil
		}
	}

	// 数据库中不存在目标保存策略，需进行保存策略创建
	q = client.Query{Command: fmt.Sprintf(`CREATE RETENTION POLICY "%s" ON "%s" DURATION %s REPLICATION 1 DEFAULT`,
		defaultRetentionPolicyName, database, defaultRetentionPolicyDuration)}
	_, err = influxClient.Query(q)
	if err != nil {
		log.RunLogger.Errorf("It is failed to create retention policy,err:%s", err.Error())
		return err
	}
	log.RunLogger.Infof("Create the duration of retention policy(%s) on database(%s) to %s",
		defaultRetentionPolicyName, database, defaultRetentionPolicyDuration)

	return nil
}

func handleSeriesValue(resp *client.Response, database string, influxClient client.Client) (error, bool) {
	keyWords := []string{retentionPolicyColumnsName, retentionPolicyColumnsDuration, retentionPolicyColumnsDefault}
	index := GetIndexFromColumnsWithKeyWords(resp.Results[0].Series[0].Columns, keyWords)
	if len(index) < 3 { // index is always 3 as following using
		return fmt.Errorf("the length of series[0]'s column is less than three"), false
	}
	for i := 0; i < len(resp.Results[0].Series[0].Values); i++ {
		name, ok := resp.Results[0].Series[0].Values[i][index[0]].(string)
		if !ok {
			log.RunLogger.Errorf("[metrics client] type of Series[0].Values[%d][%d] is not string",
				i, index[0])
			return fmt.Errorf("the type of series's value is no valid"), false
		}
		duration, ok := resp.Results[0].Series[0].Values[i][index[1]].(string)
		if !ok {
			log.RunLogger.Errorf("[metrics client] type of Series[0].Values[%d][%d] is not string",
				i, index[1])
			return fmt.Errorf("the type of series's value is no valid"), false
		}
		isDefault, ok := resp.Results[0].Series[0].Values[i][index[2]].(bool)
		if !ok {
			log.RunLogger.Errorf("[metrics client] type of Series[0].Values[%d][%d] is not bool",
				i, index[2])
			return fmt.Errorf("the type of series's value is no valid"), false
		}
		if name == defaultRetentionPolicyName && duration == defaultRetentionPolicyDuration && isDefault == true {
			log.RunLogger.Infof("The duration of retention policy(%s) on database(%s) is %s",
				name, database, duration)
			return nil, false
		} else if name == defaultRetentionPolicyName &&
			(duration != defaultRetentionPolicyDuration || isDefault == false) {
			return alterRetentionPolicy(name, database, influxClient), false
		}
	}
	return nil, true
}

func alterRetentionPolicy(name, database string, influxClient client.Client) error {
	q := client.Query{Command: fmt.Sprintf(`ALTER RETENTION POLICY "%s" on "%s" DURATION %s DEFAULT`,
		name, database, defaultRetentionPolicyDuration)}
	_, err := influxClient.Query(q)
	if err != nil {
		return err
	}
	log.RunLogger.Infof("Update the duration of retention policy(%s) on database(%s) to %s",
		name, database, defaultRetentionPolicyDuration)
	return nil
}

// GetIndexFromColumnsWithKeyWords 根据keywords中的值确定其在columns中的下标值
func GetIndexFromColumnsWithKeyWords(columns []string, keywords []string) []int {
	indexList := make([]int, len(keywords))
	for i := 0; i < len(keywords); i++ {
		for j := 0; j < len(columns); j++ {
			if strings.Contains(columns[j], keywords[i]) {
				indexList[i] = j
				break
			}
		}
	}
	return indexList
}

func Init() {
	w := &influxdbMetricWorker{}
	if config.GlobalConfig.DeployModel == config.DeployModelSingleton {
		go func() {
			InitMetric()
			GetMetricClient().Work(w.stopCh)
		}()
		return
	}
	m := distributedlock.NewDistributedLockController(common.LockMetric, common.LockBizCategory, w)
	m.Work()
}

type influxdbMetricWorker struct {
	stopCh chan struct{}
}

func (w *influxdbMetricWorker) HolderHook() {
	log.RunLogger.Infof("[influxdb worker] start to upload metrics")
	w.stopCh = make(chan struct{}, 0)
	InitMetric()
	GetMetricClient().Work(w.stopCh)
}

func (w *influxdbMetricWorker) CompetitorHook() {
	log.RunLogger.Infof("[influxdb worker] stop to upload metrics")
	close(w.stopCh)
}

// InitMetric 初始化influxDB
func InitMetric() {
	ticker := time.Tick(time.Second * 10)
	for t := range ticker {
		err := InitMetricClient()
		if err == nil {
			break
		} else {
			log.RunLogger.Errorf("[metrics] connect influxdb failed for %v", err, t)
		}
	}

	return
}

// InitMetricClient init metric client
func InitMetricClient() error {
	plainPw, err := security.GCM_Decrypt(config.GlobalConfig.InfluxPassword, 
		config.GlobalConfig.GCMKey, config.GlobalConfig.GCMNonce)
	if err != nil {
		log.RunLogger.Errorf("[metric client] failed to decrypt password for %v", err)
		return err
	}

	cli, err := client.NewHTTPClient(client.HTTPConfig{
		Addr:               fmt.Sprintf("https://%s", config.GlobalConfig.InfluxAddr),
		Username:           config.GlobalConfig.InfluxUsername,
		Password:           plainPw,
		InsecureSkipVerify: true,
	})
	if err != nil {
		log.RunLogger.Errorf("[metrics] failed to create a influxdb client for %v", err)
		return err
	}

	dur, ver, err := cli.Ping(3 * time.Second)
	if err != nil {
		log.RunLogger.Errorf("[metrics] connect influxdb failed for %v", err)
		return err
	}

	log.RunLogger.Infof("[metrics] connected to influxDB successfully! cost:%v, version: %s", dur, ver)

	// init database
	err = initInfluxDataBase(cli, config.GlobalConfig.InfluxDBName)
	if err != nil {
		return err
	}
	err = initInfluxRetentionPolicies(cli, config.GlobalConfig.InfluxDBName)
	if err != nil {
		return err
	}

	metricClient = &MetricClient{
		influxDBAddr:     config.GlobalConfig.InfluxAddr,
		influxDBDatabase: config.GlobalConfig.InfluxDBName,
		influxDBUsername: config.GlobalConfig.InfluxUsername,
		influxDBPassword: config.GlobalConfig.InfluxPassword,
		influxDBClient:   cli,
		uploadDuration:   10 * time.Second, // default update duration
	}

	return nil
}

// GetMetricClient return metric client
func GetMetricClient() *MetricClient {
	return metricClient
}

// Work let metric work
func (m *MetricClient) Work(stopCh chan struct{}) {
	go m.work(stopCh)
}

func (m *MetricClient) work(stopCh chan struct{}) {
	ticker := time.NewTicker(m.uploadDuration)
	appProcessDao := app_process.NewAppProcessDao(models.MySqlOrm)
	for {
		select {
		case <-ticker.C:
			log.RunLogger.Infof("[metric client] upload metrics to influxdb")
			apDB, err := appProcessDao.GetAllActiveAppProcess()
			if err != nil {
				log.RunLogger.Errorf("[metric client] failed to get all active processes for %v", err)
				continue
			}

			metrics := make([]*Metric, len(apDB))
			for i, ap := range apDB {
				metric := &Metric{
					ID:                  ap.ID,
					InstanceID:          ap.InstanceID,
					ScalingGroupID:      ap.ScalingGroupID,
					FleetID:             ap.FleetID,
					ServerSessionCount:  ap.ServerSessionCount,
					MaxServerSessionNum: ap.MaxServerSessionNum,
				}
				metrics[i] = metric
			}
			err = m.WriteMetrics(metrics)
			if err != nil {
				log.RunLogger.Errorf("[metric client] failed to write metric for %v", err)
				continue
			}
		case <-stopCh:
			ticker.Stop()
			return
		}
	}
}
