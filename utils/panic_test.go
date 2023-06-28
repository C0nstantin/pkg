package utils

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

// test for PanicIfError
func TestPanicIfError(t *testing.T) {
	t.Run("check nil error ", func(t *testing.T) {
		assert.NotPanics(t, func() {
			PanicIfErr(nil)
		})

	})
	t.Run("check non-nil error ", func(t *testing.T) {
		err := errors.New("test")
		assert.Panics(t, func() {
			PanicIfErr(err)
		})
	})
}
