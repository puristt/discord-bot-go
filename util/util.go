package util

import (
	"log"
	"regexp"
	"strconv"
	"time"
)

const (
	YoutubeSongPath = "youtube/song"
)

var (
	durationRegex = `P(?P<years>\d+Y)?(?P<months>\d+M)?(?P<days>\d+D)?T?(?P<hours>\d+H)?(?P<minutes>\d+M)?(?P<seconds>\d+S)?`

	// YouTube url regex
	yTubeUrlRegex = `^(?:https?\:\/\/)?(?:www\.)?(?:(?:youtube\.com\/watch\?v=)|(?:youtu.be\/))([a-zA-Z0-9\-_]{11})+.*$|^(?:https:\/\/www.youtube.com\/playlist\?list=)([a-zA-Z0-9\-_].*).*$`
	// YouTube playlist url regex
	yTubePlaylistUrlRegex = `^(?:https:\/\/www.youtube.com\/playlist\?list=)([a-zA-Z0-9\-_]{34}).*$`
)

// ParseISO8601 takes a duration in format ISO8601 and parses to
// MM:SS format.
func ParseISO8601(duration string) string {
	r, err := regexp.Compile(durationRegex)
	if err != nil {
		log.Println(err)
		return ""
	}

	matches := r.FindStringSubmatch(duration)

	years := parseInt64(matches[1])
	months := parseInt64(matches[2])
	days := parseInt64(matches[3])
	hours := parseInt64(matches[4])
	minutes := parseInt64(matches[5])
	seconds := parseInt64(matches[6])

	hour := int64(time.Hour)
	minute := int64(time.Minute)
	second := int64(time.Second)

	return time.Duration(years*24*365*hour +
		months*30*24*hour + days*24*hour +
		hours*hour + minutes*minute + seconds*second).String()
}

func parseInt64(value string) int64 {
	if len(value) == 0 {
		return 0
	}

	parsed, err := strconv.Atoi(value[:len(value)-1])
	if err != nil {
		return 0
	}

	return int64(parsed)
}

func IsValidYoutubeUrl(url string) bool {
	re := regexp.MustCompile(yTubeUrlRegex)
	if re.MatchString(url) {
		return true
	}
	return false
}

func ExtractYoutubePlaylistId(url string) string {
	re := regexp.MustCompile(yTubePlaylistUrlRegex)
	matches := re.FindStringSubmatch(url)
	if matches == nil {
		return ""
	}
	return matches[1]
}
