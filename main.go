package main

import (
	"context"
	"github.com/joho/godotenv"
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

var Router = mux.New()
var youtubeAPI *youtube.YoutubeAPI

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}
	Init()
}

func Init() {
	ctx := context.Background()
	cfg := config.InitConfig()
	youtubeAPI = youtube.NewYoutubeAPI(cfg.Youtube.ApiKey, ctx)

	dcSession, err := discordgo.New("Bot " + cfg.Discord.Token)
	if err != nil {
		log.Printf("Error while creating discord session: %v", err)
	}

	dcSession.AddHandler(Router.OnMessageCreate)

	if err := dcSession.Open(); err != nil { // TODO : webSocket open bug will be fixed
		log.Printf("error while openin dc session : %v", err)
		//dcSession.Close()
	}
	otobot.InitOtobot(dcSession, youtubeAPI)
	log.Println("Otobot is running. Press Ctrl-C to exit.")
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt, os.Kill)
	<-sc

	if err != nil {
		log.Println(err)
	}

	dcSession.Close()
}
