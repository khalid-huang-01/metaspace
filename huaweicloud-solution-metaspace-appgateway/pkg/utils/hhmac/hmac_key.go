// Copyright (c) Huawei Technologies Co., Ltd. 2022-2022. All rights reserved.

// 加密密匙
package hhmac

import (
	"github.com/pkg/errors"
	"codehub-g.huawei.com/videocloud/mediaprocesscenter/application-gateway/pkg/utils/security"
	"codehub-g.huawei.com/videocloud/mediaprocesscenter/application-gateway/config"
)

var (
	// HmacStoreSker 远端 访问 本地 时的认证秘钥存储器
	HmacStoreSker StoreSker
	// HmacLocalSker 本地 访问 远端 时的认证秘钥存储器
	HmacLocalSker LocalSker
)

const (
	LocalKeyAASS     = "aass"
	LocalKeyAuxProxy = "auxproxy"
)

// InitHMACKey 初始化 hmac 秘钥存储器
func InitHMACKey() error {
	var err error

	// 服务端 hmac 身份认证，秘钥信息初始化
	if config.ServerHmacConf.AuthEnable {
		if HmacStoreSker, err = newHmacKeyStore(); err != nil {
			return err
		}
	}

	// 客户端 hmac 签名，秘钥信息初始化
	if HmacLocalSker, err = newHmacLocalKey(); err != nil {
		return err
	}
	return nil
}

func newHmacKeyStore() (StoreSker, error) {
	store := NewStore()
	storeKeys := config.ServerHmacConf.Keys

	for _, entry := range storeKeys {
		sk, err := security.GCM_Decrypt(string(entry.SKCypher), config.GlobalConfig.GCMKey, config.GlobalConfig.GCMNonce)
		if err != nil {
			return nil, errors.New("decrypt err")
		}
		store.Add(entry.AK, KeyEntry{
			SK: []byte(sk),
		})
	}

	return store, nil
}

func newHmacLocalKey() (LocalSker, error) {
	localSker := NewlocalKey()
	localKeyEntrys := config.ClientHmacConf.Keys

	// eg: 通过如下方式，加入访问远端服务的 hmac秘钥 信息
	aassKey := localKeyEntrys[LocalKeyAASS]
	if aassKey.Enable {
		sk, err := security.GCM_Decrypt(string(aassKey.SKCypher), config.GlobalConfig.GCMKey, config.GlobalConfig.GCMNonce)
		if err != nil {
			return nil, errors.New("decrypt err")
		}
		localSker.Add(LocalKeyAASS, LocalKeyEntry{
			Enable: aassKey.Enable,
			AK:     aassKey.AK,
			SK:     []byte(sk),
		})
	}

	auxProxyKey := localKeyEntrys[LocalKeyAuxProxy]
	if auxProxyKey.Enable {
		sk, err := security.GCM_Decrypt(string(auxProxyKey.SKCypher), config.GlobalConfig.GCMKey, config.GlobalConfig.GCMNonce)
		if err != nil {
			return nil, errors.New("decrypt err")
		}
		localSker.Add(LocalKeyAuxProxy, LocalKeyEntry{
			Enable: auxProxyKey.Enable,
			AK:     auxProxyKey.AK,
			SK:     []byte(sk),
		})
	}

	return localSker, nil
}
