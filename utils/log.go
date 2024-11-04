package utils

import (
	"log"
	"os"
	"runtime"
	"strings"
)

func Debugf(format string, args ...interface{}) {
	if strings.ToLower(os.Getenv("LOG_LEVEL")) != "debug" ||
		strings.ToLower(os.Getenv("LOG_LEVEL")) != "trace" {
		return
	}
	_, file, line, _ := runtime.Caller(1)
	log.Printf("[DEBUG] "+format+" call:%s:%d", append(args, file, line)...)
}

func Tracef(format string, args ...interface{}) {
	if strings.ToLower(os.Getenv("LOG_LEVEL")) != "trace" {
		return
	}
	_, file, line, _ := runtime.Caller(1)
	log.Printf("[TRACE] "+format+" call:%s:%d", append(args, file, line)...)
}
