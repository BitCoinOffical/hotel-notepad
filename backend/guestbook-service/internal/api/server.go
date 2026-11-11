package api

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/BitCoinOffical/hotel-notepad/notebook-service/internal/api/handlers"
	"github.com/BitCoinOffical/hotel-notepad/notebook-service/internal/api/middleware"
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
		h:      h,
		m:      m,
		engine: engine,
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
	notebook.Use(s.m.AuthMiddleware())
	{
		notebook.POST("/entries", s.h.Guest.CreateEntry)
		notebook.GET("/entries", s.h.Guest.GetEntries)
		notebook.GET("/entries/:guests_id", s.h.Guest.GetEntryByID)
		notebook.PATCH("/entries/:guests_id", s.h.Guest.UpdateEntry)
		notebook.DELETE("/entries/:guests_id", s.h.Guest.UpdateEntry)

		notebook.GET("/entries/search")
	}
}

func (s *Server) ShutDown(ctx context.Context) error {
	if err := s.server.Shutdown(ctx); err != nil {
		return fmt.Errorf("s.server.Shutdown: %w", err)
	}
	return nil
}
