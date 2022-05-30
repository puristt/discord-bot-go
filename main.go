package main

import (
	"context"

	"github.com/puristt/discord-bot-go/config"
	"github.com/puristt/discord-bot-go/mux"
	"github.com/puristt/discord-bot-go/otobot"
	"github.com/puristt/discord-bot-go/youtube"
)

//https://github.com/hemreari/feanor-dcbot

var cfg config.Config
var Router = mux.New()

func main() {
	ctx := context.Background()
	config.InitConfig(&cfg, "config\\config.json")

	youtubeAPI := youtube.NewYoutubeAPI(cfg.Youtube.ApiKey, ctx)
	otobot.InitOtobot(&cfg, Router, youtubeAPI)
}
