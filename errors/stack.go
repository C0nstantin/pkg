package errors

import (
	"fmt"

	"github.com/pkg/errors"
)

type StackTracer interface {
	StackTrace() errors.StackTrace
}

type errWithStack struct {
	message string
	err     error
}

func (e *errWithStack) Unwrap() error {
	return e.err
}
func (e *errWithStack) Cause() error { return e.err }
func (e *errWithStack) Format(s fmt.State, verb rune) {
	switch verb {
	case 'v':
		if s.Flag('+') {
			fmt.Fprintf(s, "%s\n%+v", e.message, e.err)
			return
		}
		fallthrough
	case 's':
		fmt.Fprint(s, e.message)
	case 'q':
		fmt.Fprintf(s, "%q", e.message)
	}
}
func (e *errWithStack) Error() string {
	return e.message
}

func (e *errWithStack) StackTrace() errors.StackTrace {
	if i, ok := e.err.(StackTracer); ok {
		return i.StackTrace()
	}
	errorWithStack := errors.WithStack(e.err)
	return errorWithStack.(StackTracer).StackTrace()
}
