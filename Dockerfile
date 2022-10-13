# Build the go application into a binary
FROM golang:alpine as builder
WORKDIR /app
ADD . ./
RUN CGO_ENABLED=0 GOOS=linux go build -buildvcs=false -a -installsuffix cgo -o bin/discord-music-bot .

FROM alpine:3.16
ENV DISCORD_BOT_TOKEN=""
ENV BOT_ADMINS=""
ENV COMMAND_PREFIX=""
ENV MAXIMUM_AUDIO_DURATION_IN_SECONDS=""
ENV APP_HOME=/app
WORKDIR ${APP_HOME}
RUN apk --update add --no-cache ca-certificates ffmpeg opus python3
COPY --from=builder /app/bin/discord-music-bot ./bin/discord-music-bot
RUN wget --no-check-certificate https://github.com/yt-dlp/yt-dlp/releases/latest/download/yt-dlp -O /usr/local/bin/yt-dlp
RUN chmod a+rx /usr/local/bin/yt-dlp
RUN ln -s /usr/bin/python3 /usr/bin/python
RUN yt-dlp --version
ENTRYPOINT ["/app/bin/discord-music-bot"]