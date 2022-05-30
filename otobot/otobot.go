package otobot

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/bwmarrin/discordgo"
	"github.com/puristt/discord-bot-go/config"
	"github.com/puristt/discord-bot-go/mux"
	"github.com/puristt/discord-bot-go/youtube"
)

//initializes the discord bot.
func InitOtobot(cfg *config.Config, router *mux.Mux, youtubeAPI *youtube.YoutubeAPI) {
	dcSession, err := discordgo.New("Bot " + cfg.Discord.Token)
	if err != nil {
		log.Fatal("Error while creating bot instance : ", err.Error())
	}

	dcSession.AddHandler(router.OnMessageCreate)

	dcSession.Open()

	log.Println("Otobot is running. Press Ctrl-C to exit.")
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt, os.Kill)
	<-sc

	dcSession.Close()
}
