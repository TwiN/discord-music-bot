# Build the go application into a binary
FROM golang:alpine as builder
WORKDIR /app
ADD . ./
RUN CGO_ENABLED=0 GOOS=linux go build -buildvcs=false -a -installsuffix cgo -o bin/discord-music-bot .

FROM alpine:3.7
ENV DISCORD_BOT_TOKEN=""
ENV BOT_ADMINS=""
ENV COMMAND_PREFIX=""
ENV MAXIMUM_AUDIO_DURATION_IN_SECONDS=""
ENV APP_HOME=/app
WORKDIR ${APP_HOME}
RUN apk --update add --no-cache ca-certificates ffmpeg opus python
COPY --from=builder /app/bin/discord-music-bot ./bin/discord-music-bot
RUN wget --no-check-certificate https://yt-dl.org/downloads/latest/youtube-dl -O /usr/bin/youtube-dl
RUN chmod +x /usr/bin/youtube-dl
RUN youtube-dl -U
RUN youtube-dl --version
ENTRYPOINT ["/app/bin/discord-music-bot"]