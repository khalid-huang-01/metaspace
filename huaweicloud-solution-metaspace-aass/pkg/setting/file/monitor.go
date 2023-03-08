// Copyright (c) Huawei Technologies Co., Ltd. 2022-2022. All rights reserved.

// 监控服务
package file

import (
	"fmt"
	"os"
	"time"
)

type monitor struct {
	name    string
	modTime time.Time
	load    func(string) error
}

var monitors map[string]monitor

func init() {
	monitors = make(map[string]monitor)
}

// NewMonitor get new monitor for config
func NewMonitor(file string, l func(string) error) error {
	w := monitor{
		name: file,
		load: l,
	}

	f, err := os.Stat(file)
	if err != nil {
		return err
	}

	w.modTime = f.ModTime()
	monitors[w.name] = w
	return nil
}

// MonitorStart start a monitor for config info
func MonitorStart() {
	go func() {
		ticker := time.NewTicker(2 * time.Second)
		for range ticker.C {
			load()
		}
	}()
}

func load() {
	for _, w := range monitors {
		f, err := os.Stat(w.name)
		if err != nil {
			fmt.Printf("stat file %s failed %v\n", w.name, err)
			continue
		}

		modTime := f.ModTime()
		if w.modTime.Equal(modTime) {
			continue
		}

		if err := w.load(w.name); err != nil {
			fmt.Printf("load %s error %s\n", w.name, err)
		}
		w.modTime = modTime
		monitors[w.name] = w
	}
}
