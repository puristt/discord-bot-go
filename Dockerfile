# Build image
FROM golang:1.19-alpine as builder
WORKDIR /src
# Necessary to go build
RUN apk update && apk add pkgconfig gcc libc-dev \
                        opusfile \
                        opusfile-dev
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN go build -o /bin/app .


FROM alpine
COPY --from=builder /bin/app /bin/app
COPY --from=builder /src/config.json .

# Necessary packages to play audio and opus codec
RUN apk update && apk add pkgconfig gcc libc-dev \
                        opusfile \
                        opusfile-dev \
                        ffmpeg \
                        yt-dlp

ENTRYPOINT ["/bin/app"]