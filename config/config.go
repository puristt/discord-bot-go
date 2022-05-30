package config

import (
	"encoding/json"
	"log"
	"os"
	"path/filepath"
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

func InitConfig(cfg *Config, cfgFileName string) {
	configFileName, _ := filepath.Abs(cfgFileName)
	log.Printf("Loading config : %v", configFileName)

	configFile, err := os.Open(configFileName)
	if err != nil {
		log.Fatal("Error while openning file : ", err.Error())
	}

	defer configFile.Close()

	jsonParser := json.NewDecoder(configFile)
	if err := jsonParser.Decode(&cfg); err != nil {
		log.Fatal("Config Decoding error : ", err.Error())
	}
}
