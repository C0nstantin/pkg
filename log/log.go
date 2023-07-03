package log

import (
	"os"
	"runtime"

	"github.com/sirupsen/logrus"
)

func init() {
	logrus.SetFormatter(&logrus.TextFormatter{
		ForceColors:     true,
		FullTimestamp:   true,
		DisableQuote:    true,
		TimestampFormat: "2006-01-02 15:04:05",
	})
	logrus.SetLevel(logrus.InfoLevel)
	if os.Getenv("DEBUG") != "" {
		logrus.SetLevel(logrus.TraceLevel)
		logrus.Debugln("DEBUG MODE IS ENABLED")
	}
	logrus.SetOutput(os.Stdout)
}

func DefaultLogger() *logrus.Logger {
	logger := logrus.StandardLogger()

	logger.SetLevel(logrus.InfoLevel)
	if os.Getenv("DEBUG") != "" {
		logger.SetLevel(logrus.TraceLevel)
		logger.Debugln("DEBUG MODE IS ENABLED")
	}
	logger.Formatter = &logrus.TextFormatter{
		ForceColors:  true,
		DisableQuote: true,
	}
	logger.Out = os.Stdout
	return logger
}

func Debugf(format string, args ...interface{}) {
	_, file, line, _ := runtime.Caller(1)
	args = append(args, file, line)
	logrus.Debugf(format+"call:%s:%d", args...)
}

func Infof(format string, args ...interface{}) {
	logrus.Infof(format, args...)
}
func Infoln(args ...interface{}) {
	logrus.Infoln(args...)
}
func Info(args ...interface{}) {
	logrus.Info(args...)
}

func Warnf(format string, args ...interface{}) {
	logrus.Warnf(format, args...)
}

func Errorf(format string, args ...interface{}) {
	logrus.Errorf(format, args...)
}

func Fatalf(format string, args ...interface{}) {
	logrus.Fatalf(format, args...)
}
func Panicf(format string, args ...interface{}) {
	logrus.Panicf(format, args...)
}

func Printf(format string, args ...interface{}) {
	logrus.Printf(format, args...)
}

func Println(args ...interface{}) {
	logrus.Println(args...)
}

func Tracef(format string, args ...interface{}) {
	logrus.Tracef(format, args...)
}

func Panic(args ...interface{}) {
	logrus.Panic(args...)
}
