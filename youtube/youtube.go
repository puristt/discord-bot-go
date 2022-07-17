package youtube

import (
	"context"
	"errors"
	"log"

	"google.golang.org/api/option"
	"google.golang.org/api/youtube/v3"
)

const (
	defaultPlaylistItemCount int64  = 20
	maxResults               int64  = 25
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
	results, _ := y.handleSearchResults(query, maxResults)
	return results
}

func (y *YoutubeAPI) PlayVideo(query string) (SearchResult, error) {
	res, err := y.handleSearchResults(query, 1)
	if err != nil {
		log.Println(err)
		return SearchResult{}, err
	}

	log.Println(res)
	return res[0], nil
}

func (y *YoutubeAPI) handleSearchResults(query string, maxResult int64) ([]SearchResult, error) {
	service, err := youtube.NewService(y.Context, option.WithAPIKey(y.DeveloperKey))
	if err != nil {
		log.Fatalf("Error while creating new Youtube Client : %v", err)
	}

	var results []SearchResult
	call := service.Search.List([]string{"id", "snippet"}).Q(query).MaxResults(maxResult)
	resp, err := call.Do()
	if err != nil {
		log.Println(err)
		return results, err
	}

	if len(resp.Items) == 0 {
		return nil, errors.New("No results found")
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

	log.Println("-----------------------------------")
	return results, nil
}
