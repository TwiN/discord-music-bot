package main

import (
	"github.com/TwinProduction/discord-music-bot/core"
	"github.com/TwinProduction/discord-music-bot/dca"
	"github.com/bwmarrin/discordgo"
	"io"
	"log"
	"os"
	"time"
)

func worker(bot *discordgo.Session, guildId, channelId string) error {
	guildName := GetGuildNameById(bot, guildId)
	// See https://github.com/Malchemy/DankMemes/blob/master/sound.go#L26
	voice, err := bot.ChannelVoiceJoin(guildId, channelId, false, false)
	if err != nil {
		return err
	}
	defer voice.Disconnect()
	voice.Speaking(true)
	defer voice.Speaking(false)
	for media := range queues[guildId] {
		if !actionQueues[guildId].Stopped {
			play(voice, media, guildName, actionQueues[guildId])
		}
		_ = os.Remove(media.FilePath)
		if len(queues[guildId]) == 0 {
			break
		}
		log.Printf("[%s] There are currently %d medias in the queue", guildName, len(queues[guildId]))
	}
	time.Sleep(500 * time.Millisecond)

	log.Printf("[%s] Closing channel", guildName)
	close(queues[guildId])
	queues[guildId] = nil
	actionQueues[guildId] = nil
	return nil
}

func play(voice *discordgo.VoiceConnection, media *core.Media, guildName string, actions *core.Actions) {
	options := dca.StdEncodeOptions
	options.BufferedFrames = 100
	options.FrameDuration = 20
	options.CompressionLevel = 5
	options.Bitrate = 96

	encodeSession, err := dca.EncodeFile(media.FilePath, options)
	if err != nil {
		log.Printf("[%s] Failed to create encoding session for \"%s\": %s", guildName, media.FilePath, err.Error())
		return
	}
	defer encodeSession.Cleanup()

	time.Sleep(500 * time.Millisecond)

	done := make(chan error)
	dca.NewStream(encodeSession, voice, done)

	select {
	case err := <-done:
		if err != nil && err != io.EOF {
			log.Printf("[%s] Error occurred during stream for \"%s\": %s", guildName, media.FilePath, err.Error())
			return
		}
	case <-actions.SkipChan:
		log.Printf("[%s] Skipping \"%s\"", guildName, media.FilePath)
		_ = encodeSession.Stop()
	case <-actions.StopChan:
		log.Printf("[%s] Stopping", guildName)
		_ = encodeSession.Stop()
	}
}
