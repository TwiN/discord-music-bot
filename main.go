package main

import (
	"errors"
	"fmt"
	"github.com/TwinProduction/discord-music-bot/config"
	"github.com/TwinProduction/discord-music-bot/core"
	"github.com/TwinProduction/discord-music-bot/youtube"
	"github.com/bwmarrin/discordgo"
	"log"
	"os"
	"os/signal"
	"strings"
	"sync"
	"syscall"
	"time"
)

const (
	CommandPrefix = "!"
	MaxQueueSize  = 10
)

var (
	ErrUserNotInVoiceChannel = errors.New("couldn't find voice channel with user in it")

	actionQueues = make(map[string]*core.Actions)
	queues       = make(map[string]chan *core.Media)
	queuesMutex  = sync.RWMutex{}

	// guildNames is a mapping between guild id and guild name
	guildNames = make(map[string]string)

	youtubeService *youtube.Service
)

func main() {
	config.Load()
	youtubeService = youtube.NewService()
	bot, err := Connect(config.Get().DiscordToken)
	if err != nil {
		panic(err)
	}
	defer bot.Close()
	defer func() {
		if actionQueues != nil {
			for guildId, actions := range actionQueues {
				log.Printf("Shutting down, stopping queues for guild %s", guildNames[guildId])
				actions.Stop()
			}
		}
		time.Sleep(2 * time.Second)
	}()

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
		if command == "skip" {
			actionQueues[message.GuildID].Skip()
			return
		}
		if command == "stop" {
			actionQueues[message.GuildID].Stop()
			return
		}
	}
}

func HandleYoutubeCommand(bot *discordgo.Session, message *discordgo.MessageCreate, query string) {
	if len(queues[message.GuildID]) >= MaxQueueSize {
		_, _ = bot.ChannelMessageSend(message.ChannelID, "The queue is full!")
		return
	}
	guildName := GetGuildNameById(bot, message.GuildID)

	// Find the voice channel the user is in
	voiceChannelId, err := GetVoiceChannelWhereMessageAuthorIs(bot, message)
	if err != nil {
		log.Printf("[%s] Failed to find voice channel where message author is located: %s", guildName, err.Error())
		_, _ = bot.ChannelMessageSend(message.ChannelID, err.Error())
		return
	}
	log.Printf("[%s] Found user %s in voice channel %s", guildName, message.Author.Username, voiceChannelId)

	log.Printf("[%s] Searching for \"%s\"", guildName, query)
	botMessage, _ := bot.ChannelMessageSend(message.ChannelID, fmt.Sprintf(":mag: Searching for `%s`...", query))
	media, err := youtubeService.SearchAndDownload(query)
	if err != nil {
		log.Printf("[%s] Unable to find video for query \"%s\": %s", guildName, query, err.Error())
		_, _ = bot.ChannelMessageSend(message.ChannelID, fmt.Sprintf("Unable to find video for query `%s`: %s", query, err.Error()))
		return
	}
	log.Printf("[%s] Successfully searched for and extracted audio from video with title \"%s\" to \"%s\"", guildName, media.Title, media.FilePath)
	botMessage, _ = bot.ChannelMessageEdit(botMessage.ChannelID, botMessage.ID, fmt.Sprintf(":white_check_mark: Found matching video titled `%s`!", media.Title))
	go func(bot *discordgo.Session, message *discordgo.Message) {
		time.Sleep(time.Second)
		_ = bot.ChannelMessageDelete(botMessage.ChannelID, botMessage.ID)
	}(bot, botMessage)

	// Add song to guild queue
	createNewWorker := false
	queuesMutex.Lock()
	defer queuesMutex.Unlock()
	if queues[message.GuildID] == nil {
		queues[message.GuildID] = make(chan *core.Media, MaxQueueSize)
		actionQueues[message.GuildID] = core.NewActions()
		// If the channel was nil, it means that there was no worker
		createNewWorker = true
	}
	queues[message.GuildID] <- media
	log.Printf("[%s] Added media with title \"%s\" to queue at position %d", guildName, media.Title, len(queues[message.GuildID]))
	_, _ = bot.ChannelMessageSendEmbed(message.ChannelID, &discordgo.MessageEmbed{
		URL:         media.URL,
		Title:       media.Title,
		Description: fmt.Sprintf("Position in queue: %d", len(queues[message.GuildID])),
		Thumbnail: &discordgo.MessageEmbedThumbnail{
			URL: media.Thumbnail,
		},
	})

	if createNewWorker {
		log.Printf("[%s] Starting worker", guildName)
		go func() {
			err = worker(bot, message.GuildID, voiceChannelId)
			if err != nil {
				log.Printf("[%s] Failed to start worker: %s", guildName, err.Error())
				_, _ = bot.ChannelMessageSend(message.ChannelID, fmt.Sprintf("Unable to start voice worker: %s", err.Error()))
				_ = os.Remove(media.FilePath)
				return
			}
		}()
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
