package youtube

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/TwinProduction/discord-music-bot/core"
	"google.golang.org/api/option"
	youtube "google.golang.org/api/youtube/v3"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
)

var (
	ErrNoResultFound    = errors.New("no results found")
	ErrNoVideoLinkFound = errors.New("no video link found")
	ErrStreamListEmpty  = errors.New("stream list is empty")
)

type SearchResult struct {
	Title   string
	VideoId string
}

func (result *SearchResult) GetVideoUrl() string {
	return fmt.Sprintf("https://www.youtube.com/watch?v=%s", result.VideoId)
}

type Service struct {
	apiKey string

	client *http.Client
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
	data, err := svc.getVideoInfo(result)
	if err != nil {
		return nil, err
	}
	videoFile := videoFile{}
	extractVideoFileUrl(data, &videoFile)
	if len(videoFile.URL) == 0 {
		return nil, ErrNoVideoLinkFound
	}
	response, err := svc.client.Get(videoFile.URL)
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

func (svc *Service) getVideoInfo(result *SearchResult) (string, error) {
	response, err := http.Get(fmt.Sprintf("https://www.youtube.com/get_video_info?video_id=%s&el=embedded&ps=default&eurl=&gl=US&hl=en", result.VideoId))
	if err != nil {
		return "", err
	}
	encodedData, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return "", err
	}
	data, err := url.QueryUnescape(string(encodedData))
	if err != nil {
		return "", err
	}
	data, err = url.PathUnescape(data)
	if err != nil {
		return "", err
	}
	return data, nil
}

// extractVideoFileUrl is literally garbage
// "If you know it's garbage, why did you make it?"
// Because Youtube's API keeps changing, and as long as they don't modify the field names, this will never break.
func extractVideoFileUrl(data string, videoFile *videoFile) {
	data = strings.ReplaceAll(data, "\\u0026", "&")
	start := strings.Index(data, "{")
	if len(data) < start+1 {
		return
	}
	end := strings.Index(data[start:], "}") + start
	videoFileJson := data[start : end+1]
	_ = json.Unmarshal([]byte(videoFileJson), videoFile)
	if len(videoFile.Quality) != 0 && len(videoFile.URL) != 0 {
		if videoFile.Quality == "medium" {
			return
		}
	}
	extractVideoFileUrl(data[start+1:], videoFile)
}

type videoFile struct {
	Itag             int    `json:"itag"`
	URL              string `json:"url"`
	MimeType         string `json:"mimeType"`
	Bitrate          int    `json:"bitrate"`
	Width            int    `json:"width"`
	Height           int    `json:"height"`
	LastModified     string `json:"lastModified"`
	ContentLength    string `json:"contentLength"`
	Quality          string `json:"quality"`
	Fps              int    `json:"fps"`
	QualityLabel     string `json:"qualityLabel"`
	ProjectionType   string `json:"projectionType"`
	AverageBitrate   int    `json:"averageBitrate"`
	AudioQuality     string `json:"audioQuality"`
	ApproxDurationMs string `json:"approxDurationMs"`
	AudioSampleRate  string `json:"audioSampleRate"`
	AudioChannels    int    `json:"audioChannels"`
}
