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
	MaxQueueSize = 10
)

var (
	ErrUserNotInVoiceChannel = errors.New("couldn't find voice channel with user in it")

	actionQueues      = make(map[string]*core.Actions)
	actionQueuesMutex = sync.RWMutex{}

	mediaQueues      = make(map[string]chan *core.Media)
	mediaQueuesMutex = sync.RWMutex{}

	// guildNames is a mapping between guild id and guild name
	guildNames = make(map[string]string)

	youtubeService *youtube.Service
)

func main() {
	config.Load()
	youtubeService = youtube.NewService(config.Get().MaximumAudioDurationInSeconds)
	bot, err := Connect(config.Get().DiscordToken)
	if err != nil {
		panic(err)
	}
	defer bot.Close()
	_ = bot.UpdateListeningStatus(fmt.Sprintf("%shelp", config.Get().CommandPrefix))
	defer func() {
		if actionQueues != nil {
			for guildId, actions := range actionQueues {
				log.Printf("Shutting down, stopping queues for guild %s", guildNames[guildId])
				if actions != nil {
					actions.Stop()
				}
			}
		}
		time.Sleep(1 * time.Second)
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
	commandPrefix := config.Get().CommandPrefix
	if message.Author.Bot || message.Author.ID == bot.State.User.ID {
		return
	}
	if strings.HasPrefix(message.Content, commandPrefix) {
		command := strings.Replace(strings.Split(message.Content, " ")[0], commandPrefix, "", 1)
		query := strings.TrimSpace(strings.Replace(message.Content, fmt.Sprintf("%s%s", commandPrefix, command), "", 1))
		command = strings.ToLower(command)
		switch command {
		case "youtube", "yt":
			HandleYoutubeCommand(bot, message, query)
		case "skip":
			actionQueuesMutex.Lock()
			if actionQueues[message.GuildID] != nil {
				actionQueues[message.GuildID].Skip()
			}
			actionQueuesMutex.Unlock()
		case "stop":
			actionQueuesMutex.Lock()
			if actionQueues[message.GuildID] != nil {
				actionQueues[message.GuildID].Stop()
			}
			actionQueuesMutex.Unlock()
		case "info":
			_, _ = bot.ChannelMessageSend(message.ChannelID, "See https://github.com/TwinProduction/discord-music-bot")
		case "health":
			latency := bot.HeartbeatLatency()
			_, _ = bot.ChannelMessageSend(message.ChannelID, fmt.Sprintf("Heartbeat latency: %s", latency))
		case "help":
			_, _ = bot.ChannelMessageSend(message.ChannelID, fmt.Sprintf(`
__**Commands**__
**%syoutube or %syt**: Searches for a song on youtube and adds it to the playlist
**%sskip**: Skip the current audio clip
**%sstop**: Flush the bot's playlist and disconnect it from the voice channel'
**%sinfo**: Provides general information about the bot
**%shealth**: Provides information about the health of the bot

Bugs to report? Create an issue at https://github.com/TwinProduction/discord-music-bot
`, commandPrefix, commandPrefix, commandPrefix, commandPrefix, commandPrefix, commandPrefix))
		}
	}
}

func HandleYoutubeCommand(bot *discordgo.Session, message *discordgo.MessageCreate, query string) {
	mediaQueuesMutex.Lock()
	queueSize := len(mediaQueues[message.GuildID])
	mediaQueuesMutex.Unlock()
	if queueSize >= MaxQueueSize {
		_, _ = bot.ChannelMessageSend(message.ChannelID, "The queue is full!")
		return
	}
	guildName := GetGuildNameById(bot, message.GuildID)

	// Find the voice channel the user is in
	voiceChannelId, err := GetVoiceChannelWhereMessageAuthorIs(bot, message)
	if err != nil {
		log.Printf("[%s] Failed to find voice channel where message author is located: %s", guildName, err.Error())
		_ = bot.MessageReactionAdd(message.ChannelID, message.ID, "❌")
		_, _ = bot.ChannelMessageSend(message.ChannelID, err.Error())
		return
	} else {
		log.Printf("[%s] Found user %s in voice channel %s", guildName, message.Author.Username, voiceChannelId)
		_ = bot.MessageReactionAdd(message.ChannelID, message.ID, "✅")
	}

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
		time.Sleep(3 * time.Second)
		_ = bot.ChannelMessageDelete(botMessage.ChannelID, botMessage.ID)
	}(bot, botMessage)

	// Add song to guild queue
	createNewWorker := false
	mediaQueuesMutex.Lock()
	if mediaQueues[message.GuildID] == nil {
		mediaQueues[message.GuildID] = make(chan *core.Media, MaxQueueSize)
		actionQueuesMutex.Lock()
		actionQueues[message.GuildID] = core.NewActions()
		actionQueuesMutex.Unlock()
		// If the channel was nil, it means that there was no worker
		createNewWorker = true
	}
	mediaQueues[message.GuildID] <- media
	mediaQueuesMutex.Unlock()
	log.Printf("[%s] Added media with title \"%s\" to queue at position %d", guildName, media.Title, len(mediaQueues[message.GuildID]))
	_, _ = bot.ChannelMessageSendEmbed(message.ChannelID, &discordgo.MessageEmbed{
		URL:         media.URL,
		Title:       media.Title,
		Description: fmt.Sprintf("Position in queue: %d", len(mediaQueues[message.GuildID])),
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
