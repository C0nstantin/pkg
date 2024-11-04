package log

import (
	"github.com/sirupsen/logrus"
	"os"
	"runtime"
	"strings"
)

func init() {
	logrus.SetFormatter(&logrus.TextFormatter{
		ForceColors:     true,
		FullTimestamp:   true,
		DisableQuote:    true,
		TimestampFormat: "2006-01-02 15:04:05",
	})
	logrus.SetLevel(logrus.InfoLevel)
	if os.Getenv("DEBUG") != "" || strings.ToUpper(os.Getenv("LOG_LEVEL")) == "DEBUG" {
		logrus.SetLevel(logrus.DebugLevel)
		logrus.Debugln("DEBUG MODE IS ENABLED")
	}
	if strings.ToUpper(os.Getenv("LOG_LEVEL")) == "TRACE" {
		logrus.SetLevel(logrus.TraceLevel)
		logrus.Debugln("TRACE MODE IS ENABLED")

	}

	logrus.SetOutput(os.Stdout)
}

func DefaultLogger() *logrus.Logger {
	logger := logrus.StandardLogger()

	logger.SetLevel(logrus.InfoLevel)
	if os.Getenv("DEBUG") != "" || strings.ToUpper(os.Getenv("LOG_LEVEL")) == "DEBUG" {
		logger.SetLevel(logrus.DebugLevel)
	}
	if strings.ToUpper(os.Getenv("LOG_LEVEL")) == "TRACE" {
		logger.SetLevel(logrus.TraceLevel)

	}
	logger.Formatter = &logrus.TextFormatter{
		ForceColors:     true,
		FullTimestamp:   true,
		DisableQuote:    true,
		TimestampFormat: "2006-01-02 15:04:05",
	}
	logger.Out = os.Stdout
	return logger
}

func Debugf(format string, args ...interface{}) {
	_, file, line, _ := runtime.Caller(1)
	logrus.Debugf(format+" call:%s:%d", append(args, file, line)...)
	//logrus.Debugf(format, args...)
}

func Infof(format string, args ...interface{}) {
	if logrus.GetLevel() >= logrus.DebugLevel {
		_, file, line, _ := runtime.Caller(1)
		logrus.Infof(format+" call:%s:%d", append(args, file, line)...)
	} else {
		logrus.Infof(format, args...)
	}
}
func Infoln(args ...interface{}) {
	logrus.Infoln(args...)
}
func Info(args ...interface{}) {
	logrus.Info(args...)
}

func Warnf(format string, args ...interface{}) {
	if logrus.GetLevel() >= logrus.DebugLevel {
		_, file, line, _ := runtime.Caller(1)
		logrus.Warnf(format+" call:%s:%d", append(args, file, line)...)
	} else {
		logrus.Warnf(format, args...)
	}
}

func Errorf(format string, args ...interface{}) {
	if logrus.GetLevel() >= logrus.DebugLevel {
		_, file, line, _ := runtime.Caller(1)
		logrus.Errorf(format+" call:%s:%d", append(args, file, line)...)
	} else {
		logrus.Errorf(format, args...)
	}
}

func Fatalf(format string, args ...interface{}) {
	if logrus.GetLevel() >= logrus.DebugLevel {
		_, file, line, _ := runtime.Caller(1)
		logrus.Fatalf(format+" call:%s:%d", append(args, file, line)...)
	} else {
		logrus.Fatalf(format, args...)
	}
}
func Panicf(format string, args ...interface{}) {
	if logrus.GetLevel() >= logrus.DebugLevel {
		_, file, line, _ := runtime.Caller(1)
		logrus.Panicf(format+" call:%s:%d", append(args, file, line)...)
	} else {
		logrus.Panicf(format, args...)
	}
}

func Printf(format string, args ...interface{}) {
	if logrus.GetLevel() >= logrus.DebugLevel {
		_, file, line, _ := runtime.Caller(1)
		logrus.Printf(format+" call:%s:%d", append(args, file, line)...)
	} else {
		logrus.Printf(format, args...)
	}
}

func Println(args ...interface{}) {
	logrus.Println(args...)
}

func Tracef(format string, args ...interface{}) {
	_, file, line, _ := runtime.Caller(1)
	logrus.Tracef(format+" call:%s:%d", append(args, file, line)...)

}

func Panic(args ...interface{}) {
	logrus.Panic(args...)

}
