package youtube

import (
	"context"
	"log"

	"google.golang.org/api/option"
	"google.golang.org/api/youtube/v3"
)

const (
	DefaultPlaylistItemCount int64  = 20
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
	VideoPath  string
	CoverUrl   string
	CoverPath  string
}

func NewYoutubeAPI(developerKey string, ctx context.Context) *YoutubeAPI {
	return &YoutubeAPI{
		DeveloperKey: developerKey,
		Context:      ctx,
	}
}

func (y *YoutubeAPI) GetSearchResults(query string) []SearchResult {

	service, err := youtube.NewService(y.Context, option.WithAPIKey(y.DeveloperKey))
	if err != nil {
		log.Fatalf("Error while creating new Youtube Client : %v", err)
	}

	var results []SearchResult
	call := service.Search.List([]string{"id", "snippet"}).Q(query)
	resp, err := call.Do()
	if err != nil {
		log.Println(err)
	}

	for _, item := range resp.Items {
		searchResult := SearchResult{
			VideoID:    item.Id.VideoId,
			VideoTitle: item.Snippet.Title,
			VideoPath:  youtubeUrlPrefix + item.Id.VideoId,
		}

		results = append(results, searchResult)
	}

	log.Println(results)

	return results
}
