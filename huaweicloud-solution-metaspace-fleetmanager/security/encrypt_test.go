// Copyright (c) Huawei Technologies Co., Ltd. 2022-2022. All rights reserved.

// 敏感信息加密解密操作测试模块
package security

import (
	"testing"
)

func TestGCMcrypt(t *testing.T) {
	key := "************"
	nonce := "***************"
	msg := "*************"
	cipher_text, err := GCM_Encrypt(msg, key, nonce)
	if err != nil {
		t.Errorf("encode msg %s err, %s", msg, err.Error())
		return
	}
	t.Logf("%s encode to %s", msg, cipher_text)
	plain_text, err := GCM_Decrypt(cipher_text, key, nonce)
	if err != nil {
		t.Errorf("decode msg %s err, %s\n", cipher_text, err.Error())
		return
	}
	t.Logf("%s decode to %s\n", cipher_text, plain_text)
}

func TestCTRcrypt(t *testing.T) {
	key := "****************"
	nonce := "***************"
	msg := "************"
	cipher_text, err := CTR_Encrypt(msg, key, nonce)
	if err != nil {
		t.Errorf("encode msg %s err, %s", msg, err.Error())
		return
	}
	t.Logf("%s encode to %s", msg, cipher_text)
	plain_text, err := CTR_Decrypt(cipher_text, key, nonce)
	if err != nil {
		t.Errorf("decode msg %s err, %s\n", cipher_text, err.Error())
		return
	}
	t.Logf("%s decode to %s\n", cipher_text, plain_text)
}

func TestRSAcrypt(t *testing.T) {
	msg := "********"
	baseDir := "*********************"
	cipher_text, err := RSA_Encrypt(msg, baseDir + "public.pem")
	if err != nil {
		t.Errorf("encode msg %s err, %s", msg, err.Error())
		return
	}
	t.Logf("%s encode to %s", msg, cipher_text)
	plain_text, err := RSA_Decrypt(cipher_text, baseDir + "private.pem")
	if err != nil {
		t.Errorf("decode msg %s err, %s", cipher_text, err.Error())
		return
	}
	t.Logf("%s decode to %s", cipher_text, plain_text)
}