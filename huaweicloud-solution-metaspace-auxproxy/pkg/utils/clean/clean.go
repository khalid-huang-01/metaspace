// Copyright (c) Huawei Technologies Co., Ltd. 2022-2022. All rights reserved.

// 日志清理
package clean

import (
	"os"
	"path"
	"strings"

	"codehub-g.huawei.com/videocloud/mediaprocesscenter/auxproxy/pkg/utils/log"
)

var AbsPathPrefix = "/local/app/"

// DeleteLogFiles delete log files
func DeleteLogFiles(logFiles []string) error {
	for _, logFile := range logFiles {
		if !ValidLogFile(logFile) {
			log.RunLogger.Errorf("[clean] invalid log file path %s, deny to delete it", logFile)
			continue
		}
		err := os.RemoveAll(logFile)
		if err != nil {
			return err
		}
	}
	return nil
}

// ValidLogFile validate log file
func ValidLogFile(logFile string) bool {
	switch path.IsAbs(logFile) {
	case true:
		if strings.HasPrefix(logFile, AbsPathPrefix) {
			return true
		}
	case false:
	}
	return false
}
