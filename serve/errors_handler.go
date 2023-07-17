package serve

import (
	"net/http"
	
	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
	
	"github.com/pkg/errors"
)

type Notificator interface {
	Notify(err interface{}, r *http.Request)
}
type DefaultErrorHandler struct {
	Logger                  *log.Logger
	BadRequestErrors        []error
	NotFoundErrorErrors     []error
	ForbiddenErrors         []error
	UnauthorizedErrors      []error
	IgnoreErrors            []error
	UnauthorizedTypeErrors  []interface{}
	BadRequestTypeErrors    []interface{}
	NotFoundErrorTypeErrors []interface{}
	ForbiddenTypeErrors     []interface{}
	
	Notifier Notificator
}

func (e *DefaultErrorHandler) Handler() []gin.HandlerFunc {
	return []gin.HandlerFunc{func(ctx *gin.Context) {
		ctx.Next()
		err := ctx.Errors.Last()
		if err == nil {
			return
		}
		if e.Logger == nil {
			e.Logger = log.StandardLogger()
		}
		e.Logger.Infof("Error in handler: %s", err.Error())
		e.Logger.Infof("Error unwrapped: %+v", err.Unwrap())
		
		for _, er := range e.IgnoreErrors {
			if errors.Is(err, er) {
				e.Logger.Debugf("Ignore error: %s", err.Error())
			}
			return
		}
		
		if err.Error() == "EOF" {
			ctx.JSON(http.StatusBadRequest,
				gin.H{"error": "empty_request_params"},
			)
			return
		}
		
		if err.Type == gin.ErrorTypeBind {
			ctx.JSON(http.StatusBadRequest,
				gin.H{"error": "invalid_request_params"},
			)
			return
		}
		for _, err2 := range e.BadRequestErrors {
			if errors.Is(err, err2) {
				ctx.JSON(http.StatusBadRequest,
					gin.H{"error": err.Error()},
				)
				return
			}
		}
		for _, typeError := range e.BadRequestTypeErrors {
			if errors.As(err, typeError) {
				ctx.JSON(http.StatusBadRequest,
					gin.H{"error": err.Error()},
				)
				return
			}
		}
		for _, err2 := range e.NotFoundErrorErrors {
			if errors.Is(err, err2) {
				ctx.JSON(http.StatusNotFound,
					gin.H{"error": err.Error()},
				)
				return
			}
		}
		for _, typeError := range e.NotFoundErrorTypeErrors {
			if errors.As(err, typeError) {
				ctx.JSON(http.StatusNotFound,
					gin.H{"error": err.Error()},
				)
				return
			}
		}
		for _, err2 := range e.ForbiddenErrors {
			if errors.Is(err, err2) {
				ctx.JSON(http.StatusForbidden,
					gin.H{"error": err.Error()},
				)
				return
			}
		}
		for _, typeError := range e.ForbiddenTypeErrors {
			if errors.As(err, typeError) {
				ctx.JSON(http.StatusForbidden,
					gin.H{"error": err.Error()},
				)
				return
			}
		}
		for _, err2 := range e.UnauthorizedErrors {
			if errors.Is(err, err2) {
				ctx.JSON(http.StatusUnauthorized,
					gin.H{"error": err.Error()},
				)
				return
			}
		}
		for _, typeError := range e.UnauthorizedTypeErrors {
			if errors.As(err, typeError) {
				ctx.JSON(http.StatusUnauthorized,
					gin.H{"error": err.Error()},
				)
				return
			}
		}
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		e.Notify(ctx, err)
	}}
}

func (e *DefaultErrorHandler) Notify(ctx *gin.Context, err *gin.Error) {
	if err == nil {
		return
	}
	if e.Notifier == nil {
		e.Logger.Error("notifier is nil")
		return
	}
	e.Notifier.Notify(err, ctx.Request)
}
