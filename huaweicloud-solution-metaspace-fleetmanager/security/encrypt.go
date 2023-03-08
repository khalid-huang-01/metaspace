// Copyright (c) Huawei Technologies Co., Ltd. 2022-2022. All rights reserved.

// 敏感信息加密解密操作
package security

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/base64"
	"encoding/hex"
	"encoding/pem"
	"fmt"
	"os"
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

// AES-CTR模式加密，返回使用hex模式加密的密文
func CTR_Encrypt(plaintextStr string, key string, nonce string) (string, error) {
	// 转换成为字节切片
	plaintext := []byte(plaintextStr)
	keyByte := []byte(key)
	nonceByte := []byte(nonce)

	// 创建加密分组
	block, err := aes.NewCipher(keyByte)
	if err != nil {
		return "", err
	}

	ciphertext := make([]byte, len(plaintext))
	ctr := cipher.NewCTR(block, nonceByte)
	ctr.XORKeyStream(ciphertext, plaintext)
	return hex.EncodeToString(ciphertext), nil
}

// AES-CTR模式解密，解码使用hex模式加密的密文
func CTR_Decrypt(ciphertextStr string, key string, nonce string) (string, error) {
	ciphertext, err := hex.DecodeString(ciphertextStr)
	if err != nil {
		return "", fmt.Errorf("aes-ctr decode error: %s", err.Error())
	}
	keyByte := []byte(key)
	nonceByte := []byte(nonce)
	block, err := aes.NewCipher(keyByte)
	if err != nil {
		return "", fmt.Errorf("new aes cipher error, %s", err.Error())
	}
	ctr := cipher.NewCTR(block, nonceByte)
	plaintext := make([]byte, len(ciphertext))
	ctr.XORKeyStream(plaintext, ciphertext)
	return string(plaintext), nil
}

// RSA加密，使用标准的base64编码输出
func RSA_Encrypt(plaintextStr string, path string) (string, error) {
	// 打开文件
	plaintext := []byte(plaintextStr)
	file, err := os.Open(path)
	if err != nil {
		return "", fmt.Errorf("open public key file failed: %s, %s", path, err.Error())
	}
	defer file.Close()
	// 读取文件的内容
	info, err := file.Stat()
	if err != nil {
		return "", err
	}
	buf := make([]byte, info.Size())
	_, err = file.Read(buf)
	if err != nil {
		return "", err
	}
	// pem解码
	block, _ := pem.Decode(buf)
	// x509解码

	publicKeyInterface, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		return "", fmt.Errorf("parse public key file failed: %s, %s", path, err.Error())
	}
	//类型断言
	publicKey, e := publicKeyInterface.(*rsa.PublicKey)
	if !e {
		return "", fmt.Errorf("check public key type error")
	}
	//对明文进行加密
	cipherText, err := rsa.EncryptPKCS1v15(rand.Reader, publicKey, plaintext)
	if err != nil {
		return "", fmt.Errorf("encode message using public key file failed: %s", err.Error())
	}
	//返回密文
	return base64.StdEncoding.EncodeToString(cipherText), nil
}

// RSA解密，解码使用base64编码得密文
func RSA_Decrypt(cipherTextStr string, path string) (string, error) {
	//打开文件
	ciphertext, err := base64.StdEncoding.DecodeString(cipherTextStr)

	if err != nil {
		return "", fmt.Errorf("decode std base64 message failed: %s", err.Error())
	}
	file, err := os.Open(path)
	if err != nil {
		return "", fmt.Errorf("open private key file failed: %s, %s", path, err.Error())
	}
	defer file.Close()
	//获取文件内容
	info, err := file.Stat()
	if err != nil {
		return "", err
	}
	buf := make([]byte, info.Size())
	_, err = file.Read(buf)
	if err != nil {
		return "", err
	}
	//pem解码
	block, _ := pem.Decode(buf)
	//X509解码
	privateKey, err := x509.ParsePKCS1PrivateKey(block.Bytes)
	if err != nil {

		return "", fmt.Errorf("parse private key file failed: %s, %s", path, err.Error())
	}
	//对密文进行解密
	plainText, err := rsa.DecryptPKCS1v15(rand.Reader, privateKey, ciphertext)
	if err != nil {
		return "", fmt.Errorf("decode message using private key file failed: %s", err.Error())
	}
	//返回明文
	return string(plainText), nil
}
