package config

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
