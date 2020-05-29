# discord-music-bot

Minimal music bot for Discord.


## Getting started

### Discord

1. Create an application
2. Add a bot in your application
3. Save the bot's token and set it as the `DISCORD_BOT_TOKEN` environment variable
4. Go to `https://discordapp.com/oauth2/authorize?client_id=<YOUR_BOT_CLIENT_ID>&scope=bot&permissions=36785216`
5. Add music bot to server


### Usage

```
!yt Haddaway - what is love
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