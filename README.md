# discord-music-bot

[![Docker pulls](https://img.shields.io/docker/pulls/twinproduction/discord-music-bot)](https://cloud.docker.com/repository/docker/twinproduction/discord-music-bot)

This is a minimal music bot for Discord that support streaming to multiple servers concurrently.  

It uses `youtube-dl` to search and download the video as well as `ffmpeg` to convert and stream the audio.


## Usage

| Environment variable | Description | Required | Default |
| --- | --- | --- | --- |
| DISCORD_BOT_TOKEN | Discord bot token | yes | `""` |


## Getting started

### Discord

1. Create an application
2. Add a bot in your application
3. Save the bot's token and set it as the `DISCORD_BOT_TOKEN` environment variable
4. Go to `https://discordapp.com/oauth2/authorize?client_id=<YOUR_BOT_CLIENT_ID>&scope=bot&permissions=36785216`
5. Add music bot to server


### Bot commands

#### Adding a song to the queue

```
!youtube remember the name
```


#### Skipping the current song

```
!skip
```


#### Skipping all songs in the queue
```
!stop
```


## Prerequisites

If you want to run it locally, you'll need the following applications:
- youtube-dl
- ffmpeg


## Docker

Building the Docker image is done as following:

```
docker build . -t discord-music-bot
```

You can then run the container with the following command:

```
docker run -e DISCORD_BOT_TOKEN=secret --name discord-music-bot discord-music-bot
```