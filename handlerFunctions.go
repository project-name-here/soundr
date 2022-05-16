package main

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/faiface/beep"
	"github.com/faiface/beep/mp3"
	"github.com/faiface/beep/speaker"
)

func BufferSound(file string) bool {
	_, ok := streamMap[file]
	if !ok {
		fmt.Println("Not in memory, loading")
		f, err := os.Open("./sounds/" + file)
		if err != nil {
			log.Fatal(err)
		}

		fmt.Println("Opened file")
		streamer, format, _ := mp3.Decode(f)
		speaker.Init(format.SampleRate, format.SampleRate.N(time.Second/10))

		fmt.Println("Decoded file")
		buffer := beep.NewBuffer(format)
		buffer.Append(streamer)
		streamer.Close()
		fmt.Println("Bufferd file")

		// Save to streamMap
		streamMap[file] = streamBuf{
			Streamer: streamer,
			Format:   format,
			Buffer:   buffer,
		}
		return (true)
	} else {
		return (false)
	}
}

func PlaySound(file string, index int) int {
	playbacks[index] = playback{
		File:     file,
		IsLoaded: false,
		Streamer: nil,
		Control:  nil,
	}

	fmt.Println("Playing sound: " + file)
	var buffer *beep.Buffer
	BufferSound(file)
	buffer = streamMap[file].Buffer
	streamer := streamMap[file].Streamer

	fmt.Println("Trying to play sound")
	shot := buffer.Streamer(0, buffer.Len())

	done := make(chan bool)
	ctrl := &beep.Ctrl{Streamer: beep.Seq(shot, beep.Callback(func() {
		done <- true
	})), Paused: false}

	playbacks[index] = playback{
		File:     file,
		IsLoaded: true,
		Streamer: streamer,
		Control:  ctrl,
	}
	speaker.Play(ctrl)
	<-done
	fmt.Println("Finished playing sound: " + file)
	delete(playbacks, index)
	return 1
}
