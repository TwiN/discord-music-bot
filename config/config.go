package config

import (
	"os"
	"strconv"
)

type Config struct {
	DiscordToken                  string
	MaximumAudioDurationInSeconds int
}

var cfg *Config

func Load() {
	cfg = &Config{
		DiscordToken: os.Getenv("DISCORD_BOT_TOKEN"),
	}
	if len(cfg.DiscordToken) == 0 {
		panic("environment variable 'DISCORD_BOT_TOKEN' must not be empty")
	}
	maximumAudioDurationInSeconds, err := strconv.Atoi(os.Getenv("MAXIMUM_AUDIO_DURATION_IN_SECONDS"))
	if err != nil {
		cfg.MaximumAudioDurationInSeconds = 300
	} else {
		cfg.MaximumAudioDurationInSeconds = maximumAudioDurationInSeconds
	}
}

func Get() *Config {
	return cfg
}
