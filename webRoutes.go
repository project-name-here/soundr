package main

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strconv"

	"github.com/faiface/beep/speaker"
)

func handlePlay(w http.ResponseWriter, r *http.Request) {
	// Rejct everything else then GET requests
	if r.Method != "GET" {
		http.Error(w, "Method is not supported.", http.StatusNotFound)
		return
	}
	w.Header().Set("Content-Type", "application/json")  // Set the content type to json
	var cnt = r.URL.Query().Get("file")                 // Retrieve the file name from the query string
	bytArr, err := base64.StdEncoding.DecodeString(cnt) // Decode the base64 string
	if err != nil {
		log.Fatal(err)
	}

	t, err := os.Stat("./sounds/" + string(bytArr[:])) // Check if the file exists
	if errors.Is(err, os.ErrNotExist) {
		w.WriteHeader(400)
		fmt.Fprintf(w, "{\"status\":\"fail\", \"reason\":\"file not found\"}")
		return
	}

	if t.IsDir() { // Make sure it is not a folder we are trying to play
		w.WriteHeader(400)
		fmt.Fprintf(w, "{\"status\":\"fail\", \"reason\":\"target is folder\"}")
		return
	}

	var currIndex = len(playbacks)                              // Create a new index for the playback
	fmt.Fprintf(w, "{\"status\":\"ok\", \"id\":%d}", currIndex) // Return a JSON object to the user

	go PlaySound(string(bytArr[:]), currIndex) // Play the sound

}

// Handle Buffering

func handleBufferAll(w http.ResponseWriter, r *http.Request) {
	// Rejct everything else then GET requests
	if r.Method != "GET" {
		http.Error(w, "Method is not supported.", http.StatusNotFound)
		return
	}
	w.Header().Set("Content-Type", "application/json") // Set the content type to json

	var temp []string
	files, err := ioutil.ReadDir("./sounds/") // Read the directory
	if err != nil {
		log.Fatal(err)
	}
	// Loop through the files and add the file name to the temp array
	// Also triggers the buffer process for the file
	for _, f := range files {
		temp = append(temp, f.Name())
		go BufferSound(f.Name())
	}
	// Return the amount of files buffered
	fmt.Fprintf(w, "{\"status\":\"ok\", \"amount\":%d}", len(temp))

}

func handleBuffer(w http.ResponseWriter, r *http.Request) {
	// Rejct everything else then GET requests
	if r.Method != "GET" {
		http.Error(w, "Method is not supported.", http.StatusNotFound)
		return
	}
	w.Header().Set("Content-Type", "application/json")  // Set the content type to json
	var cnt = r.URL.Query().Get("file")                 // Retrieve the file name from the query string
	bytArr, err := base64.StdEncoding.DecodeString(cnt) // Decode the base64 string
	if err != nil {
		log.Fatal(err)
	}

	t, err := os.Stat("./sounds/" + string(bytArr[:])) // Check if the file exists
	if errors.Is(err, os.ErrNotExist) {
		w.WriteHeader(400)
		fmt.Fprintf(w, "{\"status\":\"fail\", \"reason\":\"file not found\"}")
		return
	}

	if t.IsDir() { // Make sure it is not a folder we are trying to play
		w.WriteHeader(400)
		fmt.Fprintf(w, "{\"status\":\"fail\", \"reason\":\"target is folder\"}")
		return
	}

	fmt.Fprintf(w, "{\"status\":\"ok\"}")
	go BufferSound(string(bytArr[:]))

}

// Handeling Stop

func handleStop(w http.ResponseWriter, r *http.Request) {
	// Rejct everything else then GET requests
	if r.Method != "GET" {
		http.Error(w, "Method is not supported.", http.StatusNotFound)
		return
	}
	w.Header().Set("Content-Type", "application/json")   // Set the content type to json
	var cnt, err = strconv.Atoi(r.URL.Query().Get("id")) // Retrieve the id, first convert it to an int
	if err != nil {
		fmt.Fprintf(w, "{\"status\":\"fail\", \"reason\":\"invalid id\"}")
	}

	value, ok := playbacks[cnt] // Get value from playbacks map
	if !ok {
		w.WriteHeader(400)
		fmt.Fprintf(w, "{\"status\":\"fail\", \"reason\":\"audio not playing\"}")
	} else {
		fmt.Fprintf(w, "{\"status\":\"ok\"}")
		// Stop by pausing first then, set the streamer to nil. Finally delete it from the map
		value.Control.Paused = true
		value.Control.Streamer = nil
		delete(playbacks, cnt)
	}
}

func handleStopAll(w http.ResponseWriter, r *http.Request) {
	// Rejct everything else then GET requests
	if r.Method != "GET" {
		http.Error(w, "Method is not supported.", http.StatusNotFound)
		return
	}
	w.Header().Set("Content-Type", "application/json") // Set the content type to json

	// Pause and stop all playbacks
	for _, v := range playbacks {
		v.Control.Paused = true
		v.Control.Streamer = nil
	}
	speaker.Clear() // Clear the speaker and make it shut up

	// Reset the map
	playbacks = make(map[int]playback)

	fmt.Fprintf(w, "{\"status\":\"ok\"}")
}

func handleCurrent(w http.ResponseWriter, r *http.Request) {
	// Rejct everything else then GET requests
	if r.Method != "GET" {
		http.Error(w, "Method is not supported.", http.StatusNotFound)
		return
	}
	w.Header().Set("Content-Type", "application/json") // Set the content type to json

	var tempResultSet map[int]playbackWebReturn = make(map[int]playbackWebReturn) // Create a new map to store the results
	// Iterate through the playbacks map and add important information to the tempResultSet map
	for index, element := range playbacks {
		tempResultSet[index] = playbackWebReturn{File: element.File, IsLoaded: element.IsLoaded, Id: index}
	}
	// Convert the map to a JSON object and return it to the user
	j, err := json.Marshal(tempResultSet)
	if err != nil {
		fmt.Printf("Error: %s", err.Error())
	} else {
		fmt.Println(string(j))
		fmt.Fprintf(w, string(j))
	}
}

func handleListing(w http.ResponseWriter, r *http.Request) {
	// Rejct everything else then GET requests
	if r.Method != "GET" {
		http.Error(w, "Method is not supported.", http.StatusNotFound)
		return
	}
	w.Header().Set("Content-Type", "application/json") // Set the content type to json

	var temp [][3]string
	files, err := ioutil.ReadDir("./sounds/") // Find all files in the sounds directory
	if err != nil {
		log.Fatal(err)
	}

	// Add the file data to the temp array
	for _, f := range files {
		var soundObj [3]string
		soundObj[0] = f.Name()
		soundObj[1] = base64.StdEncoding.EncodeToString([]byte(f.Name()))
		soundObj[2] = r.URL.Host + "/v1/play?file=" + soundObj[1]
		temp = append(temp, soundObj)
	}
	// Convert the array to a JSON object and return it to the user
	j, err := json.Marshal(temp)
	if err != nil {
		fmt.Printf("Error: %s", err.Error())
	} else {
		fmt.Println(string(j))
		fmt.Fprintf(w, string(j))
	}
}
