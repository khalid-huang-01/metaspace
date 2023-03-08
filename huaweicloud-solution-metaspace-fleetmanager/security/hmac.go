// Copyright (c) Huawei Technologies Co., Ltd. 2022-2022. All rights reserved.

// hmac加解密
package security

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

	"fleetmanager/logger"
)

const (
	timestampFormat = "Mon, 02 Jan 2006 15:04:05 -0700"
	prefix          = "SCASE-HMAC.V1"
	lengthAuthString = 2
)

// RequestSignHmac 发起请求前进行hmac加密操作，会将加密信息加入req的header
func RequestSignHmac(req *http.Request, ak []byte, sk []byte) error {
	if req == nil {
		return errors.New("hmac sign param req is nil")
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

func genStrToSign(req *http.Request, timestamp string) (string, error) {
	var body []byte
	var err error

	if req.Body != nil {
		if body, err = ioutil.ReadAll(req.Body); err != nil {
			return "", errors.New("hmac sign read req body err")
		}
		// restore body
		req.Body = ioutil.NopCloser(bytes.NewBuffer(body))
	}

	hash := sha256.New()
	if _, err = hash.Write(body); err != nil {
		// 安全考虑，不输出具体err信息
		return "", errors.New("hmac sign write req body to hash err")
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
	AK []byte
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
	if len(ss) != lengthAuthString {
		return "", "", errors.New("format error")
	}
	if ss[0] != prefix {
		return "", "", errors.New("format header not matched")
	}

	ss2 := strings.Split(ss[1], ":")
	if len(ss2) != lengthAuthString {
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
		logger.R.Error("Write data to hash err")
		return ""
	}
	return base64.StdEncoding.EncodeToString(h.Sum(nil))
}
