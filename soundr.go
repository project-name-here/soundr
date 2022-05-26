package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/faiface/beep"
)

type playback struct {
	File     string
	IsLoaded bool
	Streamer beep.Streamer
	Control  *beep.Ctrl
	Loop     bool
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

type Configuration struct {
	Port int
}

var playbacks map[int]playback

var streamMap map[string]streamBuf

func main() {

	fmt.Println("Welcome to Soundr!")

	playbacks = make(map[int]playback)
	streamMap = make(map[string]streamBuf)

	// Create /sounds if not exists
	if _, err := os.Stat("./sounds"); os.IsNotExist(err) {
		fmt.Println("Created /sounds folder")
		os.Mkdir("./sounds", 0777)
	}

	// Handle config
	fmt.Println("Opening conf.json")
	file, fOpenError := os.Open("conf.json") // Try to open the file

	if errors.Is(fOpenError, os.ErrNotExist) { // If it does not exist, create it
		fmt.Println("Creating conf.json")
		file, fOpenError = os.Create("conf.json")
		if fOpenError != nil {
			log.Fatal(fOpenError)
		}
		defer file.Close()
		fmt.Println("Writing to conf.json")
		// Write the default config to the file
		json.NewEncoder(file).Encode(Configuration{
			Port: 8080,
		})
		fmt.Println("Wrote to conf.json")
	}
	defer file.Close()
	// Decode the config
	decoder := json.NewDecoder(file)
	configuration := Configuration{}
	err := decoder.Decode(&configuration)

	if err != nil {
		fmt.Println("error:", err)
	}

	// Web server stuff
	// Play route, takes file as parameter, file is base64 encoded
	http.HandleFunc("/v1/play", handlePlay)
	// Buffer route, buffers file
	http.HandleFunc("/v1/buffer", handleBuffer)
	http.HandleFunc("/v1/bufferAll", handleBufferAll)
	http.HandleFunc("/v1/stop", handleStop)
	http.HandleFunc("/v1/stopAll", handleStopAll)
	http.HandleFunc("/v1/current", handleCurrent)
	http.HandleFunc("/v1/list", handleListing)

	fmt.Println("Listening on port " + fmt.Sprint(configuration.Port))
	log.Fatal(http.ListenAndServe(":"+fmt.Sprint(configuration.Port), nil))
}
