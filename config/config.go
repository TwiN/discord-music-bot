package config

import (
	"os"
)

type Config struct {
	YoutubeApiKey string
	DiscordToken  string
}

var cfg *Config

func Load() {
	cfg = &Config{
		YoutubeApiKey: os.Getenv("YOUTUBE_API_KEY"),
		DiscordToken:  os.Getenv("DISCORD_BOT_TOKEN"),
	}
	if len(cfg.YoutubeApiKey) == 0 {
		panic("environment variable 'YOUTUBE_API_KEY' must not be empty")
	}
	if len(cfg.DiscordToken) == 0 {
		panic("environment variable 'DISCORD_BOT_TOKEN' must not be empty")
	}
}

func Get() *Config {
	return cfg
}
