package main

import (
	"errors"
	"fmt"
	"github.com/bwmarrin/discordgo"
	"log"
	"os"
	"os/signal"
	"strings"
	"syscall"
)

const (
	CommandPrefix = "!"
)

var (
	ErrUserNotInVoiceChannel = errors.New("couldn't find voice channel with user in it")
)

func main() {
	bot, err := Connect()
	if err != nil {
		panic(err)
	}
	defer bot.Close()

	bot.AddHandler(HandleMessage)
	log.Println("Connected successfully")

	// Wait for the bot to be killed
	channel := make(chan os.Signal, 1)
	signal.Notify(channel, syscall.SIGINT, syscall.SIGTERM, os.Interrupt, os.Kill)
	<-channel

	log.Println("Terminating bot")
}

func HandleMessage(bot *discordgo.Session, message *discordgo.MessageCreate) {
	if message.Author.ID == bot.State.User.ID {
		return
	}
	if strings.HasPrefix(message.Content, CommandPrefix) {
		command := strings.Replace(strings.Split(message.Content, " ")[0], CommandPrefix, "", 1)
		query := strings.TrimSpace(strings.Replace(message.Content, fmt.Sprintf("%s%s", CommandPrefix, command), "", 1))
		command = strings.ToLower(command)
		if command == "youtube" || command == "yt" {
			// Find the voice channel the user is in
			voiceChannelId, err := GetVoiceChannelWhereMessageAuthorIs(bot, message)
			if err != nil {
				bot.ChannelMessageSend(message.ChannelID, err.Error())
				return
			}
			bot.ChannelMessageSend(message.ChannelID, fmt.Sprintf("user is in voice channel %s", voiceChannelId))
			// TODO: Search for query (can use the scraper I made on my old bot)
			log.Printf("Searching for \"%s\"", query)
			bot.ChannelMessageSend(message.ChannelID, "Search results: <...>")
			// TODO: Download audio (can use a library for this)

			// TODO: Add song to queue (queue must be per guild, i.e. map[guild id]Media)
			// XXX: Media should contain: query, length, title and youtube link

			// TODO: Join channel (if not already in one)
			// if not already in one, then start goroutine that takes care of streaming the queue.
		}

	}
}

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

func Connect() (*discordgo.Session, error) {
	token := os.Getenv("DISCORD_BOT_TOKEN")
	discord, err := discordgo.New(fmt.Sprintf("Bot %s", token))
	if err != nil {
		return nil, err
	}
	err = discord.Open()
	return discord, err
}
