# discord-music-bot
[![Docker pulls](https://img.shields.io/docker/pulls/twinproduction/discord-music-bot)](https://cloud.docker.com/repository/docker/twinproduction/discord-music-bot)

This is a minimal music bot for Discord that support streaming to multiple servers concurrently.  

It uses [yt-dlp](https://github.com/yt-dlp/yt-dlp), a fork of youtube-dl, to search and download the video as well as `ffmpeg` to convert and stream the audio.


## Usage
| Environment variable              | Description                                                                                   | Required | Default |
|:----------------------------------|:----------------------------------------------------------------------------------------------|:---------|:--------|
| DISCORD_BOT_TOKEN                 | Discord bot token                                                                             | yes      | `""`    |
| COMMAND_PREFIX                    | Character prepending all bot commands. Must be exactly 1 character, or it will default to `!` | no       | `!`     |
| MAXIMUM_AUDIO_DURATION_IN_SECONDS | Maximum duration of audio clips in second                                                     | no       | `480`   |
| MAXIMUM_QUEUE_SIZE                | Maximum number of medias that can be queued up per server/guild                               | no       | `10`    |
| BOT_ADMINS                        | Comma-separated list of user ids                                                              | no       | `""`    |


## Getting started
### Discord
1. Create an application
2. Add a bot in your application
3. Save the bot's token and set it as the `DISCORD_BOT_TOKEN` environment variable
4. Go to `https://discordapp.com/oauth2/authorize?client_id=<YOUR_BOT_CLIENT_ID>&scope=bot&permissions=3222592`
5. Add music bot to server


### Bot commands
Assuming `COMMAND_PREFIX` is not defined or is set to `!`.

| Command                    | Description                                      | Example            |
|:---------------------------|:-------------------------------------------------|:-------------------|
| `!youtube`, `!yt`, `!play` | Add a song to the queue                          | `!yt what is love` |
| `!skip`                    | Skip the current song                            |
| `!stop`                    | Skip all songs in the queue                      |
| `!help`                    | Display all commands                             |
| `!health`                  | Provides information about the health of the bot |
| `!info`                    | Provides general information about the bot       |
| `!restart`                 | Restarts the bot. Must be admin.                 |


## Prerequisites
If you want to run it locally, you'll need the following applications:
- [yt-dlp](https://github.com/yt-dlp/yt-dlp)
- ffmpeg


## Docker
### Pulling from Docker Hub
```
docker pull twinproduction/discord-music-bot
```

### Building image locally
Building the Docker image is done as following:
```
docker build . -t twinproduction/discord-music-bot
```
You can then run the container with the following command:
```
docker run -e DISCORD_BOT_TOKEN=secret --name discord-music-bot twinproduction/discord-music-bot
```


## FAQ
### How do I add my bot to a new server?
See step 4 of the [Discord](#discord) section.
