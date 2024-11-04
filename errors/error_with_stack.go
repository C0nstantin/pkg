package errors

import (
	standardError "errors"
	"fmt"
)

type StackTracer interface {
	StackTrace() StackTrace
}

type errWithStack struct {
	err   error
	stack StackTrace
}

func (e *errWithStack) Is(err error) bool {
	return standardError.Is(e.err, err)
}

/*func (e *errWithStack) As(target interface{}) bool {
	return standardError.As(e, target)
}*/

func (e *errWithStack) message() string {
	return e.err.Error()
}

func (e *errWithStack) Error() string {
	return e.message()
}

func (e *errWithStack) Format(s fmt.State, verb rune) {
	switch verb {
	case 'v':
		if s.Flag('+') {
			fmt.Fprintf(s, "%s\n%+v", e.message(), e.stack)
			return
		}
		fallthrough
	case 's':
		fmt.Fprint(s, e.message())
	case 'q':
		fmt.Fprintf(s, "%q", e.message())
	}
}

func (e *errWithStack) StackTrace() StackTrace {
	return e.stack
}

func (e *errWithStack) Unwrap() error {
	return e.err // e.err
}
