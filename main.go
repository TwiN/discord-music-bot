package main

import (
	"errors"
	"fmt"
	"github.com/TwinProduction/discord-music-bot/config"
	"github.com/TwinProduction/discord-music-bot/core"
	"github.com/TwinProduction/discord-music-bot/ffmpeg"
	"github.com/TwinProduction/discord-music-bot/youtube"
	"github.com/bwmarrin/discordgo"
	"log"
	"os"
	"os/signal"
	"strings"
	"sync"
	"syscall"
)

const (
	CommandPrefix = "!"
)

var (
	ErrUserNotInVoiceChannel = errors.New("couldn't find voice channel with user in it")

	queues      = make(map[string][]*core.Media)
	queuesMutex = sync.RWMutex{}

	// guildNames is a mapping between guild id and guild name
	guildNames = make(map[string]string)

	youtubeService *youtube.Service
)

func main() {
	config.Load()
	youtubeService = youtube.NewService(config.Get().YoutubeApiKey)
	bot, err := Connect(config.Get().DiscordToken)
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
	if message.Author.Bot || message.Author.ID == bot.State.User.ID {
		return
	}
	if strings.HasPrefix(message.Content, CommandPrefix) {
		command := strings.Replace(strings.Split(message.Content, " ")[0], CommandPrefix, "", 1)
		query := strings.TrimSpace(strings.Replace(message.Content, fmt.Sprintf("%s%s", CommandPrefix, command), "", 1))
		command = strings.ToLower(command)
		if command == "youtube" || command == "yt" {
			HandleYoutubeCommand(bot, message, query)
			return
		}
	}
}

func HandleYoutubeCommand(bot *discordgo.Session, message *discordgo.MessageCreate, query string) {
	if len(queues[message.GuildID]) > 10 {
		_, _ = bot.ChannelMessageSend(message.ChannelID, "The queue is full!")
		return
	}
	guildName := GetGuildNameById(message.GuildID, bot)

	// Find the voice channel the user is in
	voiceChannelId, err := GetVoiceChannelWhereMessageAuthorIs(bot, message)
	if err != nil {
		log.Printf("[%s] Failed to find voice channel where message author is located: %s", guildName, err.Error())
		_, _ = bot.ChannelMessageSend(message.ChannelID, err.Error())
		return
	}
	log.Printf("[%s] Found user %s in voice channel %s", guildName, message.Author.Username, voiceChannelId)

	// Search for video
	log.Printf("[%s] Searching for \"%s\"", guildName, query)
	result, err := youtubeService.Search(query)
	if err != nil {
		log.Printf("[%s] Failed to search for video: %s", guildName, err.Error())
		_, _ = bot.ChannelMessageSend(message.ChannelID, fmt.Sprintf("Unable to search for video: %s", err.Error()))
		return
	}
	log.Printf("[%s] Found video titled \"%s\" from query \"%s\"", guildName, result.Title, query)

	// Download the video
	log.Printf("[%s] Downloading video with title \"%s\"", guildName, result.Title)
	media, err := youtubeService.Download(result)
	if err != nil {
		log.Printf("[%s] Failed to download video: %s", guildName, err.Error())
		_, _ = bot.ChannelMessageSend(message.ChannelID, fmt.Sprintf("Unable to search for video based on query \"%s\"", query))
		return
	}
	log.Printf("[%s] Downloaded video with title \"%s\" at \"%s\"", guildName, media.Title, media.FilePath)

	// Convert video to audio
	log.Printf("[%s] Extracting audio from video with title \"%s\"", guildName, result.Title)
	err = ffmpeg.ConvertVideoToAudio(media)
	if err != nil {
		log.Printf("[%s] Failed to convert video to audio: %s", guildName, err.Error())
		_, _ = bot.ChannelMessageSend(message.ChannelID, fmt.Sprintf("Unable to convert video to audio: %s", err.Error()))
		_ = os.Remove(media.FilePath)
		return
	}
	log.Printf("[%s] Extracted audio from video with title \"%s\" into \"%s\"", guildName, result.Title, media.FilePath)

	// Add song to guild queue
	queuesMutex.Lock()
	defer queuesMutex.Unlock()
	queues[message.GuildID] = append(queues[message.GuildID], media)
	log.Printf("[%s] Added media with title \"%s\" to queue at position %d", guildName, media.Title, len(queues[message.GuildID]))

	// TODO: Join channel (if not already in one)
	// if not already in one, then start goroutine that takes care of streaming the queue. (guild worker)?
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

func Connect(discordToken string) (*discordgo.Session, error) {
	discord, err := discordgo.New(fmt.Sprintf("Bot %s", discordToken))
	if err != nil {
		return nil, err
	}
	err = discord.Open()
	return discord, err
}

func GetGuildNameById(guildId string, bot *discordgo.Session) string {
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
