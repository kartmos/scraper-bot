// cSpell:disable
package config

import (
	"log"

	"github.com/fsnotify/fsnotify"
	"github.com/spf13/viper"
)

var Config = &TokenConfig{}

type TokenConfig struct {
	RapidToken []string
}

func (t *TokenConfig) CheckConfig() {
	viper.OnConfigChange(func(e fsnotify.Event) {
		if e.Op&fsnotify.Write == fsnotify.Write {
			if err := viper.ReadInConfig(); err != nil {
				log.Printf("[WARN] Failed to read config file: %v", err)
				return
			}
			if viper.IsSet("botToken") {
				t.RapidToken = viper.GetStringSlice("rapidToken")
			} else {
				log.Println("[CheckConfig] Key 'botToken' not found in config")
			}
		}
	})
	viper.WatchConfig()

	if err := viper.ReadInConfig(); err != nil {
		log.Printf("[CheckConfig] Failed to read config file: %v", err)
		return
	}

	t.RapidToken = viper.GetStringSlice("rapidToken")
}
