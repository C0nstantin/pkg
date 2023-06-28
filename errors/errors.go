package errors

import (
	"fmt"

	"github.com/pkg/errors"
	logs "github.com/sirupsen/logrus"
)

// Er  wraps an error with a stack trace and adds a message to the error.
//
// params
// - e error everything implementing the error interface
// - str formatted message
// - options array of options
//
// return error with right stack trace , if e has a stack trace nothing will be added
func Er(e error, str string, options ...interface{}) error {
	return &errWithStack{
		message: fmt.Sprintf(str, options...) + " : " + e.Error(),
		err:     e,
	}
}

// E wrap  error with right stack trace
//
// Use this function to wrap error with stack trace anywhere.
// Params:
// - err error any error ( implements Error interface )
//
//	It return
//	error with stack trace
//
// if err has stack trace, return it
func E(err error) error {
	if err == nil {
		return nil
	}
	if _, ok := err.(StackTracer); ok {
		return err
	}
	return errors.WithStack(err)
}

func ELog(returnedError, loggedError error) error {
	if loggedError == nil {
		return nil
	}
	logs.Error("returnedError %+v\nLoggedError %+v\n", returnedError, loggedError)
	return E(returnedError)
}

// Is wrap standard func Is from package errors
func Is(err error, target error) bool {
	return errors.Is(err, target)
}

// As wrap standard func  As from package errors
func As(err error, target interface{}) bool {
	return errors.As(err, target)
}

// Errorf returns an error with formatted message and stack trace
func Errorf(format string, args ...interface{}) error {
	return errors.Errorf(format, args...)
}

// New returns a new error with the given message and stack trace.
func New(s string) error {
	return errors.New(s)
}

func WithStack(err error) error {
	return E(err)
}

func Wrap(err error, s string, options ...interface{}) error {
	return Er(err, s, options...)
}

func WithMessage(err error, s string) error {
	return Er(err, s)
}
