//go:build !windows
// +build !windows

// Copyright (c) Huawei Technologies Co., Ltd. 2022-2022. All rights reserved.

// signal_posix
package utils

import (
	"os"
	"syscall"
)

var shutdownSignals = []os.Signal{os.Interrupt, syscall.SIGTERM}
