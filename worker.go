package main

import (
	"fmt"
	"github.com/TwinProduction/discord-music-bot/core"
	"github.com/bwmarrin/discordgo"
	"github.com/jonas747/dca"
	"io"
	"log"
	"os"
	"time"
)

func worker(bot *discordgo.Session, guildId, channelId string) error {
	time.Sleep(2 * time.Second)
	guildName := GetGuildNameById(bot, guildId)
	// See https://github.com/Malchemy/DankMemes/blob/master/sound.go#L26
	voice, err := bot.ChannelVoiceJoin(guildId, channelId, false, false)
	if err != nil {
		panic(err)
		return err
	}
	defer voice.Disconnect()
	voice.Speaking(true)
	defer voice.Speaking(false)
	log.Printf("[%s] there are currently %d medias in the queue", guildName, len(queues[guildId]))
	for media := range queues[guildId] {
		convertMp3ToDca(media)

		log.Printf("[%s] Opening DCA audio file \"%s\"", guildName, media.FilePath)
		file, err := os.Open(fmt.Sprintf("%s.dca", media.FilePath))
		if err != nil {
			log.Printf("[%s] Couldn't open audio file \"%s\", skipping \"%s\"", guildName, media.FilePath, media.Title)
			continue
		}

		// inputReader is an io.Reader, like a file for example
		decoder := dca.NewDecoder(file)
		for {
			frame, err := decoder.OpusFrame()
			if err != nil {
				if err != io.EOF {
					log.Printf("[%s] %v", guildName, err)
				}
				log.Printf("[%s] Reached end of DCA file: %v", guildName, err)
				break
			}
			select {
			case voice.OpusSend <- frame:
				log.Println("sending a frame")
			case <-time.After(time.Second):
				log.Println("rip")
				continue
			}
		}
		//if len(queues[guildId]) == 0 {
		//	break
		//}
		log.Printf("[%s] there are currently %d medias in the queue", guildName, len(queues[guildId]))
	}

	close(queues[guildId])
	queues[guildId] = nil
	return nil
}

func convertMp3ToDca(media *core.Media) {
	options := dca.StdEncodeOptions
	options.RawOutput = true
	encodeSession, err := dca.EncodeFile(media.FilePath, options)
	if err != nil {
		panic(err)
	}
	defer encodeSession.Cleanup()
	output, err := os.Create(fmt.Sprintf("%s.dca", media.FilePath))
	if err != nil {
		panic(err)
	}
	defer output.Close()
	_, err = io.Copy(output, encodeSession)
	if err != nil {
		panic(err)
	}
}
