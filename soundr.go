package main

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"html"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strconv"
	"sync"
	"time"

	"github.com/faiface/beep"
	"github.com/faiface/beep/mp3"
	"github.com/faiface/beep/speaker"
)

type playback struct {
	File     string
	IsLoaded bool
	Streamer beep.Streamer
	Control  *beep.Ctrl
}

type playbackWebReturn struct {
	File     string
	IsLoaded bool
	Id       int
}

type streamBuf struct {
	Streamer beep.Streamer
	Format   beep.Format
	Buffer   *beep.Buffer
}

var playbacks map[int]playback
var mapMutex = sync.Mutex{}

var streamMap map[string]streamBuf

func BufferSound(file string) bool {
	_, ok := streamMap[file]
	if !ok {
		fmt.Println("Not in memory, loading")
		f, err := os.Open("./sounds/" + file)
		if err != nil {
			log.Fatal(err)
		}

		fmt.Println("Opened file")
		streamer, format, err := mp3.Decode(f)
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

func main() {

	playbacks = make(map[int]playback)
	streamMap = make(map[string]streamBuf)

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "Hello, %q", html.EscapeString(r.URL.Query().Get("name")))
	})

	http.HandleFunc("/v1/play", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		var cnt = r.URL.Query().Get("file")
		bytArr, err := base64.StdEncoding.DecodeString(cnt)
		if err != nil {
			log.Fatal(err)
		}
		fmt.Println(string(bytArr[:]))
		t, err := os.Stat("./sounds/" + string(bytArr[:]))
		t = t
		if !errors.Is(err, os.ErrNotExist) {
			var currIndex = len(playbacks)
			fmt.Fprintf(w, "{\"status\":\"ok\", \"id\":%d}", currIndex)

			go PlaySound(string(bytArr[:]), currIndex)

		} else {
			fmt.Fprintf(w, "{\"status\":\"fail\", \"reason\":\"file not found\"}")
		}

	})

	http.HandleFunc("/v1/buffer", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		var cnt = r.URL.Query().Get("file")
		bytArr, err := base64.StdEncoding.DecodeString(cnt)
		if err != nil {
			log.Fatal(err)
		}
		t, err := os.Stat("./sounds/" + string(bytArr[:]))
		t = t
		if !errors.Is(err, os.ErrNotExist) {
			fmt.Fprintf(w, "{\"status\":\"ok\"}")
			go BufferSound(string(bytArr[:]))
		} else {
			fmt.Fprintf(w, "{\"status\":\"fail\", \"reason\":\"file not found\"}")
		}

	})

	http.HandleFunc("/v1/stopAll", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		playbacks = make(map[int]playback)
		fmt.Fprintf(w, "{\"status\":\"ok\"}")
		speaker.Clear()
		//fmt.Fprintf(w, "{\"status\":\"ok\", \"id\":%d}", currIndex)

	})

	http.HandleFunc("/v1/stop", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		var cnt, err = strconv.Atoi(r.URL.Query().Get("id"))
		if err != nil {
			fmt.Fprintf(w, "{\"status\":\"fail\", \"reason\":\"invalid id\"}")
		}

		value, ok := playbacks[cnt]
		if !ok {
			fmt.Fprintf(w, "{\"status\":\"fail\", \"reason\":\"audio not playing\"}")
		} else {
			fmt.Fprintf(w, "{\"status\":\"ok\", \"id\":%d}", value)
			value.Control.Paused = true
			value.Control.Streamer = nil
			delete(playbacks, cnt)
		}
		//fmt.Fprintf(w, "{\"status\":\"ok\", \"id\":%d}", currIndex)

	})

	http.HandleFunc("/v1/current", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		var tempResultSet map[int]playbackWebReturn
		tempResultSet = make(map[int]playbackWebReturn)

		for index, element := range playbacks {
			tempResultSet[index] = playbackWebReturn{File: element.File, IsLoaded: element.IsLoaded, Id: index}
		}

		j, err := json.Marshal(tempResultSet)
		if err != nil {
			fmt.Printf("Error: %s", err.Error())
		} else {
			fmt.Println(string(j))
		}
		fmt.Fprintf(w, string(j))
	})

	http.HandleFunc("/v1/list", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		var temp [][3]string
		files, err := ioutil.ReadDir("./sounds/")
		if err != nil {
			log.Fatal(err)
		}

		for _, f := range files {
			var soundObj [3]string
			soundObj[0] = f.Name()
			soundObj[1] = base64.StdEncoding.EncodeToString([]byte(f.Name()))
			soundObj[2] = r.URL.Host + "/v1/play?file=" + soundObj[1]
			temp = append(temp, soundObj)
		}

		j, err := json.Marshal(temp)
		if err != nil {
			fmt.Printf("Error: %s", err.Error())
		} else {
			fmt.Println(string(j))
		}
		fmt.Fprintf(w, string(j))
	})
	log.Fatal(http.ListenAndServe(":8081", nil))

}
