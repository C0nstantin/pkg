package errors

import (
	standardError "errors"
	"fmt"
	"strings"
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
	if e == nil {
		return nil
	}
	var s *errWithStack
	if standardError.As(e, &s) {
		return &errWithStack{err: fmt.Errorf(fmt.Sprintf(str, options...)+" : %w", e), stack: s.stack}
	}
	return &errWithStack{err: fmt.Errorf(fmt.Sprintf(str, options...)+" : %w", e), stack: NewStackTracer()}
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
	return &errWithStack{err: err, stack: NewStackTracer()}
}

// Is wrap standard func Is from package errors
func Is(err error, target error) bool {
	return standardError.Is(err, target)
}

// As wrap standard func  As from package errors
func As(err error, target interface{}) bool {
	return standardError.As(err, target)
}

func Errorf(format string, args ...interface{}) error {
	for i := 0; i < len(args); i++ {
		if customError, ok := args[i].(error); ok {
			var e *errWithStack

			if standardError.As(customError, &e) {
				return &errWithStack{err: fmt.Errorf(strings.Replace(format, "%v", "%w", -1), args...), stack: e.stack}
			}
		}
	}
	return &errWithStack{err: fmt.Errorf(strings.Replace(format, "%v", "%w", -1), args...), stack: NewStackTracer()}
}

// New returns a new error with the given message and stack trace.
func New(s string) error {
	return &errWithStack{
		err:   fmt.Errorf(s),
		stack: NewStackTracer(),
	}
}

func Unwrap(err error) error {
	return standardError.Unwrap(err)
}

func Join(errs ...error) error {
	return standardError.Join(errs...)
}

func Wrap(err, err2 error) error {
	if err2 == nil {
		return nil
	}
	if err == nil {
		return err2
	}
	var e *errWithStack
	if As(err2, &e) {
		return &errWithStack{err: err, stack: e.stack} // err2
	}
	fmt.Printf("Warning: change error wtithout stack trace: %v to %v \n", err2, err)
	return &errWithStack{err: err, stack: NewStackTracer()}
}
