package config

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLoad_ParsesRapidTokensFromEnv(t *testing.T) {
	t.Setenv("BOT_TOKEN", "bot-token")
	t.Setenv("RAPID_TOKENS", "token1, token2 ,token3")

	cfg, err := Load("missing-config.yaml")
	require.NoError(t, err)

	assert.Equal(t, []string{"token1", "token2", "token3"}, cfg.RapidTokens)
	assert.Equal(t, "./bin/yt-dlp", cfg.YTDLPPath)
	assert.Equal(t, "./logs/bot.log", cfg.LogFile)
}

func TestLoad_ParsesRapidTokensFromYAML(t *testing.T) {
	t.Setenv("BOT_TOKEN", "bot-token")

	file, err := os.CreateTemp(t.TempDir(), "config-*.yaml")
	require.NoError(t, err)
	t.Cleanup(func() {
		_ = file.Close()
	})

	_, err = file.WriteString("rapid_tokens:\n  - token-a\n  - token-b\n")
	require.NoError(t, err)

	cfg, err := Load(file.Name())
	require.NoError(t, err)

	assert.Equal(t, []string{"token-a", "token-b"}, cfg.RapidTokens)
}
