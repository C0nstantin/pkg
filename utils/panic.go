package utils

import (
	"log"
)

// PanicIfErr panic if err not nil
func PanicIfErr(err error) {
	if err != nil {
		panic(err)
	}
}

type Closer interface {
	Close() error
}
type Notifier interface {
	Notify(err error)
}

// DeferClosePanic close connection if err not nil log error and panic
func DeferClosePanic(c Closer) {
	if err := c.Close(); err != nil {
		log.Panic(err)
	}
}

// DeferCloseLog close connection if err not nil log error
func DeferCloseLog(c Closer) {
	if err := c.Close(); err != nil {
		log.Println(err)
	}
}

// PanicAndNotify panic if err not nil and notify err
func PanicAndNotify(err error, notifier Notifier) {
	if err != nil {
		notifier.Notify(err)
		panic(err)
	}
}
