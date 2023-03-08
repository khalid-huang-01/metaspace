package downloader

import (
	"codehub-g.huawei.com/videocloud/mediaprocesscenter/auxproxy/pkg/utils/log"
	"huaweicloud.com/esdk-obs-go/obs/v3"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
)

type Downloader struct {
	BucketName string
	ObjectKey  string
	Location   string
	ObsClient  *obs.ObsClient
}

func NewDownloader(ak, sk, endpoint, bucketName, objectKey, location string) (*Downloader, error) {
	obsClient, err := obs.New(ak, sk, endpoint)
	if err != nil {
		log.RunLogger.Infof("[downloader] NewDownloader failed, err:%v", err)
		return nil, err
	}
	return &Downloader{ObsClient: obsClient, BucketName: bucketName, ObjectKey: objectKey, Location: location}, nil
}

func (d Downloader) CreateBucket() error {
	input := &obs.CreateBucketInput{}
	input.Bucket = d.BucketName
	input.Location = d.Location
	_, err := d.ObsClient.CreateBucket(input)
	if err != nil {
		log.RunLogger.Infof("[downloader] CreateBucket failed, err:%v", err)
		return err
	}
	log.RunLogger.Infof("[downloader] CreateBucket success")
	return nil
}

func (d Downloader) PutObject() error {
	input := &obs.PutObjectInput{}
	input.Bucket = d.BucketName
	input.Key = d.ObjectKey
	input.Body = strings.NewReader("Hello OBS")

	_, err := d.ObsClient.PutObject(input)
	if err != nil {
		log.RunLogger.Infof("[downloader] PutObject failed, err:%v", err)
		return err
	}
	log.RunLogger.Infof("[downloader] PutObject success")
	return nil
}

func (d Downloader) GetObject() error {
	input := &obs.GetObjectInput{}
	input.Bucket = d.BucketName
	input.Key = d.ObjectKey

	output, err := d.ObsClient.GetObject(input)
	if err != nil {
		log.RunLogger.Infof("[downloader] GetObject failed, err:%v", err)
		return err
	}
	defer func() {
		errMsg := output.Body.Close()
		if errMsg != nil {
			log.RunLogger.Infof("[downloader] GetObject failed, err:%v", errMsg)
		}
	}()
	body, err := ioutil.ReadAll(output.Body)
	if err != nil {
		log.RunLogger.Infof("[downloader] GetObject failed, err:%v", err)
		return err
	}
	log.RunLogger.Infof("[downloader] GetObject success, body:%v", body)
	return nil
}

func (d Downloader) PutFile(filePath string) error {
	input := &obs.PutFileInput{}
	input.Bucket = d.BucketName
	input.Key = d.ObjectKey
	input.SourceFile = filePath

	_, err := d.ObsClient.PutFile(input)
	if err != nil {
		log.RunLogger.Infof("[downloader] PutFile failed, err:%v", err)
		return err
	}
	log.RunLogger.Infof("[downloader] PutFile success")
	return nil
}

func (d Downloader) DeleteObject() error {
	input := &obs.DeleteObjectInput{}
	input.Bucket = d.BucketName
	input.Key = d.ObjectKey

	_, err := d.ObsClient.DeleteObject(input)
	if err != nil {
		log.RunLogger.Infof("[downloader] DeleteObject failed, err:%v", err)
		return err
	}
	log.RunLogger.Infof("[downloader] DeleteObject success")
	return nil
}

func (Downloader) CreateFile(filePath string) error {
	if err := os.MkdirAll(filepath.Dir(filePath), os.ModePerm); err != nil {
		log.RunLogger.Infof("[downloader] CreateFile failed, err:%v", err)
		return err
	}

	if err := ioutil.WriteFile(filePath, []byte("Hello OBS from file"), os.ModePerm); err != nil {
		log.RunLogger.Infof("[downloader] CreateFile failed, err:%v", err)
		return err
	}
	return nil
}
