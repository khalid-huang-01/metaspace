// Copyright (c) Huawei Technologies Co., Ltd. 2012-2018. All rights reserved.

package buildmanager

import (
	"archive/zip"
	"codehub-g.huawei.com/videocloud/mediaprocesscenter/auxproxy/config"
	"codehub-g.huawei.com/videocloud/mediaprocesscenter/auxproxy/pkg/configmanager"
	"codehub-g.huawei.com/videocloud/mediaprocesscenter/auxproxy/pkg/utils/clients"
	"codehub-g.huawei.com/videocloud/mediaprocesscenter/auxproxy/pkg/utils/downloader"
	"codehub-g.huawei.com/videocloud/mediaprocesscenter/auxproxy/pkg/utils/log"
	"fmt"
	"huaweicloud.com/esdk-obs-go/obs/v3"
	"io"
	"net/http"
	"os"
	"path"
	"time"
)

type BuildInfo struct {
	BuildID              string
	BuildPath            string
	DownloadPath         string
	FileName             string
	Bucket               string
	ObjectKey            string
	GlobalServiceAddress string
	Location             string
}

type ReportRequest struct {
	BuildID string `json:"build_id"`
	Region  string `json:"region"`
	Bucket  string `json:"bucket"`
	Object  string `json:"object"`
	Result  int    `json:"result"` // 0-成功，非0失败，在state描述原因
	State   string `json:"state"`  // Complete/No Auth/NoBucket/Download failed...
}

const (
	BuildHandleSuccess = 0
	BuildHandleFail    = 1

	// 对应ReportRequest的State字段
	Complete       = "Complete"
	ReasonNoAuth   = "No Auth"
	DownloadFailed = "Download failed"

	// MaxFileNum codeCheck要求检查待解压文件的数量，门限先设置的大一些
	MaxFileNum = 1024 * 1024
)

func Init() error {
	if !config.Opts.EnableBuild {
		return nil
	}
	log.RunLogger.Errorf("[main] start init build manager.")

	var err error
	// 获取MetaData
	meta := &configmanager.Meta{}
	meta, err = configmanager.GetMetaDataConfig()
	if err != nil {
		log.RunLogger.Infof("[build manager] GetMetaDataConfig failed, err:%v", err)
	}
	build := &BuildInfo{
		BuildID:              meta.BuildID,
		BuildPath:            config.BuildPathPrefix,
		DownloadPath:         config.DownloadPathPrefix,
		FileName:             "",
		Bucket:               meta.Bucket,
		ObjectKey:            meta.Object,
		GlobalServiceAddress: meta.GlobalServiceAddress,
		Location:             meta.Region,
	}

	// 创建下载目录和应用目录
	err = CreateBuildDocument(build)
	if err != nil {
		log.RunLogger.Infof("[build manager] CreateBuildDocument failed, err:%v", err)
		return err
	}

	// 下载obs压缩包
	err = DownloadBuild(build, meta.Ak, meta.Sk)
	if err != nil {
		log.RunLogger.Infof("[build manager] DownloadBuild failed, err:%v", err)
		return err
	}

	// 检查压缩包是否生成完成，未完成则再等待一会儿
	zipReady := checkZipReady(build)
	if !zipReady {
		log.RunLogger.Infof("[build manager] checkZipReady failed, build:%v", build.BuildID)
		return fmt.Errorf("check zip not ready")
	}

	// 解压应用包到运行路径下
	err = UnzipBuild(build)
	if err != nil {
		log.RunLogger.Infof("[build manager] UnzipBuild failed, err:%v", err)
		return err
	}

	// 回调解压结果
	err = sendBuildStateToFleetManager(build, BuildHandleSuccess, Complete)
	if err != nil {
		log.RunLogger.Infof("[build manager] sendBuildStateToFleetManager failed, err:%v", err)
		return err
	}
	return nil
}

func checkZipReady(build *BuildInfo) bool {
	// obs接口返回时可能zip文件还未完全完成，1s轮询等待一下文件生成，最长10s
	timeout := time.After(time.Second * 10)
	for {
		select {
		case <-timeout:
			log.RunLogger.Infof("[build manager] checkZipReady failed, time out")
			return false
		default:
			_, err := os.Stat(build.DownloadPath + "/" + build.FileName)
			if err == nil {
				log.RunLogger.Infof("[build manager] checkZipReady success")
				return true
			}
		}
		time.Sleep(time.Second * 1)
	}
}

func sendBuildStateToFleetManager(build *BuildInfo, result int, state string) error {
	cli := clients.NewHttpsClientWithoutCerts()
	reqBody := &ReportRequest{
		BuildID: build.BuildID,
		Region:  build.Location,
		Bucket:  build.Bucket,
		Object:  build.ObjectKey,
		Result:  result,
		State:   state,
	}

	req, err := clients.JSONEncodeRequest(http.MethodPost,
		fmt.Sprintf("%s/v1/build/%s/state", build.GlobalServiceAddress, build.BuildID), reqBody)
	if err != nil {
		return err
	}

	code, _, _, err1 := clients.DoRequest(cli, req)
	if err1 != nil || code != http.StatusOK {
		return fmt.Errorf("code %d or err %v", code, err1)
	}

	log.RunLogger.Errorf("[build manager] sendBuildStateToFleetManager success, build:%v, state:%v", build, state)
	return nil
}

func DownloadBuild(build *BuildInfo, ak, sk string) error {
	_, build.FileName = path.Split(build.ObjectKey)
	d, err := downloader.NewDownloader(ak, sk, GetOBSEndpoint(build.Location), build.Bucket, build.ObjectKey, build.Location)
	if err != nil {
		return err
	}

	input := &obs.DownloadFileInput{}
	input.Bucket = build.Bucket
	input.Key = build.ObjectKey
	// 下载对象的本地文件全路径
	input.DownloadFile = config.DownloadPathPrefix + "/" + build.BuildID + "/" + build.FileName
	// 开启断点续传模式
	input.EnableCheckpoint = true
	// 指定分段大小为9MB
	input.PartSize = 9 * 1024 * 1024
	// 指定分段下载时的最大并发数
	input.TaskNum = 5

	log.RunLogger.Infof("[build manager] DownloadBuild %v start, build:%+v", build.BuildID, build)
	output, err := d.ObsClient.DownloadFile(input)
	if err != nil {
		log.RunLogger.Errorf("[build manager] DownloadBuild %v failed, output:%v, err:%v",
			build.BuildID, output, err)
		return err
	}
	log.RunLogger.Infof("[build manager] DownloadBuild %v success, requestID:%v",
		build.BuildID, output.RequestId)
	return nil
}

func CreateBuildDocument(build *BuildInfo) error {
	var err error
	// 创建下载目录
	build.DownloadPath = config.DownloadPathPrefix + "/" + build.BuildID
	err = os.MkdirAll(build.DownloadPath, os.ModePerm)
	if err != nil {
		log.RunLogger.Infof("[build manager] create download document failed, err:%v, downloadPath:%v",
			err, build.DownloadPath)
		return err
	}

	// 创建应用运行目录
	build.BuildPath = config.BuildPathPrefix
	err = os.MkdirAll(build.BuildPath, os.ModePerm)
	if err != nil {
		log.RunLogger.Infof("[build manager] create build document failed, err:%v, buildPath:%v",
			err, build.BuildPath)
		return err
	}

	return nil
}

// UnzipBuild 解压应用文件到运行目录
func UnzipBuild(build *BuildInfo) error {
	zipReader, err := zip.OpenReader(build.DownloadPath + "/" + build.FileName)
	defer zipReader.Close()
	if err != nil {
		log.RunLogger.Infof("[build manager] unzip build file failed, err:%v, buildinfo:%v",
			err, build)
		return err
	}
	if len(zipReader.File) > MaxFileNum {
		log.RunLogger.Infof("[build manager] unzip build file failed, zip file num:%v is over limit:%v",
			len(zipReader.File), MaxFileNum)
		return fmt.Errorf("too many file will be unzip")
	}

	for _, file := range zipReader.File {
		// 如果是目录，则创建目录
		if file.FileInfo().IsDir() {
			if err = os.MkdirAll(build.BuildPath+"/"+file.Name, os.ModePerm); err != nil {
				return err
			}
			continue
		}

		fileReader, err := file.Open()
		if err != nil {
			return err
		}

		var f *os.File
		if f, err = os.OpenFile(build.BuildPath+"/"+file.Name, os.O_CREATE|os.O_RDWR|os.O_TRUNC, os.ModePerm); err != nil {
			return err
		}
		_, err = io.CopyN(f, fileReader, int64(file.UncompressedSize64))
		if err != nil {
			log.RunLogger.Infof("[build manager] UnzipBuild failed, io.CopyN err:%v", err)
			return err
		}
		err = f.Close()
		if err != nil {
			log.RunLogger.Infof("[build manager] UnzipBuild failed, f.Close err:%v", err)
			return err
		}
		err = fileReader.Close()
		if err != nil {
			log.RunLogger.Infof("[build manager] UnzipBuild failed, fileReader.Close err:%v", err)
			return err
		}
	}
	log.RunLogger.Infof("[build manager] UnzipBuild %v success", build.BuildID)
	return nil
}

// GetOBSEndpoint 乌兰三obs地址格式与其他region有区别，兼容一下测试环境
func GetOBSEndpoint(location string) string {
	if location == "cn-north-7" {
		return "https://obs.cn-north-7.ulanqab.huawei.com"
	}

	return "https://obs." + location + ".myhuaweicloud.com"
}
