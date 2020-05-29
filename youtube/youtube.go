package youtube

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/TwinProduction/discord-music-bot/core"
	"log"
	"net/http"
	"os/exec"
	"strings"
)

type Service struct {
	client *http.Client
}

func NewService() *Service {
	return &Service{
		client: &http.Client{Transport: &http.Transport{}},
	}
}

func (svc *Service) SearchAndDownload(query string) (*core.Media, error) {
	youtubeDownloader, err := exec.LookPath("youtube-dl")
	if err != nil {
		return nil, errors.New("youtube-dl not found in path")
	} else {
		args := []string{
			// TODO: sanitize input... lol
			fmt.Sprintf("ytsearch10:%s", strings.ReplaceAll(query, "\"", "")),
			"--extract-audio",
			"--audio-format", "opus",
			"--no-playlist",
			"--match-filter", "duration < 300",
			"--max-downloads", "1",
			"--output", "%(id)s.opus",
			"--print-json",
			"--ignore-errors", // Ignores unavailable videos
		}
		log.Printf("youtube-dl %s", strings.Join(args, " "))
		cmd := exec.Command(youtubeDownloader, args...)
		if data, err := cmd.Output(); err != nil && err.Error() != "exit status 101" {
			return nil, fmt.Errorf("failed to search and download audio: %s\n%s", err.Error(), string(data))
		} else {
			videoMetadata := VideoMetadata{}
			err = json.Unmarshal(data, &videoMetadata)
			if err != nil {
				return nil, fmt.Errorf("failed to unmarshal video metadata: %s", err.Error())
			}
			return core.NewMedia(
				videoMetadata.Title,
				fmt.Sprintf("%s.opus", videoMetadata.ID),
				videoMetadata.Uploader,
				fmt.Sprintf("https://www.youtube.com/watch?v=%s", videoMetadata.ID),
				videoMetadata.Thumbnail,
				videoMetadata.Duration,
			), nil
		}
	}
}

type VideoMetadata struct {
	ChannelID     string      `json:"channel_id"`
	ExtractorKey  string      `json:"extractor_key"`
	NEntries      int         `json:"n_entries"`
	DisplayID     string      `json:"display_id"`
	Filename      string      `json:"_filename"`
	PlayerURL     string      `json:"player_url"`
	FormatNote    string      `json:"format_note"`
	UploaderID    string      `json:"uploader_id"`
	Fps           interface{} `json:"fps"`
	DislikeCount  int         `json:"dislike_count"`
	Extractor     string      `json:"extractor"`
	AverageRating float64     `json:"average_rating"`
	Title         string      `json:"title"`
	Series        interface{} `json:"series"`
	LikeCount     int         `json:"like_count"`
	Track         interface{} `json:"track"`
	Formats       []struct {
		HTTPHeaders struct {
			AcceptCharset  string `json:"Accept-Charset"`
			UserAgent      string `json:"User-Agent"`
			AcceptEncoding string `json:"Accept-Encoding"`
			Accept         string `json:"Accept"`
			AcceptLanguage string `json:"Accept-Language"`
		} `json:"http_headers"`
		FormatNote        string  `json:"format_note"`
		Abr               int     `json:"abr,omitempty"`
		Vcodec            string  `json:"vcodec"`
		Tbr               float64 `json:"tbr"`
		URL               string  `json:"url"`
		Acodec            string  `json:"acodec"`
		Protocol          string  `json:"protocol"`
		FormatID          string  `json:"format_id"`
		Format            string  `json:"format"`
		DownloaderOptions struct {
			HTTPChunkSize int `json:"http_chunk_size"`
		} `json:"downloader_options,omitempty"`
		Width     interface{} `json:"width"`
		Height    interface{} `json:"height"`
		Filesize  int         `json:"filesize"`
		PlayerURL string      `json:"player_url"`
		Fps       interface{} `json:"fps"`
		Asr       int         `json:"asr"`
		Ext       string      `json:"ext"`
		Container string      `json:"container,omitempty"`
	} `json:"formats"`
	Fulltitle         string      `json:"fulltitle"`
	WebpageURL        string      `json:"webpage_url"`
	UploaderURL       string      `json:"uploader_url"`
	Categories        []string    `json:"categories"`
	AltTitle          interface{} `json:"alt_title"`
	StartTime         interface{} `json:"start_time"`
	PlaylistTitle     interface{} `json:"playlist_title"`
	Duration          int         `json:"duration"`
	Abr               int         `json:"abr"`
	ChannelURL        string      `json:"channel_url"`
	AgeLimit          int         `json:"age_limit"`
	DownloaderOptions struct {
		HTTPChunkSize int `json:"http_chunk_size"`
	} `json:"downloader_options"`
	ID                 string      `json:"id"`
	Height             interface{} `json:"height"`
	Format             string      `json:"format"`
	Protocol           string      `json:"protocol"`
	ViewCount          int         `json:"view_count"`
	RequestedSubtitles interface{} `json:"requested_subtitles"`
	PlaylistUploader   interface{} `json:"playlist_uploader"`
	Asr                int         `json:"asr"`
	Annotations        interface{} `json:"annotations"`
	Vcodec             string      `json:"vcodec"`
	Tags               []string    `json:"tags"`
	EndTime            interface{} `json:"end_time"`
	ReleaseYear        interface{} `json:"release_year"`
	WebpageURLBasename string      `json:"webpage_url_basename"`
	URL                string      `json:"url"`
	HTTPHeaders        struct {
		AcceptCharset  string `json:"Accept-Charset"`
		UserAgent      string `json:"User-Agent"`
		AcceptEncoding string `json:"Accept-Encoding"`
		Accept         string `json:"Accept"`
		AcceptLanguage string `json:"Accept-Language"`
	} `json:"http_headers"`
	Thumbnail         string      `json:"thumbnail"`
	SeasonNumber      interface{} `json:"season_number"`
	Uploader          string      `json:"uploader"`
	Width             interface{} `json:"width"`
	Chapters          interface{} `json:"chapters"`
	UploadDate        string      `json:"upload_date"`
	Artist            interface{} `json:"artist"`
	Tbr               float64     `json:"tbr"`
	Acodec            string      `json:"acodec"`
	AutomaticCaptions struct {
	} `json:"automatic_captions"`
	Subtitles struct {
	} `json:"subtitles"`
	ReleaseDate   interface{} `json:"release_date"`
	EpisodeNumber interface{} `json:"episode_number"`
	Thumbnails    []struct {
		URL string `json:"url"`
		ID  string `json:"id"`
	} `json:"thumbnails"`
	Ext                string      `json:"ext"`
	Description        string      `json:"description"`
	IsLive             interface{} `json:"is_live"`
	FormatID           string      `json:"format_id"`
	PlaylistID         string      `json:"playlist_id"`
	License            interface{} `json:"license"`
	PlaylistUploaderID interface{} `json:"playlist_uploader_id"`
	Album              interface{} `json:"album"`
	Filesize           int         `json:"filesize"`
	Playlist           string      `json:"playlist"`
	PlaylistIndex      int         `json:"playlist_index"`
	Creator            interface{} `json:"creator"`
}
