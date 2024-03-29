package otobot

import (
	"bufio"
	"encoding/binary"
	"fmt"
	"github.com/bwmarrin/discordgo"
	"github.com/puristt/discord-bot-go/model"
	"github.com/puristt/discord-bot-go/queue"
	"github.com/puristt/discord-bot-go/util"
	"github.com/puristt/discord-bot-go/youtube"
	"gopkg.in/hraban/opus.v2"
	"io"
	"log"
	"os"
	"os/exec"
	"strconv"
	"time"
)

const (
	channels     int   = 2 // 1 for mono, 2 for stereo
	frameRate    int   = 48000
	frameSize    int   = 960                 // uint16 size of each audio frame
	maxBytes     int   = (frameSize * 2) * 2 // max size of opus data
	maxQueueSize int64 = 50
)

type VoiceInstance struct {
	dvc               *discordgo.VoiceConnection
	session           *discordgo.Session
	stop              bool
	skip              bool
	isPlaying         bool
	isPlaylistPlaying bool
	playQueue         *queue.Queue
	nowPlaying        model.Song
}

var (
	yTube             *youtube.YoutubeAPI
	vi                *VoiceInstance
	nextPageToken     = "" // YouTube playlist nextPage Token
	currentPlaylistId = "" // YouTube current playlistId
)

// InitOtobot initializes the discord bot.
func InitOtobot(s *discordgo.Session, youtubeAPI *youtube.YoutubeAPI) {
	yTube = youtubeAPI

	vi = &VoiceInstance{
		session:           s,
		dvc:               nil,
		stop:              false,
		skip:              false,
		isPlaying:         false,
		isPlaylistPlaying: false,
		nowPlaying:        model.Song{},
		playQueue:         createPlayQueue(),
	}
}

func PlayPlaylist(url string, dm *discordgo.MessageCreate) {
	if !vi.validateMessageAndJoinVoiceChannel(dm) {
		return
	}

	// when "-play some playlist" command has run, it disposes the play queue and starts to play playlist immediately
	if !vi.playQueue.Empty() && vi.isPlaying == true {
		if vi.isPlaylistPlaying {
			vi.session.ChannelMessageSend(dm.ChannelID, "A playlist is already playing. Stop the playlist first!")
			return
		} else {
			StopSong(dm)
		}
	}

	playlistId := util.ExtractYoutubePlaylistId(url)
	if playlistId == "" {
		vi.session.ChannelMessageSend(dm.ChannelID, "Given playlist URL is not valid!")
		return
	}

	vi.playPlaylist(playlistId, dm.ChannelID)
}

func PlaySong(query string, dm *discordgo.MessageCreate) {
	if !vi.validateMessageAndJoinVoiceChannel(dm) {
		return
	}

	res, err := yTube.GetVideoInfo(query) // TODO : this method uses a lot of cost https://developers.google.com/youtube/v3/determine_quota_cost
	if err != nil {
		vi.session.ChannelMessageSend(dm.ChannelID, err.Error())
		return
	}

	song := model.Song{
		Title:    res.VideoTitle,
		VideoID:  res.VideoID,
		VideoUrl: res.VideoUrl,
		Duration: res.Duration,
		ImageUrl: res.VideoImageUrl,
	}
	if vi.isPlaylistPlaying {
		vi.session.ChannelMessageSend(dm.ChannelID, "A playlist already playing. Stop the playlist first!")
		return
	}

	if vi.playQueue.Empty() {
		vi.session.ChannelMessageSend(dm.ChannelID, "Got it!")
		vi.playQueue.Enqueue(song) // TODO : if playAudio method returns an error, song should not be enqueued
		go vi.playQueueFunc(dm.ChannelID)
	} else {
		vi.playQueue.Enqueue(song)
		vi.createAndSendEmbedAddedToQueueMessage(song, dm.ChannelID)
	}

	return
}

func SearchSong(query string, dm *discordgo.MessageCreate) {
	results, err := yTube.GetSearchResults(query)
	if err != nil {
		vi.session.ChannelMessageSend(dm.ChannelID, err.Error())
	}

	var songs []model.Song
	for _, res := range results { // mapping YouTube search results to song struct
		song := model.Song{
			Title:    res.VideoTitle,
			VideoID:  res.VideoID,
			VideoUrl: res.VideoUrl,
			Duration: res.Duration,
			ImageUrl: res.VideoImageUrl,
		}
		songs = append(songs, song)
	}

	err = vi.createAndSendEmbedShowSearchResultsMessage(songs, dm.ChannelID)
	if err != nil {
		log.Printf("error while showing search results : %v", err)
		return
	}
	return
}

func StopSong(dm *discordgo.MessageCreate) {
	if vi.isPlaying == false {
		return
	}

	g, err := vi.getDcGuildByMessage(dm)
	if err != nil {
		log.Printf("Error occured while getting Guild : %v", err)
		return
	}

	for _, vs := range g.VoiceStates {
		if vs.UserID == dm.Author.ID {
			vi.stop = true // set stop flag true
			return
		}
	}
}

func SkipSong(dm *discordgo.MessageCreate) {
	if vi.isPlaying == false {
		return
	}

	g, err := vi.getDcGuildByMessage(dm)
	if err != nil {
		log.Printf("Error occured while getting Guild : %v", err)
		return
	}

	for _, vs := range g.VoiceStates {
		if vs.UserID == dm.Author.ID {
			vi.skip = true // set skip flag true
			return
		}
	}
}

func ShowPlayQueue(dm *discordgo.MessageCreate) {
	if vi.playQueue.Empty() {
		vi.session.ChannelMessageSend(dm.ChannelID, "Play queue is empty.")
		return
	}

	songs := vi.playQueue.PeekAll()
	err := vi.createAndSendEmbedShowPlayQueueMessage(songs, dm.ChannelID)
	if err != nil {
		log.Printf("error while Show embed playqueue message : %v", err)
		return
	}
}

func ShowHelpDialog(dm *discordgo.MessageCreate) {
	vi.createAndSendEmbedHelpMessage(dm.ChannelID)
}

// validateMessageAndJoinVoiceChannel validates user message and checks if user is in any voice channel. Then joins
// bot to the voice channel by calling joinVoiceChannel function.
// If bot is joined to the voice channel successfully, returns true; otherwise false.
func (vi *VoiceInstance) validateMessageAndJoinVoiceChannel(dm *discordgo.MessageCreate) bool {
	dg, err := vi.getDcGuildByMessage(dm)
	if err != nil {
		log.Println(err)
		return false
	}

	if !vi.isUserInVoiceChannel(dm, dg) {
		log.Printf("Command failed by author : %s, AuthorId : %s. Not in any voice channel", dm.Author.Username, dm.Author.ID)
		_, err := vi.session.ChannelMessageSend(dm.ChannelID, "You are not in any voice channel!")
		if err != nil {
			log.Printf("Error while sending message to channel: %v", err)
		}

		return false
	}

	if !vi.joinVoiceChannel(dm, dg) {
		return false
	}

	return true
}

func (vi *VoiceInstance) getDcGuildByMessage(dm *discordgo.MessageCreate) (*discordgo.Guild, error) {
	c, err := vi.session.State.Channel(dm.ChannelID)
	if err != nil {
		return nil, fmt.Errorf("Could not find channel. \n%v", err)
	}

	dg, err := vi.session.State.Guild(c.GuildID)
	if err != nil {
		return nil, fmt.Errorf("Could not find guild. \n%v", err)
	}

	return dg, nil
}

// checks if author is in any voice channel
func (vi *VoiceInstance) isUserInVoiceChannel(dm *discordgo.MessageCreate, dg *discordgo.Guild) bool {
	for _, vs := range dg.VoiceStates {
		if vs.UserID == dm.Author.ID {
			return true
		}
	}

	return false
}

// joins bot to the specified voice channel in VoiceInstance.
func (vi *VoiceInstance) joinVoiceChannel(dm *discordgo.MessageCreate, dg *discordgo.Guild) bool {
	for _, vs := range dg.VoiceStates {
		if vs.UserID == dm.Author.ID {
			dvc, err := vi.session.ChannelVoiceJoin(dg.ID, vs.ChannelID, false, false)
			if err != nil {
				fmt.Printf("Could not join the voice channel. Error : %v \n", err)
				if _, ok := vi.session.VoiceConnections[dg.ID]; ok {
					dvc = vi.session.VoiceConnections[dg.ID]
				}
			}
			vi.dvc = dvc
			return true
		}
	}

	return false
}

func (vi *VoiceInstance) playPlaylist(playlistId string, channelId string) {
	playListItems, pageToken, err := yTube.GetPlaylistItems(playlistId, nextPageToken)
	if err != nil {
		fmt.Errorf("error occured while getting playlist items %v", err)
		return
	}

	nextPageToken = pageToken
	if currentPlaylistId == "" {
		currentPlaylistId = playlistId

	}
	for _, item := range playListItems {
		song := model.Song{
			Title:    item.VideoTitle,
			VideoID:  item.VideoID,
			VideoUrl: item.VideoUrl,
			Duration: item.Duration,
		}

		if vi.playQueue.Empty() {
			vi.playQueue.Enqueue(song) // TODO : if playAudio method returns an error, song should not be enqueued
			vi.isPlaylistPlaying = true
			go vi.playQueueFunc(channelId)
		} else {
			vi.playQueue.Enqueue(song)
		}
	}
}

func (vi *VoiceInstance) playQueueFunc(channelID string) {
	if err := vi.dvc.Speaking(true); err != nil {
		log.Printf("Bot set speaking err : %v", err)
	}

	defer func() {
		vi.disconnectOtobot()
	}()
	for {
		if vi.isPlaying == false && vi.playQueue.Empty() {
			return
		}

		if vi.isPlaying == false && !vi.playQueue.Empty() { // TODO : after -skip command, it plays 2 song simultaneously. Will be fixed
			vi.isPlaying = true
			vi.processPlayQueue(channelID)
		}

		if vi.isPlaylistPlaying == true && nextPageToken != "" && vi.playQueue.Len() == 1 {
			vi.playPlaylist(currentPlaylistId, channelID)
		}
	}
}

func (vi *VoiceInstance) processPlayQueue(channelID string) {
	vi.nowPlaying = vi.playQueue.Front()

	if err := vi.createAndSendEmbedNowPlayingMessage(&vi.nowPlaying, channelID); err != nil {
		log.Println(err)
	}

	vi.playAudio(vi.nowPlaying)
}

func (vi *VoiceInstance) playAudio(res model.Song) {
	r, w := io.Pipe()

	ytdl := exec.Command("yt-dlp", "-f", "bestaudio", res.VideoUrl, "-o-")
	ytdl.Stdout = w         // youtube-dl PIPE INPUT
	ytdl.Stderr = os.Stderr // show progress
	go func() {
		if err := ytdl.Run(); err != nil {
			log.Printf("WARN: ytdl error: %v", err)
		}
		log.Println("ytdl run command finished!")
		defer r.Close()
	}()

	ffmpeg := exec.Command("ffmpeg", "-i", "/dev/stdin", "-f", "s16le", "-ar",
		strconv.Itoa(frameRate), "-ac", strconv.Itoa(channels), "pipe:1")
	ffmpeg.Stdin = r // youtube-dl PIPE OUTPUT
	ffmpegOut, err := ffmpeg.StdoutPipe()
	if err != nil {
		log.Printf("StdoutPipe err : %v", err)
	}

	ffmpegBuf := bufio.NewReaderSize(ffmpegOut, 16384)
	if err := ffmpeg.Start(); err != nil {
		log.Printf("Ffmpeg Pipe run err : %v", err)
		return
	}

	sendChan := make(chan []int16, 2)
	defer close(sendChan)

	go func() {
		sendPCM(vi.dvc, sendChan)
	}()

	for {
		// read data from ffmpeg stdout
		audioBuf := make([]int16, frameSize*channels)
		err = binary.Read(ffmpegBuf, binary.LittleEndian, &audioBuf)
		if err == io.EOF || err == io.ErrUnexpectedEOF {
			vi.playQueue.Dequeue()
			vi.isPlaying = false

			if vi.playQueue.Empty() {
				if err := ffmpeg.Process.Kill(); err != nil {
					log.Printf("ffmpeg process killing error : %v", err)
				}
			}
			return
		}
		if err != nil {
			log.Printf("error reading from ffmpeg stdout %v", err)
			return
		}

		// Send received PCM to the sendPCM channel
		select {
		case sendChan <- audioBuf:
		}

		if vi.stop == true {
			vi.stop = false

			vi.isPlaylistPlaying = false
			vi.isPlaying = false
			nextPageToken = ""
			currentPlaylistId = ""
			if err := ffmpeg.Process.Kill(); err != nil {
				log.Printf("ytdl process killing error : %v", err)
			}
			vi.playQueue.Dispose()

			return
		}

		if vi.skip == true {
			vi.skip = false
			vi.playQueue.Dequeue()

			if vi.isPlaying == true && !vi.playQueue.Empty() {
				vi.isPlaying = false
				if err := ffmpeg.Process.Kill(); err != nil {
					log.Printf("ffmpeg process killing error : %v", err)
				}

				return
			}

			if vi.isPlaying == true && vi.playQueue.Empty() {
				vi.isPlaying = false
				vi.isPlaylistPlaying = false
				nextPageToken = ""
				currentPlaylistId = ""

				return
			}
		}
	}
}

func sendPCM(dvc *discordgo.VoiceConnection, pcm <-chan []int16) {
	if pcm == nil {
		return
	}

	opusEnc, err := opus.NewEncoder(frameRate, channels, opus.AppAudio)
	if err != nil {
		log.Printf("Error while creating opus encoder : %v", err)
		return
	}

	for {
		rcv, ok := <-pcm
		if !ok {
			log.Println("PCM channel closed")
			return
		}

		opusData := make([]byte, maxBytes)
		n, err := opusEnc.Encode(rcv, opusData)
		if err != nil {
			log.Printf("Error while encoding pcm data : %v", err)
			return
		}
		opusData = opusData[:n] // only the first N bytes are opus data. Just like io.Reader.
		if dvc.Ready == false || dvc.OpusSend == nil {
			return
		}

		dvc.OpusSend <- opusData
	}
}

// disconnectBot disconnects bot from the voice channel
func (vi *VoiceInstance) disconnectOtobot() {
	err := vi.dvc.Speaking(false)
	if err != nil {
		log.Println("Couldn't stop speaking", err)
	}
	vi.dvc.Disconnect()
	log.Printf("Bot disconnected from the voice channel.\n")
	vi.stop = false
	vi.isPlaying = false
	vi.isPlaylistPlaying = false
	nextPageToken = ""
	currentPlaylistId = ""
	return
}

func (vi *VoiceInstance) createAndSendEmbedAddedToQueueMessage(song model.Song, channelID string) error {
	embedMsg := &discordgo.MessageEmbed{
		Author: &discordgo.MessageEmbedAuthor{},
		Color:  0x145DA0,
		Fields: []*discordgo.MessageEmbedField{
			{
				Name:   "Added to Play Queue!",
				Value:  song.Title + "\n" + " Duration : " + song.Duration,
				Inline: false,
			},
		},
		Timestamp: time.Now().Format(time.RFC3339),
		Image: &discordgo.MessageEmbedImage{
			URL: song.ImageUrl,
		},
	}

	_, err := vi.session.ChannelMessageSendEmbed(channelID, embedMsg)
	if err != nil {
		return fmt.Errorf("error occured while sending now playing embed message : %v", err)
	}
	return nil
}

func (vi *VoiceInstance) createAndSendEmbedHelpMessage(channelID string) error {
	embedMsg := &discordgo.MessageEmbed{
		Author: &discordgo.MessageEmbedAuthor{},
		Color:  0x145DA0,
		Fields: []*discordgo.MessageEmbedField{
			{
				Name:   "-play Command",
				Value:  "Searches and plays the first song with given text after '-play' prefix",
				Inline: false,
			},
			{
				Name:   "-search Command",
				Value:  "Searches first 10 result from YouTube",
				Inline: false,
			},
			{
				Name:   "-showq Command",
				Value:  "Shows play queue",
				Inline: false,
			},
			{
				Name:   "-skip Command",
				Value:  "Skips playing song",
				Inline: false,
			},
			{
				Name:   "-stop Command",
				Value:  "Stops playing song and resets play queue",
				Inline: false,
			},
		},
	}

	_, err := vi.session.ChannelMessageSendEmbed(channelID, embedMsg)
	if err != nil {
		return fmt.Errorf("error occured while sending now playing embed message : %v", err)
	}
	return nil
}

// TODO : this will be refactored. It should create embed message once and update every time
func (vi *VoiceInstance) createAndSendEmbedShowPlayQueueMessage(songs []model.Song, channelID string) error {
	embedMsg := &discordgo.MessageEmbed{
		Title:     "Play Queue:",
		Color:     0xff5733,
		Fields:    createMessageEmbedFields(songs),
		Timestamp: time.Now().Format(time.RFC3339),
	}

	_, err := vi.session.ChannelMessageSendEmbed(channelID, embedMsg)
	if err != nil {
		return fmt.Errorf("error occured while sending now playing embed message : %v", err)
	}
	return nil
}

func (vi *VoiceInstance) createAndSendEmbedNowPlayingMessage(song *model.Song, channelID string) error {
	embedMsg := &discordgo.MessageEmbed{
		Author: &discordgo.MessageEmbedAuthor{},
		Color:  0x26e232,
		Fields: []*discordgo.MessageEmbedField{
			{
				Name:   "Now Playing",
				Value:  song.Title + "\n" + " Duration : " + song.Duration,
				Inline: false,
			},
		},
		Timestamp: time.Now().Format(time.RFC3339),
		Image: &discordgo.MessageEmbedImage{
			URL: song.ImageUrl,
		},
	}

	_, err := vi.session.ChannelMessageSendEmbed(channelID, embedMsg)
	if err != nil {
		return fmt.Errorf("error occured while sending now playing embed message : %v", err)
	}
	return nil
}

func (vi *VoiceInstance) createAndSendEmbedShowSearchResultsMessage(songs []model.Song, channelID string) error {
	embedMsg := &discordgo.MessageEmbed{
		Author:      &discordgo.MessageEmbedAuthor{},
		Title:       "Search Results:",
		Description: "First 10 Results Are Showing",
		Color:       0x3498DB,
		Fields:      createMessageEmbedFields(songs),
		Timestamp:   time.Now().Format(time.RFC3339),
	}

	_, err := vi.session.ChannelMessageSendEmbed(channelID, embedMsg)
	if err != nil {
		return fmt.Errorf("error occured while sending search results embed message : %v", err)
	}
	return nil
}

func createMessageEmbedFields(songs []model.Song) []*discordgo.MessageEmbedField {
	var msgEmbedFields []*discordgo.MessageEmbedField

	for i, song := range songs {
		i++
		embedField := &discordgo.MessageEmbedField{
			Name:   strconv.Itoa(i) + ")  " + song.Title + "\n" + " Duration : " + song.Duration,
			Value:  "Video Url : " + song.VideoUrl + "\n",
			Inline: false,
		}
		msgEmbedFields = append(msgEmbedFields, embedField)
	}
	return msgEmbedFields
}

func createPlayQueue() *queue.Queue {
	return queue.New(maxQueueSize)
}
