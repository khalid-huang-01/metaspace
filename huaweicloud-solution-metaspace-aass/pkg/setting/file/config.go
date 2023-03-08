// Copyright (c) Huawei Technologies Co., Ltd. 2022-2022. All rights reserved.

// 服务配置
package file

import (
	"encoding/json"
	"io/ioutil"

	"scase.io/application-auto-scaling-service/pkg/utils/config"
)

type Config struct {
	*config.Config
}

// NewConfig get new Config
func NewConfig(cfgFile string) (*Config, error) {
	c := &Config{}
	c.Config = config.NewConfig(nil)

	if err := c.init(cfgFile); err != nil {
		return nil, err
	}

	if err := c.monitor(cfgFile); err != nil {
		return nil, err
	}

	return c, nil
}

func (c *Config) init(cfgFile string) error {
	data, err := ioutil.ReadFile(cfgFile)
	if err != nil {
		return err
	}
	cfgMap := make(map[string]interface{})
	if err = json.Unmarshal(data, &cfgMap); err != nil {
		return err
	}
	c.ReNew(cfgMap)
	return nil
}

func (c *Config) monitor(cfgFile string) error {
	if err := NewMonitor(cfgFile, c.init); err != nil {
		return err
	}
	MonitorStart()
	return nil
}
