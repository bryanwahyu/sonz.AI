package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/heroiclabs/nakama/v3/apigrpc"
	"github.com/heroiclabs/nakama/v3/src/app/auth"
	"github.com/heroiclabs/nakama/v3/src/app/battles"
	"github.com/heroiclabs/nakama/v3/src/app/bot"
	"github.com/heroiclabs/nakama/v3/src/app/groups"
	leaderboardsvc "github.com/heroiclabs/nakama/v3/src/app/leaderboard"
	nakamainfra "github.com/heroiclabs/nakama/v3/src/infra/nakama"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type Config struct {
	HTTPAddress       string
	NakamaGRPCAddress string
}

func loadConfig() Config {
	cfg := Config{
		HTTPAddress:       getEnv("SANDAI_HTTP_ADDR", ":8080"),
		NakamaGRPCAddress: getEnv("SANDAI_NAKAMA_GRPC_ADDR", "127.0.0.1:7349"),
	}
	return cfg
}

func main() {
	logger, err := zap.NewProduction()
	if err != nil {
		panic(err)
	}
	defer func() { _ = logger.Sync() }()

	cfg := loadConfig()

	baseCtx, cancel := signal.NotifyContext(context.Background(), syscall.SIGTERM, syscall.SIGINT)
	defer cancel()

	shutdownTelemetry, err := setupTelemetry(baseCtx, "sandai-api")
	if err != nil {
		logger.Warn("failed to initialize telemetry", zap.Error(err))
	} else {
		defer func() {
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()
			_ = shutdownTelemetry(ctx)
		}()
	}

	conn, err := grpc.DialContext(baseCtx, cfg.NakamaGRPCAddress, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		logger.Fatal("failed to dial nakama", zap.Error(err))
	}
	defer conn.Close()

	nakamaClient := apigrpc.NewNakamaClient(conn)

	playerRepo := &nakamainfra.PlayerRepository{Client: nakamaClient}
	authProvider := &nakamainfra.AuthClient{Client: nakamaClient}
	groupRepo := &nakamainfra.GroupRepository{Client: nakamaClient}
	matchRepo := &nakamainfra.BattleRepository{Client: nakamaClient}
	leaderboardRepo := &nakamainfra.LeaderboardRepository{Client: nakamaClient}
	botRepo := &nakamainfra.BotRepository{Client: nakamaClient}
	botQueue := &nakamainfra.BotQueue{}
	notifier := &nakamainfra.NotificationClient{Client: nakamaClient}
	matchProvider := &nakamainfra.MatchClient{Client: nakamaClient}
	groupProvider := &nakamainfra.GroupClient{Client: nakamaClient}

	authService := auth.NewService(playerRepo, authProvider)
	groupService := groups.NewService(groupRepo, groupProvider)
	battleService := battles.NewService(matchRepo, matchProvider)
	leaderboardService := leaderboardsvc.NewService(leaderboardRepo)
	botService := bot.NewService(botRepo, botQueue, notifier)

	server := NewServer(ServerConfig{
		Logger:             logger,
		AuthService:        authService,
		GroupService:       groupService,
		BattleService:      battleService,
		LeaderboardService: leaderboardService,
		BotService:         botService,
	})

	httpServer := &http.Server{
		Addr:         cfg.HTTPAddress,
		Handler:      server.Handler(),
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	go func() {
		logger.Info("Sand-ai API listening", zap.String("addr", cfg.HTTPAddress))
		if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatal("http server error", zap.Error(err))
		}
	}()

	<-baseCtx.Done()
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer shutdownCancel()
	if err := httpServer.Shutdown(shutdownCtx); err != nil {
		logger.Error("graceful shutdown failed", zap.Error(err))
	}
}

func getEnv(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}
