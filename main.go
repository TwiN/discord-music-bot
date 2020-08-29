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

var (
	ErrUserNotInVoiceChannel = errors.New("couldn't find voice channel with user in it")

	guilds      = make(map[string]*core.ActiveGuild)
	guildsMutex = sync.RWMutex{}

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
		for _, guild := range guilds {
			log.Printf("[%s] Shutting down: Closing queue", guild.Name)
			if guild.UserActions != nil {
				guild.UserActions.Stop()
			}
		}
		// There shouldn't be any VC still open, but just in case
		for _, vc := range bot.VoiceConnections {
			vc.Disconnect()
		}
		time.Sleep(250 * time.Millisecond)
	}()

	bot.AddHandler(HandleMessage)
	log.Println("Connected successfully")
	go StartJanitor(bot)

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
		guildsMutex.Lock()
		activeGuild := guilds[message.GuildID]
		guildsMutex.Unlock()
		switch command {
		case "youtube", "yt", "play":
			HandleYoutubeCommand(bot, activeGuild, message, query)
		case "skip":
			if activeGuild != nil && activeGuild.UserActions != nil {
				activeGuild.UserActions.Skip()
			}
		case "stop":
			if activeGuild != nil && activeGuild.UserActions != nil {
				activeGuild.UserActions.Stop()
			} else {
				// If queue is nil and the user still wrote !stop, it's possible that there's a VC still active
				bot.Lock()
				if bot.VoiceConnections[message.GuildID] != nil {
					log.Printf("[%s] Force disconnecting VC (!stop was called and queue was already nil)", GetGuildNameById(bot, message.GuildID))
					bot.VoiceConnections[message.GuildID].Disconnect()
				}
				bot.Unlock()
			}
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

func HandleYoutubeCommand(bot *discordgo.Session, activeGuild *core.ActiveGuild, message *discordgo.MessageCreate, query string) {
	if activeGuild != nil {
		if activeGuild.IsMediaQueueFull() {
			_, _ = bot.ChannelMessageSend(message.ChannelID, "The queue is full!")
			return
		}
	} else {
		activeGuild = core.NewActiveGuild(GetGuildNameById(bot, message.GuildID))
		guildsMutex.Lock()
		guilds[message.GuildID] = activeGuild
		guildsMutex.Unlock()
	}

	// Find the voice channel the user is in
	voiceChannelId, err := GetVoiceChannelWhereMessageAuthorIs(bot, message)
	if err != nil {
		log.Printf("[%s] Failed to find voice channel where message author is located: %s", activeGuild.Name, err.Error())
		_ = bot.MessageReactionAdd(message.ChannelID, message.ID, "❌")
		_, _ = bot.ChannelMessageSend(message.ChannelID, err.Error())
		return
	} else {
		log.Printf("[%s] Found user %s in voice channel %s", activeGuild.Name, message.Author.Username, voiceChannelId)
		_ = bot.MessageReactionAdd(message.ChannelID, message.ID, "✅")
	}

	log.Printf("[%s] Searching for \"%s\"", activeGuild.Name, query)
	botMessage, _ := bot.ChannelMessageSend(message.ChannelID, fmt.Sprintf(":mag: Searching for `%s`...", query))
	media, err := youtubeService.SearchAndDownload(query)
	if err != nil {
		log.Printf("[%s] Unable to find video for query \"%s\": %s", activeGuild.Name, query, err.Error())
		_, _ = bot.ChannelMessageSend(message.ChannelID, fmt.Sprintf("Unable to find video for query `%s`: %s", query, err.Error()))
		return
	}
	log.Printf("[%s] Successfully searched for and extracted audio from video with title \"%s\" to \"%s\"", activeGuild.Name, media.Title, media.FilePath)
	botMessage, _ = bot.ChannelMessageEdit(botMessage.ChannelID, botMessage.ID, fmt.Sprintf(":white_check_mark: Found matching video titled `%s`!", media.Title))
	go func(bot *discordgo.Session, message *discordgo.Message) {
		time.Sleep(5 * time.Second)
		_ = bot.ChannelMessageDelete(botMessage.ChannelID, botMessage.ID)
	}(bot, botMessage)

	// Add song to guild queue
	createNewWorker := false
	if !activeGuild.IsStreaming() {
		log.Printf("[%s] Preparing for streaming", activeGuild.Name)
		activeGuild.PrepareForStreaming(config.Get().MaximumQueueSize)
		// If the channel was nil, it means that there was no worker
		createNewWorker = true
	}
	activeGuild.EnqueueMedia(media)

	log.Printf("[%s] Added media with title \"%s\" to queue at position %d", activeGuild.Name, media.Title, activeGuild.MediaQueueSize())
	_, _ = bot.ChannelMessageSendEmbed(message.ChannelID, &discordgo.MessageEmbed{
		URL:         media.URL,
		Title:       media.Title,
		Description: fmt.Sprintf("Position in queue: %d", activeGuild.MediaQueueSize()),
		Thumbnail: &discordgo.MessageEmbedThumbnail{
			URL: media.Thumbnail,
		},
	})

	if createNewWorker {
		log.Printf("[%s] Starting worker", activeGuild.Name)
		go func() {
			err = worker(bot, activeGuild, message.GuildID, voiceChannelId)
			if err != nil {
				log.Printf("[%s] Failed to start worker: %s", activeGuild.Name, err.Error())
				_, _ = bot.ChannelMessageSend(message.ChannelID, fmt.Sprintf("❌ Unable to start voice worker: %s", err.Error()))
				_ = os.Remove(media.FilePath)
				return
			}
		}()
	}
}
