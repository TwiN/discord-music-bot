package main

import (
	"log"
	"time"

	"github.com/bwmarrin/discordgo"
)

func StartJanitor(bot *discordgo.Session) {
	for {
		CleanUpVoiceConnections(bot)
		time.Sleep(30 * time.Second)
	}
}

func CleanUpVoiceConnections(bot *discordgo.Session) {
	for _, vc := range bot.VoiceConnections {
		guildsMutex.Lock()
		guild := guilds[vc.GuildID]
		guildsMutex.Unlock()
		if !guild.IsStreaming() {
			log.Printf("[janitor] Disconnecting VC in Guild %s because its media queue isn't even initialized", vc.GuildID)
			vc.Disconnect()
		}
	}
}
