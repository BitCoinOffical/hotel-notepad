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
	api := s.engine.Group("/api/v1")
	notebook := api.Group("/guestbook")
	{
		notebook.POST("/entries")
		notebook.GET("/entries")
		notebook.GET("/entries/:guests_id")
		notebook.PATCH("/entries/:guests_id")
		notebook.DELETE("/entries/:guests_id")
	}
}

func (s *Server) ShutDown() {

}
