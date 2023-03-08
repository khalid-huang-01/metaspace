// Copyright (c) Huawei Technologies Co., Ltd. 2022-2022. All rights reserved.

// influxdb控制
package influxdb

import (
	"encoding/json"
	"fmt"
	"time"

	influx "github.com/influxdata/influxdb1-client/v2"
	"github.com/pkg/errors"

	"scase.io/application-auto-scaling-service/pkg/setting"
	"scase.io/application-auto-scaling-service/pkg/utils/logger"
)

const (
	percentAvailableServerSession = "(sum(max_server_session_num)-sum(server_session_count))/sum(max_server_session_num)"
	usedServerSession             = "sum(server_session_count)"
	maxServerSession              = "sum(max_server_session_num)"
	tagsOfGroupByInstance         = "scaling_group_id,instance_id"
)

const (
	queryDuration     int64 = 10 * 1e9
	pingTimeOutSecond       = 3 * time.Second
)

type Controller struct {
	client        influx.Client
	database      string
	measurement   string
	timePrecision string
}

func NewController() (*Controller, error) {
	conf := influx.HTTPConfig{
		Addr:               setting.InfluxAddress,
		Username:           setting.InfluxUser,
		Password:           string(setting.InfluxPassword),
		InsecureSkipVerify: true,
	}
	client, err := influx.NewHTTPClient(conf)
	if err != nil {
		return nil, errors.Wrap(err, "get new influx http client failed")
	}
	_, _, err = client.Ping(pingTimeOutSecond)
	if err != nil {
		return nil, errors.Wrap(err, "connect influx failed")
	}
	c := Controller{
		client:        client,
		database:      setting.InfluxDatabase,
		measurement:   setting.GetInfluxdbMeasurement(),
		timePrecision: "ns",
	}
	return &c, nil
}

// GetServerSessionMetricsOfScalingGroup 获取伸缩的ServerSession相关指标
func (c *Controller) GetServerSessionMetricsOfScalingGroup(log *logger.FMLogger,
	groupID string) (*GroupServerSessionMetrics, error) {
	command := fmt.Sprintf("SELECT %s AS PRE,%s AS MAX,%s AS USED FROM %s WHERE scaling_group_id = '%s' "+
		"AND time >= now()-20s AND time < now()-10s",
		percentAvailableServerSession, maxServerSession, usedServerSession, c.measurement, groupID)
	log.Info("influxDB query command of getting server session metrics: [%s]", command)
	q := influx.NewQuery(command, c.database, c.timePrecision)
	resp, err := c.client.Query(q)
	if err != nil {
		return nil, err
	} else if resp.Error() != nil {
		return nil, resp.Error()
	}
	if resp.Results == nil || resp.Results[0].Series == nil {
		return nil, nil
	}

	if resp.Results[0].Series[0].Values != nil {
		timestamp, err := getInt64ForInfluxValue(resp.Results[0].Series[0].Values[0][0])
		if err != nil {
			return nil, err
		}
		pre, err := getFloat64ForInfluxValue(resp.Results[0].Series[0].Values[0][1])
		if err != nil {
			return nil, err
		}
		max, err := getInt64ForInfluxValue(resp.Results[0].Series[0].Values[0][2])
		if err != nil {
			return nil, err
		}
		used, err := getInt64ForInfluxValue(resp.Results[0].Series[0].Values[0][3])
		if err != nil {
			return nil, err
		}
		return &GroupServerSessionMetrics{
			AvailablePercent: pre,
			MaxNum:           max,
			UsedNum:          used,
			StartTime:        timestamp,
			EndTime:          timestamp + queryDuration,
		}, nil
	}
	return nil, nil
}

// GetTopUsedServerSessionOfInstance 获取伸缩组中UsedServerSession前N的实例id列表
func (c *Controller) GetTopUsedServerSessionOfInstance(log *logger.FMLogger, groupId string,
	n float64, start, end int64) []string {
	var instances []string
	availableServerSession := fmt.Sprintf("0-%s", usedServerSession)
	command := fmt.Sprintf("SELECT TOP(Avail,instance_id,%d) FROM (SELECT %s AS Avail FROM %s "+
		"WHERE scaling_group_id = '%s' AND time >= %dns AND time < %dns GROUP BY %s)",
		int(n), availableServerSession, c.measurement, groupId, start, end, tagsOfGroupByInstance)
	log.Info("influxDB query command of getting topN UsedServerSession instances: [%s]", command)
	q := influx.NewQuery(command, c.database, c.timePrecision)
	resp, err := c.client.Query(q)
	if err != nil || resp.Error() != nil {
		return instances
	}
	if resp.Results == nil || resp.Results[0].Series == nil {
		return instances
	}

	for _, v := range resp.Results[0].Series[0].Values {
		groupID, ok := v[2].(string)
		if !ok {
			return nil
		}
		instances = append(instances, groupID)
	}
	return instances
}

func getFloat64ForInfluxValue(influxValue interface{}) (float64, error) {
	valueJson, ok := influxValue.(json.Number)
	if !ok {
		return -1, errors.New("It is failed to turn interface to json.Number")
	}
	value, err := valueJson.Float64()
	if err != nil {
		return -1, err
	}
	return value, nil
}

func getInt64ForInfluxValue(influxValue interface{}) (int64, error) {
	valueJson, ok := influxValue.(json.Number)
	if !ok {
		return -1, errors.New("It is failed to turn interface to json.Number")
	}
	value, err := valueJson.Int64()
	if err != nil {
		return -1, err
	}
	return value, nil
}
