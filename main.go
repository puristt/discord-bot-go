package main

import (
	"encoding/json"
	"log"
	"os"
	"path/filepath"

	"github.com/puristt/discord-bot-go/config"
	"github.com/puristt/discord-bot-go/otobot"
)

//https://github.com/hemreari/feanor-dcbot

var cfg config.Config

func readConfig(cfg *config.Config, cfgFileName string) {
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

func main() {
	readConfig(&cfg, "config\\config.json")
	otobot.InitOtobot(&cfg)
}
