package otobot

import (
	"log"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/bwmarrin/discordgo"
	"github.com/puristt/discord-bot-go/config"
)

func InitOtobot(cfg *config.Config) {
	dcSession, err := discordgo.New("Bot " + cfg.Discord.Token)
	if err != nil {
		log.Fatal("Error while creating bot instance : ", err.Error())
	}

	dcSession.AddHandler(onMessageCreate)

	dcSession.Open()

	log.Println("Otobot is running. Press Ctrl-C to exit.")
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt, os.Kill)
	<-sc

	dcSession.Close()
}

func onMessageCreate(ds *discordgo.Session, dm *discordgo.MessageCreate) {

	if strings.Contains("blitz", dm.Content) {
		ds.ChannelMessageSend(dm.ChannelID, "Gaza geldim, hizmete hazirim!")
	}
}
