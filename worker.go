package main

import "github.com/bwmarrin/discordgo"

func worker(bot *discordgo.Session, guildId, channelId string) {
	//voice, err := bot.ChannelVoiceJoin(guildId, channelId, false, true)
	//if err != nil {
	//	// XXX: return the error
	//	return
	//} else {
	//	defer voice.Disconnect()
	//	for media := range queues[guildId] {
	//		voice.OpusRecv <-
	//	}
	//}
	//close(queues[guildId])
}
