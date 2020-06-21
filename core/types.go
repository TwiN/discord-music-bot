package core

import (
	"errors"
	"fmt"
	"time"
)

var (
	ErrAlreadyReadyForStreaming = errors.New("already ready for streaming")
)

// ActiveGuild is a guild that is currently being streamed to.
type ActiveGuild struct {
	Name        string
	UserActions *UserActions
	MediaChan   chan *Media
}

func NewActiveGuild(name string) *ActiveGuild {
	guild := &ActiveGuild{
		Name:        name,
		UserActions: nil,
		MediaChan:   nil,
	}
	return guild
}

func (g *ActiveGuild) PrepareForStreaming(maxQueueSize int) {
	g.MediaChan = make(chan *Media, maxQueueSize)
	g.UserActions = NewActions()
}

func (g *ActiveGuild) IsStreaming() bool {
	return g.MediaChan != nil
}

func (g *ActiveGuild) EnqueueMedia(media *Media) {
	g.MediaChan <- media
}

func (g *ActiveGuild) MediaQueueSize() int {
	return len(g.MediaChan)
}

func (g *ActiveGuild) IsMediaQueueFull() bool {
	return g.MediaChan != nil && len(g.MediaChan) == cap(g.MediaChan)
}

func (g *ActiveGuild) StopStreaming() {
	close(g.MediaChan)
	g.MediaChan = nil
	g.UserActions = nil
}

type Media struct {
	Title     string
	FilePath  string
	Uploader  string
	URL       string
	Thumbnail string
	Duration  time.Duration
}

func NewMedia(title, filePath, uploader, url, thumbnail string, durationInSeconds int) *Media {
	duration, _ := time.ParseDuration(fmt.Sprintf("%ds", durationInSeconds))
	return &Media{
		Title:     title,
		FilePath:  filePath,
		Uploader:  uploader,
		URL:       url,
		Thumbnail: thumbnail,
		Duration:  duration,
	}
}

type UserActions struct {
	SkipChan chan bool
	StopChan chan bool

	Stopped bool
}

func NewActions() *UserActions {
	return &UserActions{
		SkipChan: make(chan bool, 1),
		StopChan: make(chan bool, 1),
	}
}

func (a *UserActions) Stop() {
	a.Stopped = true
	a.StopChan <- true
}

func (a *UserActions) Skip() {
	a.SkipChan <- true
}
