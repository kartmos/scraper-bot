package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/kartmos/scraper-bot/config"
	rapi "github.com/kartmos/scraper-bot/internal/repository/rapid"
	vrepo "github.com/kartmos/scraper-bot/internal/repository/video"
	"github.com/kartmos/scraper-bot/internal/service"
	"github.com/kartmos/scraper-bot/internal/transport/telegram"
)

func main() {
	cfgPath := os.Getenv("APP_CONFIG_PATH")
	if cfgPath == "" {
		cfgPath = "config/config.yaml"
	}

	cfg, err := config.Load(cfgPath)
	if err != nil {
		log.Fatalf("failed to load config: %v", err)
	}

	client, err := telegram.NewClient(cfg.BotToken)
	if err != nil {
		log.Fatalf("failed to create telegram client: %v", err)
	}

	reelResolver := rapi.NewReelResolver(cfg.RapidTokens, 5*time.Second)
	videoRepo := vrepo.NewRepository(cfg.DownloadDir, cfg.YTDLPPath, reelResolver)

	botService := service.NewBotService(service.Params{
		Chat:   client,
		Videos: videoRepo,
		Settings: service.Settings{
			AdminID:     cfg.AdminID,
			HelpFile:    cfg.HelpFile,
			WelcomeFile: cfg.WelcomeFile,
			CommandFile: cfg.CommandFile,
			LogFile:     cfg.LogFile,
		},
		Started: time.Now(),
	})

	handler := telegram.NewHandler(botService)

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	if err := client.Run(ctx, handler); err != nil {
		log.Fatalf("telegram client stopped with error: %v", err)
	}
}
