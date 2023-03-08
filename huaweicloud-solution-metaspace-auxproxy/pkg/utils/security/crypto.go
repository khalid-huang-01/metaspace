// Copyright (c) Huawei Technologies Co., Ltd. 2022-2022. All rights reserved.

// 安全模块初始化
package security

import (
	"encoding/base64"
	"crypto/aes"
	"crypto/cipher"
	"fmt"
)
const (
	GCMKey 	 		= "mTqBnYoyhQm8xOkRFkaUU7X2"
	GCMNonce 		= "0xCWIOTotifzJMnD"	
)

// AES-GCM模式加密，密文使用base64编码输出
func GCM_Encrypt(plaintextStr string, key string, nonce string) (string, error) {
	// 将明文和密钥转换为字节切片
	plaintext := []byte(plaintextStr)
	keyByte := []byte(key)
	// 创建加密分组
	block, err := aes.NewCipher(keyByte)
	if err != nil {
		return "", fmt.Errorf("key 长度必须 16/24/32长度: %s", err.Error())
	}

	// 创建 GCM 模式的 AEAD
	aesgcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}
	// 生成密文
	nonceByte, err := base64.RawURLEncoding.DecodeString(nonce)
	if err != nil {
		return "", fmt.Errorf("decode nonce error: %s", err.Error())
	}
	ciphertext := aesgcm.Seal(nil, nonceByte, plaintext, nil)
	// 返回密文及随机数的 base64 编码
	return base64.RawURLEncoding.EncodeToString(ciphertext), nil
}

// gcm模式解密，解密使用base64编码的密文
func GCM_Decrypt(ciphertextStr string, key string, nonce string) (string, error) {
	// 将密文,密钥和生成的随机数转换为字节切片
	ciphertext, err := base64.RawURLEncoding.DecodeString(ciphertextStr)
	if err != nil {
		return "", err
	}
	nonceByte, err := base64.RawURLEncoding.DecodeString(nonce)
	if err != nil {
		return "", err
	}
	keyByte := []byte(key)

	// 创建加密分组
	block, err := aes.NewCipher(keyByte)
	if err != nil {
		return "", err
	}
	// 创建 GCM 模式的 AEAD
	aesgcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}
	// 明文内容
	plaintext, err := aesgcm.Open(nil, nonceByte, ciphertext, nil)
	if err != nil {
		return "", err
	}
	return string(plaintext), nil
}
