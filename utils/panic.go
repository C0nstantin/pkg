package utils

import (
	`log`
)

// func panic if err input error  panic if err not nil
func PanicIfErr(err error) {
	if err != nil {
		panic(err)
	}
}

type Closer interface {
	Close() error
}

func DeferClosePanic(c Closer) {
	if err := c.Close(); err != nil {
		log.Panic(err)
	}
}

func DeferCloseLog(c Closer) {
	if err := c.Close(); err != nil {
		log.Println(err)
	}
}
