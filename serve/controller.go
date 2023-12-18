package serve

import (
	"github.com/gin-gonic/gin"
)

type BaseController struct {
}

func (b *BaseController) GET(r gin.IRoutes, p string, f func(ctx *gin.Context) error) {
	r.GET(p, func(ctx *gin.Context) {
		if err := f(ctx); err != nil {
			b.Abort(ctx, err)
		}
	})
}

func (b *BaseController) POST(r gin.IRoutes, p string, f func(ctx *gin.Context) error) {
	r.POST(p, func(ctx *gin.Context) {
		if err := f(ctx); err != nil {
			b.Abort(ctx, err)
		}
	})
}

func (b *BaseController) PATCH(r gin.IRoutes, p string, f func(ctx *gin.Context) error) {
	r.PATCH(p, func(ctx *gin.Context) {
		if err := f(ctx); err != nil {
			b.Abort(ctx, err)
		}
	})
}

func (b *BaseController) DELETE(r gin.IRoutes, p string, f func(ctx *gin.Context) error) {
	r.DELETE(p, func(ctx *gin.Context) {
		if err := f(ctx); err != nil {
			b.Abort(ctx, err)
		}
	})
}

func (b *BaseController) PUT(r gin.IRoutes, p string, f func(ctx *gin.Context) error) {
	r.PUT(p, func(ctx *gin.Context) {
		if err := f(ctx); err != nil {
			b.Abort(ctx, err)
		}
	})
}

func (b *BaseController) Abort(ctx *gin.Context, err error) {
	ctx.Abort()
	// logs.Printf("%+v", err)
	_ = ctx.Error(err)
}
