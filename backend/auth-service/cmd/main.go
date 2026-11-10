package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/BitCoinOffical/hotel-notepad/auth-service/config"
	postgresdb "github.com/BitCoinOffical/hotel-notepad/auth-service/internal/adapters/postgres"
	redisdb "github.com/BitCoinOffical/hotel-notepad/auth-service/internal/adapters/redis"
	"github.com/BitCoinOffical/hotel-notepad/auth-service/internal/api"
	"github.com/BitCoinOffical/hotel-notepad/auth-service/internal/api/handlers"
	"github.com/BitCoinOffical/hotel-notepad/auth-service/pkg/logger"
	"go.uber.org/zap"
)

const (
	timeoutSecond = 5
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	cfg, err := config.NewLoad()
	if err != nil {
		log.Fatal(err)
	}

	logger, err := logger.NewLogger(string(cfg.App.DebugLevel))
	if err != nil {
		log.Fatal(err)
	}

	logger.Info("config load")
	logger.Info("logger start")

	pool, err := postgresdb.NewPool(ctx, &postgresdb.PostgresConfig{
		DB:   cfg.Postgres.DB,
		User: cfg.Postgres.User,
		Pass: cfg.Postgres.Pass,
		Port: cfg.Postgres.Port,
		Host: cfg.Postgres.Host,
	})
	if err != nil {
		logger.Fatal("failed create new pool", zap.Error(err))
	}
	logger.Info("postgres pool applied successfully")

	rdb, err := redisdb.NewRedis(context.Background(), &redisdb.RedisConfig{
		Host: cfg.Redis.Host,
		Port: cfg.Redis.Port,
		User: cfg.Redis.User,
		Pass: cfg.Redis.Pass,
		DB:   cfg.Redis.DB,
	})
	if err != nil {
		logger.Fatal("failed create new redis client", zap.Error(err))
	}
	logger.Info("redis applied successfully")

	srvc := handlers.NewServices(pool, rdb)
	h := handlers.NewHandlers(srvc, logger)

	server := api.NewServer(&api.AppConfig{
		Port: cfg.App.Port,
	}, h)

	server.Run()

	<-ctx.Done()
	shutdownCtx, cancel := context.WithTimeout(context.Background(), timeoutSecond*time.Second)
	defer cancel()
	logger.Info("received a signal indicating the completion of operations")

	if err := server.ShutDown(shutdownCtx); err != nil {
		logger.Error("failed shut down server", zap.Error(err))
	}
}
