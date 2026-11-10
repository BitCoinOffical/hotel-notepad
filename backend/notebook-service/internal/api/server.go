package api

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

const (
	timeoutSecond = 5
)

type AppConfig struct {
	Port string
}

type Server struct {
	engine *gin.Engine
	server *http.Server
}

func NewServer(cfg *AppConfig) *Server {
	engine := gin.New()
	return &Server{
		engine: &gin.Engine{},
		server: &http.Server{
			Addr:        ":" + cfg.Port,
			Handler:     engine,
			ReadTimeout: timeoutSecond * time.Second,
		},
	}
}

func (s *Server) Run() {
	s.engine.Group("/api/v1")

}

func (s *Server) ShutDown() {

}
