package api

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/BitCoinOffical/hotel-notepad/auth-service/internal/api/handlers"
	"github.com/gin-gonic/gin"
)

const (
	timeoutSecond = 5
)

type AppConfig struct {
	Port string
}

type Server struct {
	h      *handlers.Handlers
	engine *gin.Engine
	server *http.Server
}

func NewServer(cfg *AppConfig, h *handlers.Handlers) *Server {
	engine := gin.New()
	return &Server{
		engine: engine,
		h:      h,
		server: &http.Server{
			Addr:        ":" + cfg.Port,
			Handler:     engine,
			ReadTimeout: timeoutSecond * time.Second,
		},
	}
}

func (s *Server) Run() {
	api := s.engine.Group("/api/v1")
	auth := api.Group("/auth")
	{
		auth.POST("/register", s.h.Auth.Register)
		auth.POST("/login")
		auth.GET("/logout")
	}
}

func (s *Server) ShutDown(ctx context.Context) error {
	if err := s.server.Shutdown(ctx); err != nil {
		return fmt.Errorf("s.server.Shutdown: %w", err)
	}
	return nil
}
