// Copyright (c) Huawei Technologies Co., Ltd. 2022-2022. All rights reserved.

// aksk加密
package hhmac

import (
	"fmt"
	"sync"
)

// Sker Get sk through ak.
type Sker interface {
	GetSk(ak string) ([]byte, error)
}

type StoreSker interface {
	Sker
	Add(ak string, ent KeyEntry)
}

type KeyEntry struct {
	SK []byte
}

// 远端 访问 本地 时的认证秘钥存储器
type keyStore struct {
	table map[string]KeyEntry
	lock  sync.RWMutex

	getter func() ([]byte, error)
}

// StoreKeyEntry used for remote to local
type StoreKeyEntry struct {
	AK       string `json:"ak"`
	SKCypher []byte `json:"sk"`
}

func NewStore() StoreSker {
	return &keyStore{
		table: make(map[string]KeyEntry),
	}
}

func (std *keyStore) Add(ak string, ent KeyEntry) {
	std.lock.Lock()
	std.table[ak] = ent
	std.lock.Unlock()
}

// GetSk gets key from setting
func (std *keyStore) GetSk(ak string) ([]byte, error) {
	std.lock.RLock()
	ent, ok := std.table[ak]
	std.lock.RUnlock()

	if !ok {
		return nil, fmt.Errorf("no key for %q", ak)
	}

	return ent.SK, nil
}

// ----------------------------------------------------

type LocalSker interface {
	Sker
	Add(receiver string, ent LocalKeyEntry)
	GetAk(receiver string) (string, error)
	EnableAuth(receiver string) bool
}

type LocalKeyEntry struct {
	Enable bool   `json:"enable"`
	AK     string `json:"ak"`
	SK     []byte `json:"sk"`
}

// 本地 访问 远端 时的认证秘钥存储器
type localKey struct {
	table map[string]LocalKeyEntry
	lock  sync.RWMutex

	getter func() ([]byte, error)
}

func NewlocalKey() *localKey {
	return &localKey{
		table: make(map[string]LocalKeyEntry),
	}
}

func (std *localKey) Add(receiver string, ent LocalKeyEntry) {
	std.lock.Lock()
	std.table[receiver] = ent
	std.lock.Unlock()
}

// GetSk gets key from setting
func (std *localKey) GetSk(receiver string) ([]byte, error) {
	std.lock.RLock()
	ent, ok := std.table[receiver]
	std.lock.RUnlock()

	if !ok {
		return nil, fmt.Errorf("local key no key for %q", receiver)
	}

	return ent.SK, nil
}

func (std *localKey) GetAk(receiver string) (string, error) {
	std.lock.RLock()
	ent, ok := std.table[receiver]
	std.lock.RUnlock()

	if !ok || ent.AK == "" {
		return "", fmt.Errorf("local key no key for %q", receiver)
	}

	return ent.AK, nil
}

func (std *localKey) EnableAuth(receiver string) bool {
	std.lock.RLock()
	ent, ok := std.table[receiver]
	std.lock.RUnlock()

	if !ok {
		return false
	}

	return ent.Enable
}
