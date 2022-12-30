package youtube

import (
	"context"
	"errors"
	"fmt"
	"github.com/puristt/discord-bot-go/util"
	"google.golang.org/api/option"
	"google.golang.org/api/youtube/v3"
	"log"
)

const (
	defaultPlaylistItemCount int64  = 15
	maxResults               int64  = 20
	youtubeUrlPrefix         string = "https://www.youtube.com/watch?v="
)

type YoutubeAPI struct {
	DeveloperKey string
	Context      context.Context
}

type SearchResult struct {
	VideoID       string
	VideoTitle    string
	Duration      string
	VideoUrl      string
	VideoImageUrl string
}

func NewYoutubeAPI(developerKey string, ctx context.Context) *YoutubeAPI {
	return &YoutubeAPI{
		DeveloperKey: developerKey,
		Context:      ctx,
	}
}

func (y *YoutubeAPI) GetSearchResults(query string) ([]SearchResult, error) {
	results, err := y.handleSearchResults(query, maxResults/2)
	return results, err
}

func (y *YoutubeAPI) GetVideoInfo(query string) (SearchResult, error) {
	results, err := y.handleSearchResults(query, 1)
	if err != nil {
		log.Println(err)
		return SearchResult{}, err
	}
	res := results[0]
	log.Println(res)

	return res, nil
}

func (y *YoutubeAPI) handleSearchResults(query string, maxResult int64) ([]SearchResult, error) {
	service, err := youtube.NewService(y.Context, option.WithAPIKey(y.DeveloperKey)) // TODO : NewService method will be called once in app lifetime
	if err != nil {
		log.Fatalf("Error while creating new Youtube Client : %v", err)
	}

	var results []SearchResult
	call := service.Search.List([]string{"id", "snippet"}).Q(query).MaxResults(maxResult)
	resp, err := call.Do()
	if err != nil {
		return results, err
	}

	if len(resp.Items) == 0 {
		return nil, errors.New("no results found")
	}

	var videoIds []string
	for _, item := range resp.Items {
		videoIds = append(videoIds, item.Id.VideoId)
	}

	durations, err := y.getDurationsByIds(videoIds)
	if err != nil {
		return results, err
	}

	for _, item := range resp.Items {
		searchResult := SearchResult{
			VideoID:       item.Id.VideoId,
			VideoTitle:    item.Snippet.Title,
			VideoUrl:      youtubeUrlPrefix + item.Id.VideoId,
			Duration:      durations[item.Id.VideoId],
			VideoImageUrl: item.Snippet.Thumbnails.Default.Url,
		}

		results = append(results, searchResult)
	}

	return results, nil
}

func (y *YoutubeAPI) GetPlaylistItems(playlistId string, pageToken string) (result []SearchResult, nextPageToken string, err error) {
	service, err := youtube.NewService(y.Context, option.WithAPIKey(y.DeveloperKey))
	if err != nil {
		log.Fatalf("Error while creating new Youtube Client : %v", err)
	}

	var results []SearchResult
	call := service.PlaylistItems.List([]string{"snippet"}).PlaylistId(playlistId).MaxResults(defaultPlaylistItemCount)
	if pageToken != "" {
		call.PageToken(pageToken)
	}
	playList, err := call.Do()
	if err != nil {
		return results, "", err
	}

	if len(playList.Items) == 0 {
		return nil, "", errors.New("No results found")
	}

	var videoIds []string
	for _, item := range playList.Items {
		videoIds = append(videoIds, item.Snippet.ResourceId.VideoId)
	}

	durations, err := y.getDurationsByIds(videoIds)
	if err != nil {
		return results, "", err
	}

	for _, playListItem := range playList.Items {
		searchResult := SearchResult{
			VideoID:       playListItem.Snippet.ResourceId.VideoId,
			VideoTitle:    playListItem.Snippet.Title,
			VideoUrl:      youtubeUrlPrefix + playListItem.Snippet.ResourceId.VideoId,
			Duration:      durations[playListItem.Snippet.ResourceId.VideoId],
			VideoImageUrl: playListItem.Snippet.Thumbnails.Default.Url,
		}

		results = append(results, searchResult)
	}

	return results, playList.NextPageToken, nil
}

func (y *YoutubeAPI) getDurationsByIds(ids []string) (map[string]string, error) {
	durations := make(map[string]string, len(ids))

	service, err := youtube.NewService(y.Context, option.WithAPIKey(y.DeveloperKey))
	if err != nil {
		log.Fatalf("Error while creating new Youtube Client : %v", err)
	}

	call := service.Videos.List([]string{"id", "contentDetails"}).Id(ids...)
	resp, err := call.Do()
	if err != nil {
		return durations, fmt.Errorf("error while getting video duration by id : %v", err)
	}

	for _, item := range resp.Items {
		durations[item.Id] = util.ParseISO8601(item.ContentDetails.Duration)
	}

	return durations, nil
}
