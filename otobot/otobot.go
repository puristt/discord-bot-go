package otobot

import (
	"bufio"
	"container/list"
	"encoding/binary"
	"fmt"

	"github.com/bwmarrin/discordgo"
	"github.com/hraban/opus"
	"github.com/puristt/discord-bot-go/config"
	"github.com/puristt/discord-bot-go/youtube"
	"io"
	"log"
	"os"
	"os/exec"
	"strconv"
)

const (
	channels     int = 2 // 1 for mono, 2 for stereo
	frameRate    int = 48000
	frameSize    int = 960                 // uint16 size of each audio frame
	maxBytes     int = (frameSize * 2) * 2 // max size of opus data
	maxQueueSize int = 50
)

type VoiceInstance struct {
	dvc             *discordgo.VoiceConnection
	session         *discordgo.Session
	stop            bool
	skip            bool
	isPlaying       bool
	playQueue       []Song
	nowPlaying      Song
	playHistoryList *list.List
}

type Song struct { // TODO : naming will be refactored
	title     string
	artist    string
	songPath  string
	videoUrl  string
	coverPath string
	videoID   string
	duration  string
}

var (
	ytube *youtube.YoutubeAPI
	cfg   *config.Config
	vi    *VoiceInstance
)

// InitOtobot initializes the discord bot.
func InitOtobot(config *config.Config, s *discordgo.Session, youtubeAPI *youtube.YoutubeAPI) {
	ytube = youtubeAPI
	cfg = config

	vi = &VoiceInstance{
		session:         s,
		dvc:             nil,
		stop:            false,
		skip:            false,
		isPlaying:       false,
		nowPlaying:      Song{},
		playQueue:       createPlayQueue(),
		playHistoryList: list.New(),
	}
}

func PlaySong(query string, dm *discordgo.MessageCreate) {
	if !vi.validateMessageAndJoinVoiceChannel(dm) {
		return
	}

	res, err := ytube.GetVideoInfo(query)
	if err != nil {
		fmt.Errorf("error occured while downloading %v", err)
	}

	song := Song{
		title:    res.VideoTitle,
		videoID:  res.VideoID,
		videoUrl: res.VideoUrl,
	}
	if len(vi.playQueue) == 0 {
		vi.session.ChannelMessageSend(dm.ChannelID, "tamamdir!")
		vi.playQueue = append(vi.playQueue, song)
		playAudio(song)
	} else {
		vi.playQueue = append(vi.playQueue, song)
		log.Printf("queue : %v ", vi.playQueue)
	}
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

func playAudio(res Song) {
	r, w := io.Pipe()
	defer r.Close()

	ytdl := exec.Command("youtube-dl", "-f", "bestaudio", res.videoUrl, "-o-")
	ytdl.Stdout = w         // youtube-dl PIPE INPUT
	ytdl.Stderr = os.Stderr // show progress
	go func() {
		if err := ytdl.Run(); err != nil {
			log.Printf("WARN: ytdl error: %v", err)
		}
	}()

	ffmpeg := exec.Command("ffmpeg", "-i", "/dev/stdin", "-f", "s16le", "-ar",
		strconv.Itoa(frameRate), "-ac", strconv.Itoa(channels), "pipe:1")
	ffmpeg.Stdin = r // youtube-dl PIPE OUTPUT
	ffmpegOut, err := ffmpeg.StdoutPipe()
	if err != nil {
		log.Printf("StdoutPipe err : %v", err)
	}

	ffmpegBuf := bufio.NewReaderSize(ffmpegOut, frameSize*channels)
	if err := ffmpeg.Start(); err != nil {
		log.Printf("Ffmpeg Pipe run err : %v", err)
		return
	}

	sendChan := make(chan []int16, 2)
	defer close(sendChan)

	go func() {
		sendPCM(vi.dvc, sendChan)
	}()

	vi.isPlaying = true
	for {
		// read data from ffmpeg stdout
		audioBuf := make([]int16, frameSize*channels)
		err = binary.Read(ffmpegBuf, binary.LittleEndian, &audioBuf)
		if err == io.EOF || err == io.ErrUnexpectedEOF {
			vi.playQueue = vi.playQueue[1:]
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
			if err := ytdl.Process.Kill(); err != nil {
				log.Printf("ytdl process killing error : %v", err)
			}
			vi.isPlaying = false
			return
		}

		if vi.skip == true {
			if vi.isPlaying == true {
				if len(vi.playQueue) != 0 {
					vi.skip = false
					/*if err := ytdl.Process.Kill(); err != nil {
						log.Printf("ytdl process killing error : %v", err)
					}*/
					vi.playQueue = vi.playQueue[1:]
					log.Printf("slice capacity : %v  ---- slice length : %v", cap(vi.playQueue), len(vi.playQueue))
				}
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

func createPlayQueue() []Song {
	newQueue := make([]Song, 0, maxQueueSize)
	return newQueue
}
