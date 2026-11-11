package api

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/BitCoinOffical/hotel-notepad/auth-service/internal/api/handlers"
	"github.com/BitCoinOffical/hotel-notepad/auth-service/internal/api/middleware"
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
	m      *middleware.Middleware
	engine *gin.Engine
	server *http.Server
}

func NewServer(cfg *AppConfig, h *handlers.Handlers, m *middleware.Middleware) *Server {
	engine := gin.New()
	return &Server{
		engine: engine,
		h:      h,
		m:      m,
		server: &http.Server{
			Addr:        ":" + cfg.Port,
			Handler:     engine,
			ReadTimeout: timeoutSecond * time.Second,
		},
	}
}

func (s *Server) Run() error {
	api := s.engine.Group("/api/v1")
	auth := api.Group("/auth")
	{
		auth.POST("/register", s.h.Auth.Register)
		auth.POST("/login", s.h.Auth.Login)
		auth.GET("/logout", s.m.AuthMiddleware(), s.h.Auth.Logout)
	}

	if err := s.server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		return fmt.Errorf("s.server.ListenAndServe: %w", err)
	}
	return nil
}

func (s *Server) ShutDown(ctx context.Context) error {
	if err := s.server.Shutdown(ctx); err != nil {
		return fmt.Errorf("s.server.Shutdown: %w", err)
	}
	return nil
}
