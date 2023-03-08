// Copyright (c) Huawei Technologies Co., Ltd. 2022-2022. All rights reserved.

// 客户端请求服务定义
package client

import (
	"bytes"
	"fleetmanager/logger"
	"fleetmanager/security"
	"fleetmanager/setting"
	"io/ioutil"
	"net/http"
	"strings"
	"time"
)

const (
	BadRequestCode    = 400
	NormalRequestCode = 200
)

type Request struct {
	logger     *logger.FMLogger
	req        *http.Request
	rsp        *http.Response
	client     *http.Client
	queries    map[string]string
	service    string
	method     string
	url        string
	reqBody    []byte
	rspBody    []byte
	startTime  time.Time
	endTime    time.Time
	enableHmac bool
	hmacAK     []byte
	hmacSK     []byte
}

func (r *Request) WriteCallLog(err error) {
	r.endTime = time.Now()
	fields := make(map[string]interface{}, 8)
	fields[logger.StartTime] = r.startTime.Unix()
	fields[logger.EndTime] = r.endTime.Unix()
	fields[logger.DurationMs] = r.endTime.Sub(r.startTime).Milliseconds()
	fields[logger.ServiceName] = r.service
	fields[logger.RequestMethod] = r.method
	fields[logger.RequestBody] = string(r.reqBody)
	fields[logger.ResponseBody] = string(r.rspBody)
	fields[logger.RequestRawUri] = r.url
	if r.rsp != nil {
		fields[logger.ResponseCode] = r.rsp.StatusCode
		if r.rsp.StatusCode >= NormalRequestCode && r.rsp.StatusCode < BadRequestCode {
			fields[logger.ResponseStatus] = 1
		} else {
			fields[logger.ResponseStatus] = 0
		}
	}
	if err != nil {
		fields[logger.Error] = err.Error()
	}

	r.logger.WithFields(fields).Info("service call event")
}

func (r *Request) SetHmacConf(enableHmac bool, ak []byte, sk []byte) {
	r.enableHmac = enableHmac
	r.hmacAK = ak
	r.hmacSK = sk
}

// 内部服务使用hmac签名
func (r *Request) DoAuth() {
	// 调用aass/appgateway前进行hmac签名
	if r.enableHmac {
		err := security.RequestSignHmac(r.req, r.hmacAK, r.hmacSK)
		if err != nil {
			logger.R.Error("Req sign hmac err: %+v", err)
			return
		}
	}
}

func (r *Request) SetLogger(l *logger.FMLogger) {
	r.logger = l
}

// SetAuthToken 添加X-Auth-Token
func (r *Request) SetAuthToken(token []byte) {
	r.req.Header.Set(AuthToken, string(token))
}

// SetHeader 添加Header
func (r *Request) SetHeader(header map[string]string) {
	if header == nil {
		return
	}
	for k, v := range header {
		r.req.Header.Set(k, v)
	}
}

// SetSubjectToken 添加Subject-Auth-Token
func (r *Request) SetSubjectToken(token []byte) {
	r.req.Header.Set(SubjectToken, string(token))
}

// SetContentType 添加Content-Type
func (r *Request) SetContentType(ct string) {
	r.req.Header.Set(ContentType, ct)
}

// SetQuery 设置查询
func (r *Request) SetQuery(k string, v string) {
	if len(k) > 0 && len(v) > 0 {
		r.queries[k] = v
	}
}

func (r *Request) QueryString() string {
	if len(r.queries) == 0 {
		return ""
	}
	qs := make([]string, 0)
	for k, v := range r.queries {
		qs = append(qs, k+"="+v)
	}
	return "?" + strings.Join(qs, "&")
}

func (r *Request) InitRequest() error {
	var err error
	
	r.req, err = http.NewRequest(r.method, r.url, bytes.NewReader(r.reqBody))

	return err
}

// GetHeader 获取请求header
func (r *Request) GetHeader() http.Header {
	return r.rsp.Header
}

// DoRequest执行请求
func (r *Request) DoRequest() (code int, rsp []byte, err error) {
	defer func() { r.WriteCallLog(err) }()
	
	// 增加query请求
	query := r.req.URL.Query()
	for key, value := range r.queries {
		query.Add(key, value)
	}
	r.req.URL.RawQuery = query.Encode()

	r.startTime = time.Now()
	r.DoAuth()

	r.rsp, err = r.client.Do(r.req)
	if err != nil {
		return code, rsp, err
	}

	if r.rsp == nil {
		return code, rsp, nil
	}

	code = r.rsp.StatusCode
	if r.rsp.Body != nil {
		r.rspBody, err = ioutil.ReadAll(r.rsp.Body)
		if err != nil {
			return code, rsp, err
		}
		defer func() {
			e := r.rsp.Body.Close()
			if e != nil {
				logger.R.Error("close rsp error: %v, try to ignore", e)
			}
		}()
		if len(r.rspBody) > 0 {
			rsp = make([]byte, len(r.rspBody))
			copy(rsp, r.rspBody)
			return code, rsp, nil
		}
	}

	return code, rsp, nil
}

func NewRequest(service string, url string, method string, body []byte) IRequest {
	r := &Request{
		service: service,
		url:     url,
		method:  method,
		reqBody: body,
		client:  httpsClient,
		logger:  logger.C,
		queries: make(map[string]string),
	}
	err := r.InitRequest()
	if err != nil {
		return nil
	}
	// 基于service名称增加hmac认证
	if service == ServiceNameAASS {
		r.SetHmacConf(setting.AASSEnableHmac, setting.AASSHmacAK, setting.AASSHmacSK)
	} else if service == ServiceNameAPPGW {
		r.SetHmacConf(setting.AppGatewayEnableHmac, setting.AppGatewayHmacAK, setting.AppGatewayHmacSK)
	}

	return r
}
