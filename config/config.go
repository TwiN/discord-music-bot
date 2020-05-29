package config

import (
	"os"
)

type Config struct {
	DiscordToken string
}

var cfg *Config

func Load() {
	cfg = &Config{
		DiscordToken: os.Getenv("DISCORD_BOT_TOKEN"),
	}
	if len(cfg.DiscordToken) == 0 {
		panic("environment variable 'DISCORD_BOT_TOKEN' must not be empty")
	}
}

func Get() *Config {
	return cfg
}
