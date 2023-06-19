package config

import (
	"os"
)

type Config struct {
	Discord DiscordConfig `json:"discord"`
	Youtube YoutubeConfig `json:"youtube"`
}

type DiscordConfig struct {
	Token string `json:"token"`
}

type YoutubeConfig struct {
	ClientID       string `json:"clientID"`
	ClientSecretID string `json:"clientSecretID"`
	ApiKey         string `json:"apiKey"`
}

func InitConfig() *Config {
	return &Config{
		Discord: DiscordConfig{
			Token: os.Getenv("Discord_Token"),
		},
		Youtube: YoutubeConfig{
			ApiKey: os.Getenv("Youtube_ApiKey"),
		},
	}
}
