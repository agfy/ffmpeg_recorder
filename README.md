# ffmpeg_recorder
simple Golang example of recording a video using ffmpeg with cgo.

## Installation
sudo apt-get install libavcodec-dev libavformat-dev libavutil-dev libswscale-dev

## Environment variables
set env variables CGO_CFLAGS and CGO_LDFLAGS if ffmpeg is not in standard place:

export CGO_CFLAGS="-I/path/to/ffmpeg/include"
export CGO_LDFLAGS="-I/path/to/ffmpeg/libs -lavcodec -lavformat -lavutil"