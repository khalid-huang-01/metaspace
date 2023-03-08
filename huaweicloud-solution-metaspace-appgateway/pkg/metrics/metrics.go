// Copyright (c) Huawei Technologies Co., Ltd. 2022-2022. All rights reserved.

// 策略相关结构定义
package metrics

import (
	client "github.com/influxdata/influxdb1-client/v2"
)

const (
	MeasurementNameProcess       = "process"
	TagNameProcessID             = "id"
	TagNameFleetID               = "fleet_id"
	TagNameInstanceID            = "instance_id"
	TagNameScalingGroupID        = "scaling_group_id"
	FieldNameServerSessionCount  = "server_session_count"
	FieldNameMaxServerSessionNum = "max_server_session_num"
)

type Metric struct {
	ID                  string
	InstanceID          string
	ScalingGroupID      string
	FleetID             string
	ServerSessionCount  int
	MaxServerSessionNum int
}

func (m *MetricClient) writeMetrics(pts []*client.Point) error {
	bps, err := client.NewBatchPoints(client.BatchPointsConfig{
		Database: m.influxDBDatabase,
	})
	if err != nil {
		return err
	}

	bps.AddPoints(pts)

	err = m.influxDBClient.Write(bps)
	if err != nil {
		return err
	}

	return nil
}

// WriteMetrics write metric
func (m *MetricClient) WriteMetrics(metrics []*Metric) error {
	pts := make([]*client.Point, len(metrics))
	for i, metric := range metrics {
		pt, err := client.NewPoint(MeasurementNameProcess,
			map[string]string{
				TagNameFleetID:        metric.FleetID,
				TagNameScalingGroupID: metric.ScalingGroupID,
				TagNameInstanceID:     metric.InstanceID,
				TagNameProcessID:      metric.ID,
			},
			map[string]interface{}{
				FieldNameServerSessionCount:  metric.ServerSessionCount,
				FieldNameMaxServerSessionNum: metric.MaxServerSessionNum,
			},
		)
		if err != nil {
			return err
		}
		pts[i] = pt

	}
	return m.writeMetrics(pts)
}
