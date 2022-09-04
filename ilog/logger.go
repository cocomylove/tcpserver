package ilog

import "go.uber.org/zap"

type Logger interface {
	Debug(msg string, f ...zap.Field)
	Info(msg string, f ...zap.Field)
	Warn(msg string, f ...zap.Field)
	Error(msg string, f ...zap.Field)
}
