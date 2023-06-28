package serve

import (
	"net/http"

	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
)

type Controller interface {
	InitRoute(routes gin.IRoutes, path string)
}

type Auth interface {
	AuthMiddleware() gin.HandlerFunc
}

type Middleware interface {
	Handler() []gin.HandlerFunc
}

type OneMiddleware interface {
	Handler() gin.HandlerFunc
}

type HTTPServe struct {
	Controllers map[string]Controller
	Auth        Auth
	BasePath    string
	R           *gin.Engine
	Listen      string
	Middleware  []Middleware
}

func (s *HTTPServe) Init() {
	if s.BasePath == "" {
		s.BasePath = "/"
	}
	if s.Listen == "" {
		s.Listen = ":8080"
	}
	if s.R == nil {
		s.R = gin.Default()
	}
}

func (s *HTTPServe) Run() error {
	group := s.R.Group(s.BasePath)
	if s.Auth != nil {
		group.Use(s.Auth.AuthMiddleware())
	}
	for _, middleware := range s.Middleware {
		group.Use(middleware.Handler()...)
	}

	for p, c := range s.Controllers {
		if c != nil {
			c.InitRoute(group, p)
		}
	}
	s.R.GET("/ping", func(context *gin.Context) {
		context.String(http.StatusOK, "pong")
	})

	err := s.R.Run(s.Listen)
	if err != nil {
		return err
	}
	return nil
}

func (s *HTTPServe) Start() {
	s.Init()
	log.Panic(s.Run())
}
