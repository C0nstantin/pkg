package log

import (
	"github.com/sirupsen/logrus"
	"io"
)

type Logger interface {
	Debugf(format string, args ...interface{})
	Infof(format string, args ...interface{})
	Infoln(args ...interface{})
	Info(args ...interface{})
	Warnf(format string, args ...interface{})
	Errorf(format string, args ...interface{})
	Fatalf(format string, args ...interface{})
	Panicf(format string, args ...interface{})
	Tracef(format string, args ...interface{})
	Printf(format string, args ...interface{})
	AddField(key string, value interface{})
}

type LoggerImpl struct {
	logger *logrus.Logger
	prefix string
}

func (l *LoggerImpl) Debugf(format string, args ...interface{}) {
	l.logger.Debugf(format, args...)
}

func (l *LoggerImpl) Infof(format string, args ...interface{}) {
	l.logger.Infof(format, args...)
}

func (l *LoggerImpl) Infoln(args ...interface{}) {
	l.logger.Infoln(args...)
}

func (l *LoggerImpl) Info(args ...interface{}) {
	l.logger.Info(args...)
}

func (l *LoggerImpl) Warnf(format string, args ...interface{}) {
	l.logger.Warnf(format, args...)
}

func (l *LoggerImpl) Errorf(format string, args ...interface{}) {
	l.logger.Errorf(format, args...)
}

func (l *LoggerImpl) Fatalf(format string, args ...interface{}) {
	l.logger.Fatalf(format, args...)
}

func (l *LoggerImpl) Panicf(format string, args ...interface{}) {
	l.logger.Panicf(format, args...)
}

func (l *LoggerImpl) Tracef(format string, args ...interface{}) {
	l.logger.Tracef(format, args...)
}
func (l *LoggerImpl) Printf(format string, args ...interface{}) {
	l.logger.Printf(format, args...)
}

func (l *LoggerImpl) AddField(key string, value interface{}) {
	l.logger.WithField(key, value)
}

func NewLogger() *LoggerImpl {
	return &LoggerImpl{logger: DefaultLogger()}
}

func NewNopLogger() Logger {
	logger := logrus.New()
	logger.SetOutput(io.Discard)
	return &LoggerImpl{logger: logger}
}
