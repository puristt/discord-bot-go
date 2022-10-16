package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/bwmarrin/discordgo"
	"github.com/puristt/discord-bot-go/config"
	"github.com/puristt/discord-bot-go/mux"
	"github.com/puristt/discord-bot-go/otobot"
	"github.com/puristt/discord-bot-go/youtube"
)

//https://github.com/hemreari/feanor-dcbot

var cfg config.Config
var Router = mux.New()
var youtubeAPI *youtube.YoutubeAPI

func main() {
	ctx := context.Background()
	config.InitConfig(&cfg, "config/config.json")

	youtubeAPI = youtube.NewYoutubeAPI(cfg.Youtube.ApiKey, ctx)
	//youtubeAPI.GetSearchResults("boating")
	//youtubeAPI.GetVideoInfo("boatindsgssdagasddgsdgsadgsadgsagdsgdsadgsagdsagsdagdsgdsgdsgsaddsgasdjhfadjgfjhfgh1012g")
	Init()

}

func Init() {
	dcSession, err := discordgo.New("Bot " + cfg.Discord.Token)
	if err != nil {
		log.Printf("Error while creating discord session: %v", err)
	}

	dcSession.AddHandler(Router.OnMessageCreate)

	if err := dcSession.Open(); err != nil { // TODO : webSocket open bug will be fixed
		log.Printf("error while openin dc session : %v", err)
		//dcSession.Close()
	}
	otobot.InitOtobot(&cfg, dcSession, youtubeAPI)
	log.Println("Otobot is running. Press Ctrl-C to exit.")
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt, os.Kill)
	<-sc

	if err != nil {
		log.Println(err)
	}

	dcSession.Close()
}
