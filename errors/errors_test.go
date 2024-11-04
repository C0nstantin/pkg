package errors

import (
	e2 "errors"
	"fmt"
)
import "testing"

// test Errorf and Is
func TestErrorfAndIs(t *testing.T) {
	// std lib
	t.Run("std lib", func(t *testing.T) {
		FirstErr := e2.New("first error")
		SecondErr := fmt.Errorf("second error %w", FirstErr)
		if !e2.Is(SecondErr, FirstErr) {
			t.Error("SecondErr should be equal to FirstErr")
		}
		if !Is(SecondErr, FirstErr) {
			t.Error("SecondErr should be equal to FirstErr")
		}
	})
	t.Run("simple custom", func(t *testing.T) {
		FirstCustomErr := New("first error")
		SecondCustomErr := fmt.Errorf("second error %w", FirstCustomErr)
		if !e2.Is(SecondCustomErr, FirstCustomErr) {
			t.Error("SecondErr should be equal to FirstErr")
		}
		if !Is(SecondCustomErr, FirstCustomErr) {
			t.Error("SecondErr should be equal to FirstErr")
		}

		if _, ok := FirstCustomErr.(StackTracer); !ok {
			t.Error("FirstErr should have stack trace")
		}
		//fmt.Printf("%+v\n", FirstCustomErr)  //with stack trace
		//fmt.Printf("%+v\n", SecondCustomErr) // without  stack trace
	})
	t.Run("std errors", func(t *testing.T) {
		FirstError := e2.New("first error")
		SecondError := Errorf("second error %v", FirstError)
		if !e2.Is(SecondError.(*errWithStack).err, FirstError) {
			t.Error("SecondErr should be equal to FirstErr")
		}

		if _, ok := SecondError.(StackTracer); !ok {
			t.Error("FirstErr should have stack trace")
		}

	})
	t.Run("custom", func(t *testing.T) {
		FirstError := New("first error")
		SecondError := Errorf("second error %v", FirstError)
		if !e2.Is(SecondError, FirstError) {
			t.Error("SecondErr should be equal to FirstErr")
		}

		if _, ok := FirstError.(StackTracer); !ok {
			t.Error("FirstErr should have stack trace")
		}
		if _, ok := SecondError.(StackTracer); !ok {
			t.Error("SecondError should have stack trace")
		}
		if !Is(SecondError, FirstError) {
			t.Error("SecondError should be equal to FirstErr")
		}
	})

}

func TestE(t *testing.T) {

	t.Run("NilError", func(t *testing.T) {
		err := E(nil)
		if err != nil {
			t.Errorf("Expected nil error, got %v", err)
		}
	})

	t.Run("CustomeError", func(t *testing.T) {
		customErr := New("custom error")
		err := E(customErr)
		if err == nil {
			t.Error("Expected a non-nil error, got nil")
		}
		if !e2.Is(err, customErr) {
			t.Errorf("Expected %v, got %v", customErr, err)
		}
		_, ok := err.(StackTracer)
		if !ok {
			t.Error("Expected error to implement StackTracer interface")
		}
	})

	t.Run("ErrorWithStack", func(t *testing.T) {
		errWithStack := &errWithStack{err: e2.New("error with stack"), stack: NewStackTracer()}
		err := E(errWithStack)
		if err == nil {
			t.Error("Expected a non-nil error, got nil")
		}
		if !Is(err, errWithStack.err) {
			t.Errorf("Expected %v, got %v", errWithStack.err, err)
		}
		_, ok := err.(StackTracer)
		if !ok {
			t.Error("Expected error to implement StackTracer interface")
		}
	})

	t.Run("ErrorWithoutStack", func(t *testing.T) {
		errWithoutStack := e2.New("error without stack")
		err := E(errWithoutStack)
		if err == nil {
			t.Error("Expected a non-nil error, got nil")
		}
		if !Is(err, errWithoutStack) {
			t.Errorf("Expected %v, got %v", errWithoutStack, err)
		}
		_, ok := err.(StackTracer)
		if !ok {
			t.Error("Expected error to implement StackTracer interface")
		}
	})

}

func TestEr(t *testing.T) {
	t.Run("NilError", func(t *testing.T) {
		err := Er(nil, "error")
		if err != nil {
			t.Errorf("Expected nil error, got %v", err)
		}
	})
	t.Run("CustomError", func(t *testing.T) {
		customErr := New("custom error")
		err := Er(customErr, "error")
		if err == nil {
			t.Error("Expected a non-nil error, got nil")
		}
		if !e2.Is(err, customErr) {
			t.Errorf("Expected %v, got %v", customErr, err)
		}
		_, ok := err.(StackTracer)
		if !ok {
			t.Error("Expected error to implement StackTracer interface")
		}
		errr := Er(err, "ffff")
		var r *errWithStack
		if e2.As(errr, &r) {
			t.Log("error has stack trace")
		} else {
			t.Error("error does not have stack trace")
		}
	})

	t.Run("noErrorWithStack", func(t *testing.T) {
		errnWithStack := e2.New("error without stack")
		err := Er(errnWithStack, "error")
		if err == nil {
			t.Error("Expected a non-nil error, got nil")
		}
		if !Is(err, errnWithStack) {
			t.Errorf("Expected %v, got %v", errnWithStack, err)
		}
		_, ok := err.(StackTracer)
		if !ok {
			t.Error("Expected error to implement StackTracer interface")
		}
		var errWithStack *errWithStack
		if !e2.As(err, &errWithStack) {
			t.Error("error should not have stack trace")
		}
	})

}

func TestWrap(t *testing.T) {
	t.Run("WrapsErrorWithStackTrace", func(t *testing.T) {
		err1 := e2.New("original error")
		err2 := e2.New("cause of error")
		wrappedErr := Wrap(err1, err2)

		if !e2.Is(wrappedErr, err1) {
			t.Errorf("Expected wrapped error to be equal to err1, got %v", wrappedErr)
		}

		var e *errWithStack
		if !As(wrappedErr, &e) {
			t.Error("Expected wrapped error to have a stack trace")
		}
	})

	t.Run("WrapsErrorWithoutStackTrace", func(t *testing.T) {
		err1 := e2.New("original error")
		err2 := e2.New("cause of error")
		wrappedErr := Wrap(err1, err2)

		if !e2.Is(wrappedErr, err1) {
			t.Errorf("Expected wrapped error to be equal to err1, got %v", wrappedErr)
		}

		var e *errWithStack
		if !e2.As(wrappedErr, &e) {
			t.Error("Expected wrapped error to not have a stack trace")
		}
	})

	t.Run("DoesNotWrapNilError", func(t *testing.T) {
		var err1 error
		err2 := New("cause of error")
		wrappedErr := Wrap(err1, err2)

		if wrappedErr != err2 {
			t.Errorf("Expected wrapped error to be nil, got %v", wrappedErr)
		}
	})

	t.Run("DoesNotWrapNilCause", func(t *testing.T) {
		err1 := New("original error")
		var err2 error
		wrappedErr := Wrap(err1, err2)
		if wrappedErr != nil {
			t.Errorf("Expected wrapped error to be nil, got %v", wrappedErr)
		}

	})
}
