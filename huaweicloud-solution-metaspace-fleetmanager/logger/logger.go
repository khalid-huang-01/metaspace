// Copyright (c) Huawei Technologies Co., Ltd. 2022-2022. All rights reserved.

// 日志服务
package logger

import (
	"context"
	"fleetmanager/env"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
	"os"
	"fmt"
	"strconv"
)

const (
	DefaultMaxBackupNum = 100	// 日志文件的最大数量
	DefaultRotateSize   = 100	// 每一个日志文件的最大尺寸
	DefaultMaxAge		= 7		// 日志文件的最大保存时间
)

var encoderConfig = zapcore.EncoderConfig{
	// Keys can be anything except the empty string.
	TimeKey:        Timestamp,
	LevelKey:       Level,
	NameKey:        "logger",
	CallerKey:      Caller,
	MessageKey:     Msg,
	StacktraceKey:  Stacktrace,
	LineEnding:     zapcore.DefaultLineEnding,
	EncodeLevel:    zapcore.LowercaseLevelEncoder,
	EncodeTime:     zapcore.ISO8601TimeEncoder,
	EncodeDuration: zapcore.SecondsDurationEncoder,
	EncodeCaller:   zapcore.ShortCallerEncoder,
}

type FMLogger struct {
	logger *zap.SugaredLogger
	ctx    context.Context
}

// Debug Debug级日志
func (l *FMLogger) Debug(format string, a ...interface{}) {
	l.logger.Debugf(format, a...)
}

// Info Info级日志
func (l *FMLogger) Info(format string, a ...interface{}) {
	l.logger.Infof(format, a...)
}

// Warn Warn级日志
func (l *FMLogger) Warn(format string, a ...interface{}) {
	l.logger.Warnf(format, a...)
}

// Error Error级日志
func (l *FMLogger) Error(format string, a ...interface{}) {
	l.logger.Errorf(format, a...)
}

// WithField 构造logger
func (l *FMLogger) WithField(path string, value interface{}) *FMLogger {
	logger := &FMLogger{
		ctx:    context.WithValue(l.ctx, "key", "value"),
		logger: l.logger.With(path, value),
	}

	return logger
}

// WithFields 构造logger
func (l *FMLogger) WithFields(fields map[string]interface{}) *FMLogger {
	args := make([]interface{}, 0)
	for k, v := range fields {
		args = append(args, k)
		args = append(args, v)
	}
	logger := &FMLogger{
		ctx:    context.WithValue(l.ctx, "key", "value"),
		logger: l.logger.With(args...),
	}

	return logger
}

func initLogger(filePath string, atomicLevel *zap.AtomicLevel, fields ...zap.Field) (*FMLogger, error) {
	consoleEncoder := zapcore.NewJSONEncoder(encoderConfig)
	consoleCore := zapcore.NewCore(consoleEncoder, zapcore.Lock(os.Stdout), atomicLevel)

	maxSize, err := strconv.Atoi(os.Getenv(env.LogRotateSize))
	if err != nil {
		err = nil
		maxSize = DefaultRotateSize
	}

	backupCount, err := strconv.Atoi(os.Getenv(env.LogBackupCount))
	if err != nil {
		backupCount = DefaultMaxBackupNum
	}

	maxAge, err := strconv.Atoi(os.Getenv(env.LogMaxAge))
	if err != nil {
		maxAge = DefaultMaxAge
	}
	fmt.Printf("Init logger setting: MaxSize: %d, MaxAge: %d, MaxBackups: %d, logPath: %s\n", 
		maxSize, maxAge, backupCount, filePath)
	fileWriter := zapcore.AddSync(&lumberjack.Logger{
		Filename:   filePath,
		MaxSize:    maxSize, // megabytes
		MaxBackups: backupCount,
		MaxAge: 	maxAge,
		Compress:   true,
	})
	fileCore := zapcore.NewCore(
		zapcore.NewJSONEncoder(encoderConfig),
		fileWriter,
		atomicLevel,
	)

	cores := zapcore.NewTee(fileCore, consoleCore)
	log := zap.New(cores)
	log = log.WithOptions(
		zap.ErrorOutput(os.Stdout), // error message output to stdout
		zap.AddCaller(),            // add function caller info to log
		zap.AddCallerSkip(1),       // make stack having right depth to get function call
		zap.Fields(fields...),      // add common log info, like local_ip
	)

	logger := FMLogger{
		logger: log.Sugar(),
		ctx:    context.Background(),
	}

	return &logger, nil
}
