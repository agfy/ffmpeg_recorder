package main

import (
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/agfy/doom_environment"
)

func main() {
	//ffmpeg.ExampleEncoderUsage()
	env, err := doom_environment.Create(1)
	if err != nil {
		println(err.Error())
	}
	time.Sleep(10 * time.Second)
	err = env.Start()
	if err != nil {
		println(err.Error())
	}
	sigc := make(chan os.Signal, 1)
	signal.Notify(sigc,
		syscall.SIGINT,
		syscall.SIGTERM,
	)
	_ = <-sigc
}
