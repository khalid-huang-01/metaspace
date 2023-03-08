// Copyright (c) Huawei Technologies Co., Ltd. 2022-2022. All rights reserved.

// hmac
package hhmac

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"time"

	"github.com/pkg/errors"

	"codehub-g.huawei.com/videocloud/mediaprocesscenter/auxproxy/config"
	"codehub-g.huawei.com/videocloud/mediaprocesscenter/auxproxy/pkg/utils/log"
)

const (
	timestampFormat = "Mon, 02 Jan 2006 15:04:05 -0700"
	prefix          = "SCASE-HMAC.V1"
	hmacStringLen   = 2
	// defaultHmacValidPeriod 默认加密信息的有效时长
	defaultHmacValidPeriod = 600 * time.Second
)

// RequestSignHmac 发起请求前进行hmac加密操作，会将加密信息加入req的header
// 注意：此方法 err 需要做脱敏处理
func RequestSignHmac(req *http.Request, ak string, sk []byte) error {
	if req == nil {
		return errors.New("param req is nil")
	}

	timestamp := time.Now().UTC().Format(timestampFormat)
	toSign, err := genStrToSign(req, timestamp)
	if err != nil {
		return err
	}
	hmc := Hmac{
		AK:        ak,
		SK:        sk,
		ToSign:    toSign,
		Timestamp: timestamp,
	}
	hi := hmc.sign()

	req.Header.Add("Date", hi.Date)
	req.Header.Add("Authorization", hi.Authorization)
	return nil
}

// ValidateHmac 服务端接收到请求后，对req进行hmac身份认证
// 认证方式：会和客户端对req数据进行相同的加密操作，对比加密摘要是否相同
// 注意：此方法 err 需要做脱敏处理
func ValidateHmac(req *http.Request) error {
	h, err := decodeHmac(req)
	if err != nil {
		return err
	}

	// 判断摘要是否相同
	signature := hmacSign(h.ToSign, h.SK)
	fmt.Println("signature: ", signature)
	fmt.Println("digest", h.Digest)
	if signature != h.Digest {
		return errors.New("signature not equal")
	}

	// 签名超时校验
	tm, err := time.Parse(timestampFormat, h.Timestamp)
	if err != nil {
		return err
	}
	diff := time.Now().Sub(tm)
	expireDur := time.Duration(config.ServerHmacConf.ExpireSeconds) * time.Second
	if expireDur.Seconds() == 0 {
		expireDur = defaultHmacValidPeriod
	}
	if diff > expireDur || diff < -expireDur {
		return errors.Errorf("timstamp diverse more than %v second(s)",
			config.ServerHmacConf.ExpireSeconds)
	}
	return nil
}

// 解析req的hmac信息
func decodeHmac(req *http.Request) (*Hmac, error) {
	h := Hmac{
		Timestamp: req.Header.Get("Date"),
	}
	ak, digest, err := h.decodeAuthorization(req.Header.Get("Authorization"))
	if err != nil {
		return nil, err
	}

	fmt.Println("ak: ", ak)
	fmt.Println("digest: ", digest)
	// 通过 ak 获取对应的 sk
	h.AK = ak
	sk, err := HmacStoreSker.GetSk(ak)
	if err != nil {
		return nil, err
	}
	fmt.Println("sk: ", string(sk))
	h.SK = sk
	h.Digest = digest
	toSign, err := genStrToSign(req, h.Timestamp)
	if err != nil {
		return nil, err
	}
	h.ToSign = toSign
	return &h, nil
}

func genStrToSign(req *http.Request, timestamp string) (string, error) {
	if req == nil {
		return "", errors.New("param req is nil")
	}
	var body []byte
	var err error

	if req.Body != nil {
		if body, err = ioutil.ReadAll(req.Body); err != nil {
			return "", errors.Wrap(err, "read req body err")
		}
		// restore body
		req.Body = ioutil.NopCloser(bytes.NewBuffer(body))
	}

	hash := sha256.New()
	if _, err = hash.Write(body); err != nil {
		// 安全考虑，不输出具体err信息
		return "", errors.New("write req body to hash err")
	}
	contentSHA2 := hash.Sum(nil)
	httpVerb := req.Method
	contentType := req.Header.Get("Content-Type")
	toSign := httpVerb + "\n" + base64.StdEncoding.EncodeToString(contentSHA2) + "\n" +
		contentType + "\n" + timestamp
	return toSign, nil
}

type Hmac struct {
	// 秘钥信息
	AK string
	SK []byte
	// 加密后的摘要信息
	Digest string
	// 加密内容
	ToSign string
	// 加密时间
	Timestamp string
}

type HmacInfo struct {
	Date          string
	Authorization string
}

func (h *Hmac) sign() *HmacInfo {
	h.Digest = hmacSign(h.ToSign, h.SK)

	return &HmacInfo{
		Date:          h.Timestamp,
		Authorization: h.encodeAuthorization(),
	}
}

// Authorization = "SCASE-HMAC.V1" + " " + AccessKeyId + ":" + Signature
func (h *Hmac) encodeAuthorization() string {
	return fmt.Sprintf("%s %s:%s", prefix, h.AK, h.Digest)
}

// Authorization = "SCASE-HMAC.V1" + " " + AccessKeyId + ":" + Signature
func (h *Hmac) decodeAuthorization(auth string) (string, string, error) {
	ss := strings.Split(auth, " ")
	if len(ss) != hmacStringLen {
		return "", "", errors.New("format error")
	}
	if ss[0] != prefix {
		return "", "", errors.New("format header not matched")
	}

	ss2 := strings.Split(ss[1], ":")
	if len(ss2) != hmacStringLen {
		return "", "", errors.New("format error")
	}

	return ss2[0], ss2[1], nil
}

// hmac 加密操作
// Base64( HMAC-SHA2( YourSecretAccessKeyID, UTF-8-Encoding-Of( StringToSign ) ) )
func hmacSign(data string, secret []byte) string {
	h := hmac.New(sha256.New, secret)
	_, err := h.Write([]byte(data))
	if err != nil {
		log.RunLogger.Errorf("Write data to hash err")
		return ""
	}
	return base64.StdEncoding.EncodeToString(h.Sum(nil))
}
