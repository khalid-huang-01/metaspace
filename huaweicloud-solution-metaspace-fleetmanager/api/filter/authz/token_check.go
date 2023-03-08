// Copyright (c) Huawei Technologies Co., Ltd. 2022-2022. All rights reserved.

// token认证模块
package authz

import (
	"go.mozilla.org/pkcs7"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"encoding/pem"
	"fleetmanager/api/errors"
	"fleetmanager/api/filter/authz/token"
	"fleetmanager/api/filter/authz/user"
	"fleetmanager/api/filter/authz/util"
	"fleetmanager/setting"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"sync"
	"time"
)

const (
	// URLGetSignCert get sign Certificate uri
	URLGetSignCert = "/v3/OS-SIMPLE-CERT/certificates"
	// HeaderParameterAccept is the request's header title key Accept
	HeaderParameterAccept = "Accept"
	// HeaderContentApplicationJSON is the request's header title value application/json
	HeaderContentApplicationJSON = "application/json;charset=utf8"
)

var safeMap = struct {
	sync.RWMutex
	signCerts map[string]*x509.Certificate
}{signCerts: make(map[string]*x509.Certificate)}

var (
	mutex sync.Mutex
)

func getInternalIamEndpoint() string {
	return setting.Config.Get(setting.InternalIamEndpoint).ToString("")
}

// AuthenticateToken: Token认证
func AuthenticateToken(tokenString []byte) (user.Info, error, errors.ErrCode) {
	if string(tokenString) == "" {
		return nil, fmt.Errorf("AuthenticateToken, token not found"), errors.TokenNotFound
	}

	// 替换符号"-"为"/"
	// 把上面一整行字符串切分为每行64个字符的长度
	// 添加头"-----BEGIN CMS-----"，添加尾"-----END CMS-----"
	decodeToken, err := base64.StdEncoding.DecodeString(strings.Replace(string(tokenString), "-", "/", -1))
	if err != nil {
		return nil, fmt.Errorf("AuthenticateToken, Decode Base64: %v", err), errors.TokenPermissionFailed
	}

	p7, err := pkcs7.Parse(decodeToken)
	if err != nil {
		return nil, fmt.Errorf("AuthenticateToken, pkcs7.Parse: %v", err), errors.TokenPermissionFailed
	}

	if len(p7.Signers) != 1 {
		return nil, fmt.Errorf("AuthenticateToken, should be only one signature found"),
			errors.TokenPermissionFailed
	}

	// 通过SerialNumber匹配缓存中对应的证书，如果没有匹配到从IAM重新获取证书
	serialNumber := p7.Signers[0].IssuerAndSerialNumber.SerialNumber.String()
	cert, err := getMatchCert(serialNumber, true)
	if err != nil {
		return nil, err, errors.TokenPermissionFailed
	}

	err = cert.CheckSignature(x509.SHA256WithRSA, p7.Content, p7.Signers[0].EncryptedDigest)
	if err != nil {
		return nil, fmt.Errorf("AuthenticateToken, CheckSignature error: %v, sn is: %v", err, serialNumber),
			errors.TokenPermissionFailed
	}

	var parsedToken token.Token
	err = json.Unmarshal(p7.Content, &parsedToken)
	if err != nil {
		return nil, fmt.Errorf("AuthenticateToken, Unmarshal token error: %v", err),
			errors.TokenPermissionFailed
	}

	expiresAt := parsedToken.TokenInfo.ExpiresAt.UTC()
	if time.Now().UTC().After(expiresAt) {
		return nil, fmt.Errorf("Token expired!"), errors.TokenExpired
	}

	info := userInfo(parsedToken)
	return info, nil, errors.NoError
}

func getMatchCert(serialNumber string, refresh bool) (*x509.Certificate, error) {
	safeMap.RLock()
	cert, exist := safeMap.signCerts[serialNumber]
	safeMap.RUnlock()
	if !exist {
		if refresh {
			mutex.Lock()
			// 实际应调用IAM获取证书接口重新获取token
			err := GetSignCertFromIAM()
			mutex.Unlock()
			if err != nil {
				return nil, err
			}
			return getMatchCert(serialNumber, false)
		} else {
			return nil, fmt.Errorf("AuthenticateToken, no cert match token's signer")
		}
	}
	return cert, nil
}

// GetSignCertFromIAM: 获取签名证书
func GetSignCertFromIAM() error {
	req, err := http.NewRequest(http.MethodGet, getInternalIamEndpoint()+URLGetSignCert, nil)
	if err != nil {
		return err
	}
	req.Header.Add(HeaderParameterAccept, HeaderContentApplicationJSON)
	resp, err := util.GetDefaultHttpClient().Do(req)
	if err != nil {
		return err
	}

	defer func() { resp.Body.Close() }()
	respBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to get sign cert, StatusCode: %d, Message: %s", resp.StatusCode, respBody)
	}

	for len(respBody) != 0 {
		var derBlock *pem.Block
		derBlock, respBody = pem.Decode(respBody)
		if derBlock == nil {
			break
		}
		signCert, err := x509.ParseCertificate(derBlock.Bytes)
		if err != nil {
			return err
		}

		safeMap.Lock()
		safeMap.signCerts[signCert.SerialNumber.String()] = signCert
		safeMap.Unlock()
	}

	return nil
}

func userInfo(parsedToken token.Token) user.Info {
	ret := &user.DefaultUser{
		Name:        parsedToken.TokenInfo.User.Name,
		UID:         parsedToken.TokenInfo.User.ID,
		DomainID:    parsedToken.TokenInfo.User.Domain.ID,
		DomainName:  parsedToken.TokenInfo.User.Domain.Name,
		XDomainID:   parsedToken.TokenInfo.User.Domain.XDomainID,
		XDomainType: parsedToken.TokenInfo.User.Domain.XDomainType,
		ProjectID:   parsedToken.TokenInfo.Project.ID,
		ProjectName: parsedToken.TokenInfo.Project.Name,
		Roles:       parsedToken.TokenInfo.Roles,
		Extra:       make(map[string]string),
	}
	return ret
}
