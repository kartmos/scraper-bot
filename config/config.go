package config

import (
	"fmt"
	"os"
	"strings"

	"github.com/spf13/viper"
)

type AppConfig struct {
	BotToken    string   `mapstructure:"bot_token"`
	RapidTokens []string `mapstructure:"rapid_tokens"`
	AdminID     int64    `mapstructure:"admin_id"`
	DownloadDir string   `mapstructure:"download_dir"`
	YTDLPPath   string   `mapstructure:"yt_dlp_path"`
	HelpFile    string   `mapstructure:"help_file"`
	WelcomeFile string   `mapstructure:"welcome_file"`
	CommandFile string   `mapstructure:"command_file"`
	LogFile     string   `mapstructure:"log_file"`
}

func Load(configPath string) (AppConfig, error) {
	v := viper.New()
	v.SetConfigFile(configPath)
	v.SetConfigType("yaml")
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	v.AutomaticEnv()

	setDefaults(v)

	if _, err := os.Stat(configPath); err == nil {
		if err := v.ReadInConfig(); err != nil {
			return AppConfig{}, fmt.Errorf("read config: %w", err)
		}
	}

	cfg := AppConfig{
		BotToken:    v.GetString("bot_token"),
		RapidTokens: parseRapidTokens(v),
		AdminID:     v.GetInt64("admin_id"),
		DownloadDir: v.GetString("download_dir"),
		YTDLPPath:   v.GetString("yt_dlp_path"),
		HelpFile:    v.GetString("help_file"),
		WelcomeFile: v.GetString("welcome_file"),
		CommandFile: v.GetString("command_file"),
		LogFile:     v.GetString("log_file"),
	}

	if cfg.BotToken == "" {
		return AppConfig{}, fmt.Errorf("bot token is empty (set bot_token in config or BOT_TOKEN env)")
	}

	return cfg, nil
}

func setDefaults(v *viper.Viper) {
	v.SetDefault("bot_token", os.Getenv("BOT_TOKEN"))
	v.SetDefault("rapid_tokens", []string{})
	v.SetDefault("admin_id", int64(291182090))
	v.SetDefault("download_dir", "./downloads")
	v.SetDefault("yt_dlp_path", "./bin/yt-dlp")
	v.SetDefault("help_file", "./asserts/help.txt")
	v.SetDefault("welcome_file", "./asserts/welcome.txt")
	v.SetDefault("command_file", "./asserts/command.txt")
	v.SetDefault("log_file", "./logs/bot.log")
}

func parseRapidTokens(v *viper.Viper) []string {
	raw := strings.TrimSpace(v.GetString("rapid_tokens"))
	if raw == "" {
		return v.GetStringSlice("rapid_tokens")
	}

	parts := strings.Split(raw, ",")
	tokens := make([]string, 0, len(parts))
	for _, part := range parts {
		token := strings.TrimSpace(part)
		if token != "" {
			tokens = append(tokens, token)
		}
	}

	return tokens
}
