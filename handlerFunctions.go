package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"time"

	"github.com/gopxl/beep"
	"github.com/gopxl/beep/flac"
	"github.com/gopxl/beep/mp3"
	"github.com/gopxl/beep/speaker"
	"github.com/gopxl/beep/vorbis"
	"github.com/gopxl/beep/wav"
	"github.com/h2non/filetype"
)

var firstLoad = true

func BufferSound(file string) bool {
	_, ok := streamMap[file]
	if !ok {
		fmt.Println("----Not in memory, loading----")
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
			streamer, format, err = mp3.Decode(f)
		} else if kind.MIME.Subtype == "x-wav" {
			streamer, format, err = wav.Decode(f)
		} else if kind.MIME.Subtype == "x-flac" {
			streamer, format, err = flac.Decode(f)
		} else if kind.MIME.Subtype == "ogg" {
			streamer, format, err = vorbis.Decode(f)
		} else {
			fmt.Println("!!!!! Unsupported file type for " + file)
			return false
		}
		if firstLoad {
			speaker.Init(format.SampleRate, format.SampleRate.N(time.Second/10))
			firstLoad = false
		}
		if err != nil {
			fmt.Println("Error while decoding file: " + file)
			fmt.Println("Error: " + fmt.Sprintf("%v", err))
			return false
		}

		fmt.Println("Decoded file")
		fmt.Println("Current file: " + file)
		fmt.Println("Current streamer: " + fmt.Sprintf("%v", streamer))
		fmt.Println("Error: " + fmt.Sprintf("%v", err))
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
		fmt.Println("Looping sound: " + file)
	}
	shot := buffer.Streamer(0, buffer.Len())

	done := make(chan bool)
	ctrl := &beep.Ctrl{Streamer: beep.Loop(amountOfLoops, shot), Paused: false}

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
