# ffmpeg_recorder
simple Golang example of recording a video using ffmpeg with cgo.

## Installation
sudo apt-get install libavcodec-dev libavformat-dev libavutil-dev libswscale-dev

## If ffmpeg is not in standard place set environment variables
```
export CGO_CFLAGS="-I/path/to/ffmpeg/include"
export CGO_LDFLAGS="-L/path/to/ffmpeg/libs -lavcodec -lavformat -lavutil"
```
 
