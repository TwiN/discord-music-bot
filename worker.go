package main

import "github.com/bwmarrin/discordgo"

func worker(bot *discordgo.Session, guildId, channelId string) {
	// See https://github.com/Malchemy/DankMemes/blob/master/sound.go#L26
	//voice, err := bot.ChannelVoiceJoin(guildId, channelId, false, true)
	//if err != nil {
	//	// XXX: return the error
	//	return
	//} else {
	//	defer voice.Disconnect()
	//	for media := range queues[guildId] {
	//		voice.OpusSend <-
	//	}
	//}
	//close(queues[guildId])
}
