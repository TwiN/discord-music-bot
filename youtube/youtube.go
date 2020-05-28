package youtube

import (
	"context"
	"errors"
	"fmt"
	"github.com/TwinProduction/discord-music-bot/core"
	downloader "github.com/kkdai/youtube"
	"google.golang.org/api/option"
	youtube "google.golang.org/api/youtube/v3"
	"io"
	"net/http"
	"os"
	"path/filepath"
)

var (
	ErrNoResultFound   = errors.New("no results found")
	ErrStreamListEmpty = errors.New("stream list is empty")
)

type Service struct {
	apiKey string

	client *http.Client
}

type SearchResult struct {
	Title   string
	VideoId string
}

func NewService(apiKey string) *Service {
	return &Service{
		apiKey: apiKey,
		client: &http.Client{Transport: &http.Transport{}},
	}
}

func (svc *Service) Search(query string) (*SearchResult, error) {
	api, err := youtube.NewService(context.Background(), option.WithAPIKey(svc.apiKey))
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

func (svc *Service) Download(result *SearchResult) (*core.Media, error) {
	yt := downloader.NewYoutube(false)
	if err := yt.DecodeURL(result.GetVideoUrl()); err != nil {
		return nil, err
	}
	if len(yt.StreamList) == 0 {
		return nil, ErrStreamListEmpty
	}
	response, err := svc.client.Get(yt.StreamList[0]["url"])
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()
	destinationFileName := fmt.Sprintf("%s.mp4", result.VideoId)
	if response.StatusCode == 200 {
		err = os.MkdirAll(filepath.Dir(destinationFileName), 0755)
		if err != nil {
			return nil, err
		}
		destinationFile, err := os.Create(destinationFileName)
		if err != nil {
			return nil, err
		}
		defer destinationFile.Close()
		_, err = io.Copy(destinationFile, response.Body)
	} else {
		return nil, errors.New("video url returned non-200 status code")
	}
	return core.NewMedia(result.Title, destinationFileName), nil
}

func (result *SearchResult) GetVideoUrl() string {
	return fmt.Sprintf("https://www.youtube.com/watch?v=%s", result.VideoId)
}
