package hmac

import (
	"errors"

	"github.com/gin-gonic/gin"
)

var ErrAuthHeaderNotFound = errors.New("authorization header not found")
var ErrRequestUnAuthorized = errors.New("request not authorize")
var ErrAuthHeaderInvalid = errors.New("invalid format auth header")
var ErrMethodNotSupported = errors.New("hash method not supported")
var ErrAccessIDNotSet = errors.New("access id not set")

type AuthMiddleware struct {
	HmacAuth APIAuth
}

func (h AuthMiddleware) AuthMiddleware() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		err := func(ctx *gin.Context) error {
			return h.HmacAuth.CheckRequest(ctx.Request)
		}(ctx)
		if err != nil {
			_ = ctx.Error(err)
			ctx.Abort()
			return
		}
		ctx.Next()
	}
}
