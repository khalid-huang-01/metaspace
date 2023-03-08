// Copyright (c) Huawei Technologies Co., Ltd. 2022-2022. All rights reserved.

// 清理数据库数据
package clean

import (
	"codehub-g.huawei.com/videocloud/mediaprocesscenter/application-gateway/config"
	"codehub-g.huawei.com/videocloud/mediaprocesscenter/application-gateway/pkg/models"
	app_process "codehub-g.huawei.com/videocloud/mediaprocesscenter/application-gateway/pkg/models/appprocess"
	client_session "codehub-g.huawei.com/videocloud/mediaprocesscenter/application-gateway/pkg/models/clientsession"
	server_session "codehub-g.huawei.com/videocloud/mediaprocesscenter/application-gateway/pkg/models/serversession"
	"codehub-g.huawei.com/videocloud/mediaprocesscenter/application-gateway/pkg/utils/log"
	"fmt"
	"time"
)

const (
	DataBaseInitTime = 10
)

// Init 用于定期软删除数据库中数据
func Init() {
	if config.GlobalConfig.CleanStrategy == "on" || config.GlobalConfig.CleanStrategy == "ON" {
		go InitDataBase(DataBaseInitTime * time.Second)
	}
}

func InitDataBase(t time.Duration) {
	clientSessionDao := client_session.NewClientSessionDao(models.MySqlOrm)
	serverSessionDao := server_session.NewServerSessionDao(models.MySqlOrm)
	appProcessDao := app_process.NewAppProcessDao(models.MySqlOrm)
	ticker := time.Tick(t)
	for i := range ticker {
		fmt.Println(i)
		go func() {
			err := clientSessionDao.CleanClientSession()
			if err != nil {
				log.RunLogger.Errorf("clear useless server session error for %v", err)
			}
		}()
		go func() {
			err := serverSessionDao.CleanServerSession()
			if err != nil {
				log.RunLogger.Errorf("clear useless server session error for %v", err)
			}
		}()
		go func() {
			err := appProcessDao.CleanAppProcess()
			if err != nil {
				log.RunLogger.Errorf("clear useless client session error for %v", err)
			}
		}()
	}
}
