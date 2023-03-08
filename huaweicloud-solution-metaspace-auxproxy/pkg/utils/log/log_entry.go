// Copyright (c) Huawei Technologies Co., Ltd. 2012-2018. All rights reserved.

package log

import (
	"context"
	"os"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
)

type FMLogger struct {
	logger *zap.SugaredLogger
	ctx    context.Context
}

// Debugf print debug message
func (l *FMLogger) Debugf(format string, a ...interface{}) {
	l.logger.Debugf(format, a...)
}

// Infof print info message
func (l *FMLogger) Infof(format string, a ...interface{}) {
	l.logger.Infof(format, a...)
}

// Warnf print warning message
func (l *FMLogger) Warnf(format string, a ...interface{}) {
	l.logger.Warnf(format, a...)
}

// Errorf print error message
func (l *FMLogger) Errorf(format string, a ...interface{}) {
	l.logger.Errorf(format, a...)
}

// WithField tag logger with key and value
func (l *FMLogger) WithField(k string, v interface{}) *FMLogger {
	logger := &FMLogger{
		ctx:    context.WithValue(l.ctx, "key", "value"),
		logger: l.logger.With(k, v),
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
	// 创建日志配置
	var encoderConfig = zap.NewProductionEncoderConfig()
	encoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder

	// 创建Console输出Core
	consoleEncoder := zapcore.NewJSONEncoder(encoderConfig)
	consoleCore := zapcore.NewCore(consoleEncoder, zapcore.Lock(os.Stdout), atomicLevel)

	// 创建日志输出Core
	fileWriter := zapcore.AddSync(&lumberjack.Logger{
		Filename: filePath,
		MaxSize:  50, // megabytes
		MaxAge:   7,
		Compress: true,
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
