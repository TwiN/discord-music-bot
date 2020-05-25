package youtube

import (
	"context"
	"errors"
	"fmt"
	"github.com/TwinProduction/discord-music-bot/core"
	downloader "github.com/kkdai/youtube"
	"google.golang.org/api/option"
	youtube "google.golang.org/api/youtube/v3"
)

var (
	ErrNoResultFound = errors.New("no results found")
)

type Service struct {
	apiKey string
}

type SearchResult struct {
	Title   string
	VideoId string
}

func NewService(apiKey string) *Service {
	return &Service{apiKey: apiKey}
}

func (yt *Service) Search(query string) (*SearchResult, error) {
	api, err := youtube.NewService(context.Background(), option.WithAPIKey(yt.apiKey))
	if err != nil {
		return nil, err
	}
	response, err := api.Search.List("snippet").Type("video").VideoDuration("short").MaxResults(1).Q(query).Do()
	if err != nil {
		return nil, err
	}
	if len(response.Items) == 1 {
		result := &SearchResult{
			Title:   response.Items[0].Snippet.Title,
			VideoId: response.Items[0].Id.VideoId,
		}
		return result, nil

	}
	return nil, ErrNoResultFound
}

func (result *SearchResult) GetVideoUrl() string {
	return fmt.Sprintf("https://www.youtube.com/watch?v=%s", result.VideoId)
}

func (result *SearchResult) Download() (*core.Media, error) {
	yt := downloader.NewYoutube(false)
	if err := yt.DecodeURL(result.GetVideoUrl()); err != nil {
		return nil, err
	}
	fileName := fmt.Sprintf("%s.mp4", result.VideoId)
	if err := yt.StartDownload(fileName); err != nil {
		return nil, err
	}
	return core.NewMedia(result.Title, fileName), nil
}
