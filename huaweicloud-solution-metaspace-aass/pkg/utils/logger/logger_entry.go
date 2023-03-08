// Copyright (c) Huawei Technologies Co., Ltd. 2022-2022. All rights reserved.

// 日志入口
package logger

import (
	"context"
	"os"
	"fmt"
	"strconv"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
)

const (
	// 避免循环依赖，这里env常量未归入setting包
	envLogRotateSize  = "LOG_ROTATE_SIZE"
	envLogBackupCount = "LOG_BACKUP_COUNT"
	envLogMaxAge	  = "LOG_MAX_AGE"
 
	defaultLogRotateSize  = 100
	defaultLogBackupCount = 100
	defaultLogMaxAge	  = 7
)

var encoderConfig = zapcore.EncoderConfig{
	// Keys can be anything except the empty string.
	LineEnding:     zapcore.DefaultLineEnding,
	EncodeLevel:    zapcore.LowercaseLevelEncoder,
	EncodeTime:     zapcore.ISO8601TimeEncoder,
	EncodeDuration: zapcore.SecondsDurationEncoder,
	EncodeCaller:   zapcore.ShortCallerEncoder,
	TimeKey:        Timestamp,
	LevelKey:       Level,
	NameKey:        "logger",
	CallerKey:      Caller,
	MessageKey:     Msg,
	StacktraceKey:  Stacktrace,
}

type FMLogger struct {
	logger *zap.SugaredLogger
	ctx    context.Context
}

// Debug print debug message
func (l *FMLogger) Debug(format string, a ...interface{}) {
	l.logger.Debugf(format, a...)
}

// Info print info message
func (l *FMLogger) Info(format string, a ...interface{}) {
	l.logger.Infof(format, a...)
}

// Warn print warning message
func (l *FMLogger) Warn(format string, a ...interface{}) {
	l.logger.Warnf(format, a...)
}

// Error print error message
func (l *FMLogger) Error(format string, a ...interface{}) {
	l.logger.Errorf(format, a...)
}

// WithField tag logger with key and value
func (l *FMLogger) WithField(key string, value interface{}) *FMLogger {
	logger := &FMLogger{
		ctx:    context.WithValue(l.ctx, "key", "value"),
		logger: l.logger.With(key, value),
	}

	return logger
}

// WithFields tag logger with multiple key and values
func (l *FMLogger) WithFields(fields map[string]interface{}) *FMLogger {
	var args []interface{}
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

	maxSize, err := strconv.Atoi(os.Getenv(envLogRotateSize))
	if err != nil {
		err = nil
		maxSize = defaultLogRotateSize
	}
	backupCount, err := strconv.Atoi(os.Getenv(envLogBackupCount))
	if err != nil {
		backupCount = defaultLogBackupCount
	}
	maxAge, err := strconv.Atoi(os.Getenv(envLogMaxAge))
	if err != nil {
		maxAge = defaultLogMaxAge
	}
	fmt.Printf("Init logger setting: MaxSize: %d, MaxAge: %d, MaxBackups: %d, logPath: %s\n", 
		maxSize, maxAge, backupCount, filePath)
	fileWriter := zapcore.AddSync(&lumberjack.Logger{
		Filename:   filePath,
		MaxSize:    maxSize,     // megabytes
		MaxBackups: backupCount, // backup log files
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
