package errors

import (
	"fmt"
	"strings"
	"testing"
)

type TestError struct {
}

func (e *TestError) Error() string {
	return "test error"
}

type FatalError struct {
}

func (e *FatalError) Error() string {
	return "fatal error"
}
func TestErrorfAs(t *testing.T) {
	e := Errorf("%s bla bla bla %v fff", "hello", &TestError{})
	e2 := Errorf("%s bla bla bla %v fff", "hello", e)
	var f *FatalError
	if As(e, &f) {
		t.Error("should not be fatal error")
	}
	if As(e2, &f) {
		t.Error("should  not be fatal error")
	}
}
func TestMultiWrap(t *testing.T) {
	e := Errorf("%s bla bla bla %v fff", "hello", &TestError{})
	e2 := Wrap(&FatalError{}, e)
	e3 := Wrap(&FatalError{}, e2)
	var f *FatalError
	if !As(e2, &f) {
		t.Error("should be fatal error")
	}
	if !As(e3, &f) {
		t.Error("should be fatal error")
	}
	ee, ok := e3.(*errWithStack)
	if !ok {
		t.Error("should be error")
	}
	if !strings.Contains(fmt.Sprintf("%+v", ee), "TestMultiWrap") {
		t.Error("should contain TestMultiWrap")
	}
	if !strings.Contains(fmt.Sprintf("%+v", ee), "fatal error") {
		t.Error("should contain FatalError")
	}
	if ee.Error() != "fatal error" {
		t.Error("should be fatal error")
	}
}
