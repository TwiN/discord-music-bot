package main

import (
	"github.com/TwinProduction/discord-music-bot/core"
	"github.com/bwmarrin/discordgo"
	"github.com/jonas747/dca"
	"io"
	"log"
	"time"
)

func worker(bot *discordgo.Session, guildId, channelId string) error {
	time.Sleep(250 * time.Millisecond)
	guildName := GetGuildNameById(bot, guildId)
	// See https://github.com/Malchemy/DankMemes/blob/master/sound.go#L26
	voice, err := bot.ChannelVoiceJoin(guildId, channelId, false, false)
	if err != nil {
		return err
	}
	defer voice.Disconnect()
	voice.Speaking(true)
	defer voice.Speaking(false)
	log.Printf("[%s] there are currently %d medias in the queue", guildName, len(queues[guildId]))
	for media := range queues[guildId] {
		play(voice, media, guildName)
		if len(queues[guildId]) == 0 {
			break
		}
		log.Printf("[%s] there are currently %d medias in the queue", guildName, len(queues[guildId]))
	}
	time.Sleep(500 * time.Millisecond)

	close(queues[guildId])
	queues[guildId] = nil
	return nil
}

func play(voice *discordgo.VoiceConnection, media *core.Media, guildName string) {
	//if filepath.Ext(media.FilePath) != ".dca" {
	//	err := ffmpeg.ConvertMp3ToDca(media)
	//	if err != nil {
	//		log.Printf("[%s] Skipping, because failed to convert \"%s\" to dca: %s", guildName, media.FilePath, err.Error())
	//		return
	//	}
	//}

	options := dca.StdEncodeOptions
	options.RawOutput = true
	options.Bitrate = 96

	encodeSession, err := dca.EncodeFile(media.FilePath, options)
	if err != nil {
		log.Printf("[%s] Failed to create encoding session for \"%s\": %s", guildName, media.FilePath, err.Error())
		return
	}

	done := make(chan error)
	dca.NewStream(encodeSession, voice, done)

	err = <-done
	if err != nil && err != io.EOF {
		log.Printf("[%s] Error occurred during stream for \"%s\": %s", guildName, media.FilePath, err.Error())
		return
	}

	// Clean up in case something happened and ffmpeg is still running
	encodeSession.Truncate()

	/*

		log.Printf("[%s] Opening DCA audio file \"%s\"", guildName, media.FilePath)
		file, err := os.Open(media.FilePath)
		if err != nil {
			log.Printf("[%s] Skipping, because couldn't open DCA file \"%s\": %s", guildName, media.FilePath, err.Error())
			return
		}

		decoder := dca.NewDecoder(file)
		fmt.Println(decoder.FrameDuration().String())
		for {
			frame, err := decoder.OpusFrame()
			if err != nil {
				if err != io.EOF {
					log.Printf("[%s] %v", guildName, err)
				} else {
					log.Printf("[%s] Reached end of DCA file: %v", guildName, err)
				}
				break
			}
			select {
			case voice.OpusSend <- frame:
			case <-time.After(time.Second):
				continue
			}
		}
	*/
}
