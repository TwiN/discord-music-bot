package main

import (
	"fmt"
	"github.com/bwmarrin/discordgo"
)

func GetVoiceChannelWhereMessageAuthorIs(bot *discordgo.Session, message *discordgo.MessageCreate) (string, error) {
	guild, err := bot.Guild(message.GuildID)
	if err != nil {
		return "", fmt.Errorf("unable to find voice channel user is in: %s", err.Error())
	}
	for _, voiceState := range guild.VoiceStates {
		if voiceState.UserID == message.Author.ID {
			return voiceState.ChannelID, nil
		}
	}
	return "", ErrUserNotInVoiceChannel
}

func Connect(discordToken string) (*discordgo.Session, error) {
	discord, err := discordgo.New(fmt.Sprintf("Bot %s", discordToken))
	if err != nil {
		return nil, err
	}
	err = discord.Open()
	return discord, err
}

func GetGuildNameById(bot *discordgo.Session, guildId string) string {
	guildName, ok := guildNames[guildId]
	if !ok {
		guild, err := bot.Guild(guildId)
		if err != nil {
			// Failed to get the guild? Whatever, we'll just use the guild id
			guildNames[guildId] = guildId
			return guildId
		}
		guildNames[guildId] = guild.Name
		return guild.Name
	}
	return guildName
}
