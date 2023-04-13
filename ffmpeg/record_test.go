package ffmpeg

import (
	"github.com/go-vgo/robotgo"
	"log"
	"testing"
)

func TestEncoder(t *testing.T) {
	e, err := NewEncoder("test.mpeg")
	if err != nil {
		log.Panicf("Unable to start encoder: %q", err)
	}

	pixels, err := robotgo.OpenImg("test_img.png")
	if err != nil {
		t.Error(err)
	}

	println(pixels[0])

	for i := 0; i < 25; i++ {
		e.CreateFrame(i)
		e.Encode()
	}
	e.WriteLastFrame()
	e.Encode()

	e.Close()
}
