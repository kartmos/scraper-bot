package main

import (
	"github.com/kartmos/scraper-bot/internal/bot"
	"github.com/spf13/viper"
)

func init() {
	configPath := "/app/config/config.yaml"
	viper.SetConfigFile(configPath)
}

func main() {
	bot.StartBot()
}
