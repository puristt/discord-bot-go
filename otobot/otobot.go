package otobot

import (
	"container/list"
	"fmt"
	"log"

	"github.com/Workiva/go-datastructures/queue"
	"github.com/bwmarrin/discordgo"
	"github.com/puristt/discord-bot-go/config"
	"github.com/puristt/discord-bot-go/youtube"
)

type VoiceInstance struct {
	dvc                 *discordgo.VoiceConnection
	session             *discordgo.Session
	stop                bool
	skip                bool
	isPlaying           bool
	playQueue           *queue.Queue
	errQueue            *queue.Queue
	nowPlayingMessageID string
	playHistoryList     *list.List
}

type SongInstance struct {
	title     string
	artist    string
	songPath  string
	coverUrl  string
	coverPath string
	videoID   string
	duration  string
}

var (
	ytube *youtube.YoutubeAPI
	cfg   *config.Config
	vi    *VoiceInstance
)

//initializes the discord bot.
func InitOtobot(config *config.Config, s *discordgo.Session, youtubeAPI *youtube.YoutubeAPI) {
	ytube = youtubeAPI
	cfg = config

	vi = &VoiceInstance{
		session:             s,
		dvc:                 nil,
		stop:                false,
		skip:                false,
		isPlaying:           false,
		nowPlayingMessageID: "",
		playHistoryList:     list.New(),
	}
}

func PlayRequestedSong(ds *discordgo.Session, dm *discordgo.MessageCreate) {
	vi.validateMessageAndJoinVoiceChannel(ds, dm)
}

//validateMessageAndJoinVoiceChannel validates user message and checks if user is in any voice channel. Then joins
//bot to the voice channel by calling joinVoiceChannel function.
//If bot is joined to the voice channel successfully, returns true; otherwise false.
func (vi *VoiceInstance) validateMessageAndJoinVoiceChannel(ds *discordgo.Session, dm *discordgo.MessageCreate) bool {
	dg, err := vi.getDcGuildByMessage(ds, dm)
	if err != nil {
		log.Println(err)
		return false
	}

	if !vi.isUserInVoiceChannel(dm, dg) {
		log.Printf("Command failed by author : %s, AuthorId : %s. Not in any voice channel", dm.Author.Username, dm.Author.ID)
		_, err := ds.ChannelMessageSend(dm.ChannelID, "You are not in any voice channel!") // TODO : vi.session olarak düzeltilecek.
		if err != nil {
			log.Printf("Error while sending message to channel: %v", err)
		}

		return false
	}
	if !vi.joinVoiceChannel(ds, dm, dg) {
		return false
	}

	return true
}

func (vi *VoiceInstance) getDcGuildByMessage(ds *discordgo.Session, dm *discordgo.MessageCreate) (*discordgo.Guild, error) {
	log.Printf("Message writed by : %s, AuthorId : %s", dm.Author.Username, dm.Author.ID)
	vi.session.ChannelMessageSend(dm.ChannelID, "tamamdir!")
	c, err := ds.State.Channel(dm.ChannelID)
	if err != nil {
		return nil, fmt.Errorf("Could not find channel. \n%v", err)
	}

	dg, err := ds.State.Guild(c.GuildID)
	if err != nil {
		return nil, fmt.Errorf("Could not find guild. \n%v", err)
	}

	return dg, nil
}

//checks if author is in any voice channel
func (vi *VoiceInstance) isUserInVoiceChannel(dm *discordgo.MessageCreate, dg *discordgo.Guild) bool {
	for _, vs := range dg.VoiceStates {
		if vs.UserID == dm.Author.ID {
			return true
		}
	}

	return false
}

//joins bot to the specified voice channel in VoiceInstance.
func (vi *VoiceInstance) joinVoiceChannel(ds *discordgo.Session, dm *discordgo.MessageCreate, dg *discordgo.Guild) bool {
	for _, vs := range dg.VoiceStates {
		if vs.UserID == dm.Author.ID {
			dvc, err := ds.ChannelVoiceJoin(dg.ID, vs.ChannelID, false, false)
			if err != nil {
				fmt.Printf("Could not join the voice channel. Error : %v \n", err)
				if _, ok := ds.VoiceConnections[dg.ID]; ok {
					dvc = ds.VoiceConnections[dg.ID]
				}
			}
			vi.dvc = dvc
			return true
		}
	}

	return false
}
