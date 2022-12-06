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
	defaultPlaylistItemCount int64  = 20
	maxResults               int64  = 20
	youtubeUrlPrefix         string = "https://www.youtube.com/watch?v="
)

type YoutubeAPI struct {
	DeveloperKey string
	Context      context.Context
}

type SearchResult struct {
	VideoID    string
	VideoTitle string
	Duration   string
	VideoUrl   string
}

func NewYoutubeAPI(developerKey string, ctx context.Context) *YoutubeAPI {
	return &YoutubeAPI{
		DeveloperKey: developerKey,
		Context:      ctx,
	}
}

func (y *YoutubeAPI) GetSearchResults(query string) []SearchResult {
	results, _ := y.handleSearchResults(query, maxResults/2)
	return results
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

// TODO: Is Valid YouTube url control will be added
func (y *YoutubeAPI) handleSearchResults(query string, maxResult int64) ([]SearchResult, error) {
	service, err := youtube.NewService(y.Context, option.WithAPIKey(y.DeveloperKey))
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
		return nil, errors.New("No results found")
	}

	var videoIds []string
	for _, item := range resp.Items {
		videoIds = append(videoIds, item.Id.VideoId)
	}

	durations, err := y.GetDurationsByIds(videoIds)
	if err != nil {
		return results, err
	}

	for i, item := range resp.Items {
		searchResult := SearchResult{
			VideoID:    item.Id.VideoId,
			VideoTitle: item.Snippet.Title,
			VideoUrl:   youtubeUrlPrefix + item.Id.VideoId,
			Duration:   durations[i],
			// TODO: Video cover image can be obtained
		}

		results = append(results, searchResult)
	}

	return results, nil
}

func (y *YoutubeAPI) GetDurationsByIds(ids []string) ([]string, error) {
	var durations []string
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
		durations = append(durations, util.ParseISO8601(item.ContentDetails.Duration))
	}

	return durations, nil
}
