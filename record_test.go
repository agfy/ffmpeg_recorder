package main

import (
	"github.com/go-vgo/robotgo"
	"log"
	"testing"
)

func TestEncoder(t *testing.T) {
	e, err := NewEncoder("test.mpeg", 640, 514, 25)
	if err != nil {
		log.Panicf("Unable to start encoder: %q", err)
	}
	defer e.Close()

	img, err := robotgo.Read("test.png")
	if err != nil {
		t.Error(err)
	}

	bmp := robotgo.ImgToBitmap(img)

	println(bmp.Width)

	for i := 0; i < 250; i++ {
		err = e.Encode(img)
		if err != nil {
			t.Error(err)
		}
	}
}
