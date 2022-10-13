package youtube

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/TwiN/discord-music-bot/core"
)

type Service struct {
	maxDurationInSeconds int
	fileDirectory        string
}

func NewService(maxDurationInSeconds int) *Service {
	_ = os.Mkdir("data", os.ModePerm)
	return &Service{
		maxDurationInSeconds: maxDurationInSeconds,
		fileDirectory:        "data",
	}
}

func (svc *Service) SearchAndDownload(query string) (*core.Media, error) {
	timeout := make(chan bool, 1)
	result := make(chan searchAndDownloadResult, 1)
	go func() {
		time.Sleep(60 * time.Second)
		timeout <- true
	}()
	go func() {
		result <- svc.doSearchAndDownload(query)
	}()
	select {
	case <-timeout:
		return nil, errors.New("timed out")
	case result := <-result:
		return result.Media, result.Error
	}
}

func (svc *Service) doSearchAndDownload(query string) searchAndDownloadResult {
	start := time.Now()
	youtubeDownloader, err := exec.LookPath("yt-dlp")
	if err != nil {
		return searchAndDownloadResult{Error: errors.New("yt-dlp not found in path")}
	} else {
		args := []string{
			fmt.Sprintf("ytsearch10:%s", strings.ReplaceAll(query, "\"", "")),
			"--extract-audio",
			"--audio-format", "opus",
			"--no-playlist",
			"--match-filter", fmt.Sprintf("duration < %d & !is_live", svc.maxDurationInSeconds),
			"--max-downloads", "1",
			"--output", fmt.Sprintf("%s/%d-%%(id)s.opus", svc.fileDirectory, start.Unix()),
			"--quiet",
			"--print-json",
			"--ignore-errors", // Ignores unavailable videos
			"--no-color",
			"--no-check-formats",
		}
		log.Printf("yt-dlp %s", strings.Join(args, " "))
		cmd := exec.Command(youtubeDownloader, args...)
		if data, err := cmd.Output(); err != nil && err.Error() != "exit status 101" {
			return searchAndDownloadResult{Error: fmt.Errorf("failed to search and download audio: %s\n%s", err.Error(), string(data))}
		} else {
			videoMetadata := videoMetadata{}
			err = json.Unmarshal(data, &videoMetadata)
			if err != nil {
				fmt.Println(string(data))
				return searchAndDownloadResult{Error: fmt.Errorf("failed to unmarshal video metadata: %w", err)}
			}
			return searchAndDownloadResult{
				Media: core.NewMedia(
					videoMetadata.Title,
					videoMetadata.Filename,
					videoMetadata.Uploader,
					fmt.Sprintf("https://www.youtube.com/watch?v=%s", videoMetadata.ID),
					videoMetadata.Thumbnail,
					int(videoMetadata.Duration),
				),
			}
		}
	}
}

type searchAndDownloadResult struct {
	Media *core.Media
	Error error
}

type videoMetadata struct {
	ID                   string      `json:"id"`
	Title                string      `json:"title"`
	Thumbnail            string      `json:"thumbnail"`
	Description          string      `json:"description"`
	Uploader             string      `json:"uploader"`
	UploaderID           string      `json:"uploader_id"`
	UploaderURL          string      `json:"uploader_url"`
	ChannelID            string      `json:"channel_id"`
	ChannelURL           string      `json:"channel_url"`
	Duration             int         `json:"duration"`
	ViewCount            int         `json:"view_count"`
	AverageRating        interface{} `json:"average_rating"`
	AgeLimit             int         `json:"age_limit"`
	WebpageURL           string      `json:"webpage_url"`
	Categories           []string    `json:"categories"`
	Tags                 []string    `json:"tags"`
	PlayableInEmbed      bool        `json:"playable_in_embed"`
	LiveStatus           interface{} `json:"live_status"`
	ReleaseTimestamp     interface{} `json:"release_timestamp"`
	CommentCount         interface{} `json:"comment_count"`
	LikeCount            int         `json:"like_count"`
	Channel              string      `json:"channel"`
	ChannelFollowerCount int         `json:"channel_follower_count"`
	UploadDate           string      `json:"upload_date"`
	Availability         string      `json:"availability"`
	OriginalURL          string      `json:"original_url"`
	WebpageURLBasename   string      `json:"webpage_url_basename"`
	WebpageURLDomain     string      `json:"webpage_url_domain"`
	Extractor            string      `json:"extractor"`
	ExtractorKey         string      `json:"extractor_key"`
	PlaylistCount        int         `json:"playlist_count"`
	Playlist             string      `json:"playlist"`
	PlaylistID           string      `json:"playlist_id"`
	PlaylistTitle        string      `json:"playlist_title"`
	PlaylistUploader     interface{} `json:"playlist_uploader"`
	PlaylistUploaderID   interface{} `json:"playlist_uploader_id"`
	NEntries             int         `json:"n_entries"`
	PlaylistIndex        int         `json:"playlist_index"`
	LastPlaylistIndex    int         `json:"__last_playlist_index"`
	PlaylistAutonumber   int         `json:"playlist_autonumber"`
	DisplayID            string      `json:"display_id"`
	Fulltitle            string      `json:"fulltitle"`
	DurationString       string      `json:"duration_string"`
	RequestedSubtitles   interface{} `json:"requested_subtitles"`
	Asr                  int         `json:"asr"`
	Filesize             int         `json:"filesize"`
	FormatID             string      `json:"format_id"`
	FormatNote           string      `json:"format_note"`
	SourcePreference     int         `json:"source_preference"`
	Fps                  interface{} `json:"fps"`
	AudioChannels        int         `json:"audio_channels"`
	Height               interface{} `json:"height"`
	Quality              int         `json:"quality"`
	HasDrm               bool        `json:"has_drm"`
	Tbr                  float64     `json:"tbr"`
	URL                  string      `json:"url"`
	Width                interface{} `json:"width"`
	Language             string      `json:"language"`
	LanguagePreference   int         `json:"language_preference"`
	Preference           interface{} `json:"preference"`
	Ext                  string      `json:"ext"`
	Vcodec               string      `json:"vcodec"`
	Acodec               string      `json:"acodec"`
	DynamicRange         interface{} `json:"dynamic_range"`
	Abr                  float64     `json:"abr"`
	Filename             string      `json:"filename"`
}
