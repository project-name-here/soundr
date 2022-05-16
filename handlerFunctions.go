package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"time"

	"github.com/faiface/beep"
	"github.com/faiface/beep/flac"
	"github.com/faiface/beep/mp3"
	"github.com/faiface/beep/speaker"
	"github.com/faiface/beep/vorbis"
	"github.com/faiface/beep/wav"
	"github.com/h2non/filetype"
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
		buf, _ := ioutil.ReadFile("sounds/" + string(file))

		kind, _ := filetype.Match(buf)

		fmt.Println("File type: " + kind.MIME.Subtype)
		var streamer beep.StreamSeekCloser
		var format beep.Format
		if kind.MIME.Subtype == "mpeg" {
			streamer, format, _ = mp3.Decode(f)
		} else if kind.MIME.Subtype == "x-wav" {
			streamer, format, _ = wav.Decode(f)
		} else if kind.MIME.Subtype == "x-flac" {
			streamer, format, _ = flac.Decode(f)
		} else if kind.MIME.Subtype == "ogg" {
			streamer, format, _ = vorbis.Decode(f)
		} else {
			fmt.Println("!!!!! Unsupported file type for " + file)
			return false
		}

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
