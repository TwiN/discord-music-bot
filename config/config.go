package config

import (
	"os"
	"strconv"
	"strings"
)

type Config struct {
	DiscordToken                  string
	CommandPrefix                 string
	MaximumAudioDurationInSeconds int
}

var cfg *Config

// Load initializes the configuration
func Load() {
	cfg = &Config{
		DiscordToken:  strings.TrimSpace(os.Getenv("DISCORD_BOT_TOKEN")),
		CommandPrefix: strings.TrimSpace(os.Getenv("COMMAND_PREFIX")),
	}
	if len(cfg.DiscordToken) == 0 {
		panic("environment variable 'DISCORD_BOT_TOKEN' must not be empty")
	}
	if len(cfg.CommandPrefix) != 1 {
		cfg.CommandPrefix = "!"
	}
	maximumAudioDurationInSeconds, err := strconv.Atoi(strings.TrimSpace(os.Getenv("MAXIMUM_AUDIO_DURATION_IN_SECONDS")))
	if err != nil {
		cfg.MaximumAudioDurationInSeconds = 300
	} else {
		cfg.MaximumAudioDurationInSeconds = maximumAudioDurationInSeconds
	}
}

func Get() *Config {
	return cfg
}
