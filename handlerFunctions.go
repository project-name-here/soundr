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

var firstLoad = true

func BufferSound(file string) bool {
	_, ok := streamMap[file]
	if !ok {
		fmt.Println("Not in memory, loading")
		f, err := os.Open("./sounds/" + file)
		if err != nil {
			log.Fatal(err)
		}

		fmt.Println("Opened file")
		buf, err := ioutil.ReadFile("sounds/" + string(file))
		if err != nil {
			log.Fatal("Fatal error while opening: " + err.Error())
			return false
		}

		kind, err := filetype.Match(buf)

		if err != nil {
			log.Fatal("Fatal error while detecting file type: " + err.Error())
			return false
		}

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
		if firstLoad {
			speaker.Init(format.SampleRate, format.SampleRate.N(time.Second/10))
			firstLoad = false
		}

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

func PlaySound(file string, index int, loop bool) int {
	playbacks[index] = playback{
		File:     file,
		IsLoaded: false,
		Streamer: nil,
		Control:  nil,
		Loop:     loop,
		Format:   streamMap[file].Format,
	}

	fmt.Println("Playing sound: " + file)
	var buffer *beep.Buffer
	BufferSound(file)
	buffer = streamMap[file].Buffer
	// streamer := streamMap[file].Streamer

	fmt.Println("Trying to play sound")
	amountOfLoops := 1
	if loop {
		amountOfLoops = -1
	}
	shot := buffer.Streamer(amountOfLoops, buffer.Len())

	done := make(chan bool)
	ctrl := &beep.Ctrl{Streamer: shot, Paused: false}

	playbacks[index] = playback{
		File:     file,
		IsLoaded: true,
		Streamer: shot,
		Control:  ctrl,
		Loop:     loop,
		Format:   streamMap[file].Format,
		Done:     done,
	}
	speaker.Play(ctrl)
	<-done
	fmt.Println("Finished playing sound: " + file)
	delete(playbacks, index)

	return 1
}
