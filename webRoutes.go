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
	"github.com/h2non/filetype"
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
		fmt.Fprintf(w, "{\"status\":\"fail\", \"reason\":\""+err.Error()+"\"}")
		return
	}
	loop := r.URL.Query().Get("loop") // Retrieve the loop value from the query string
	loopBool := false
	if loop == "true" {
		loopBool = true
	}

	wantedId := r.URL.Query().Get("id") // Retrieve the id value from the query string, it's optional

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

	buf, _ := ioutil.ReadFile("sounds/" + string(bytArr[:]))

	kind, _ := filetype.Match(buf)
	if kind == filetype.Unknown {
		fmt.Println("Unknown file type")
		fmt.Fprintf(w, "{\"status\":\"fail\", \"reason\":\"file has unknown type\"}")
		return
	}
	if kind.MIME.Type != "audio" {
		fmt.Fprintf(w, "{\"status\":\"fail\", \"reason\":\"file is not an audio file\"}")
		return
	}

	var currIndex = len(playbacks) // Create a new index for the playback

	if len(wantedId) > 0 { // If the id is set, check if it is already in use
		id, err := strconv.Atoi(wantedId)
		if err != nil {
			fmt.Fprintf(w, "{\"status\":\"fail\", \"reason\":\"id is not a number\"}")
			return
		}
		if _, ok := playbacks[id]; ok {
			fmt.Fprintf(w, "{\"status\":\"fail\", \"reason\":\"id is already in use\"}")
			return
		}
		currIndex = id
	}

	fmt.Fprintf(w, "{\"status\":\"ok\", \"id\":%d}", currIndex) // Return a JSON object to the user

	go PlaySound(string(bytArr[:]), currIndex, loopBool) // Play the sound

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
		fmt.Fprintf(w, "{\"status\":\"fail\", \"reason\":\""+err.Error()+"\"}")
		return
	}
	// Loop through the files and add the file name to the temp array
	// Also triggers the buffer process for the file
	for _, f := range files {
		buf, _ := ioutil.ReadFile("sounds/" + string(f.Name()))

		kind, _ := filetype.Match(buf)
		if kind == filetype.Unknown {
			fmt.Println("Unknown file type")

			continue
		}
		if kind.MIME.Type != "audio" {
			fmt.Println("Not an audio file")
			continue
		}
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
		fmt.Fprintf(w, "{\"status\":\"fail\", \"reason\":\""+err.Error()+"\"}")
		return
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

	buf, _ := ioutil.ReadFile("sounds/" + string(bytArr[:]))

	kind, _ := filetype.Match(buf)
	if kind == filetype.Unknown {
		fmt.Println("Unknown file type")
		fmt.Fprintf(w, "{\"status\":\"fail\", \"reason\":\"file has unknown type\"}")
		return
	}
	if kind.MIME.Type != "audio" {
		fmt.Fprintf(w, "{\"status\":\"fail\", \"reason\":\"file is not an audio file\"}")
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

		buf, _ := ioutil.ReadFile("sounds/" + f.Name())

		kind, _ := filetype.Match(buf)
		fmt.Println(f.Name() + " " + kind.MIME.Type)
		if kind == filetype.Unknown {
			fmt.Println("Unknown file type")
			continue
		}
		if kind.MIME.Type != "audio" {
			fmt.Println("Not an audio file")
			continue
		}

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

func handleRemaining(w http.ResponseWriter, r *http.Request) {
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
	fmt.Println(cnt)
	plyB := playbacks[cnt]
	// fmt.Println(beep.SampleRate.D(plyB.Streamer.Stream().Len()))
	seeker := plyB.Streamer
	format := plyB.Format
	n := plyB.Format.SampleRate // Streamer.Stream() // .At(beep.SampleRate.D(plyB.Streamer.Stream().Len()))

	if seeker != nil {
		fmt.Println(format.SampleRate)
		// fmt.Println(plyB.Seeker.)
		position := plyB.Format.SampleRate.D(seeker.Position())
		length := plyB.Format.SampleRate.D(seeker.Len())
		remaining := length - position
		if remaining == 0 {
			plyB.Done <- true
		}
		fmt.Println(position)
		fmt.Fprintf(w, "{\"status\":\"ok\", \"SampleRate\":%d, \"Length\":%d, \"Position\":%d, \"Remaining\": %d, \"LengthSec\":\"%v\", \"PosSec\":\"%v\", \"RemaningSec\":\"%v\"}", n, length, position, remaining, length, position, remaining)
	} else {
		fmt.Println("Seeker is nil")
		fmt.Fprintf(w, "{\"status\":\"ok\", \"SampleRate\":%d}", n)
	}

}

func handleRoot(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Soundr is running.")
}
