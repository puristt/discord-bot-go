package util

import (
	"path"
	"strings"
)

const (
	YoutubeSongPath = "youtube/song"
)

// FormatVideoTitle formats given string appropriate for file name.
func FormatVideoTitle(videoTitle string) string {
	newTitle := strings.TrimSpace(videoTitle)

	strReplacer := strings.NewReplacer("/", "_", "-", "_", ",", "_", " ", "", "'", "")
	newTitle = strReplacer.Replace(newTitle)

	return newTitle
}

// GetVideoPath returns formatted version of the given video title and video id
// as full file video path.
func GetVideoPath(videoTitle string, videoID string) string {
	formattedTitlePath := FormatVideoTitle(videoTitle) + "_" + videoID + ".m4a"

	formattedTitleFullPath := path.Join(YoutubeSongPath, formattedTitlePath)
	return formattedTitleFullPath
}
