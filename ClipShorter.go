package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/atotto/clipboard"
	"io/ioutil"
	"log"
	"net/http"
	"strings"
)

var readerChan chan string
var writerChan chan string
var shortenChan chan string

// needed to extract shortened url from google
type apiData struct {
	Kind, Id, LongUrl string
}

func main() {

	readerChan = make(chan string)
	writerChan = make(chan string)
	shortenChan = make(chan string)

	fmt.Println("ClipShorter now watches the clipboard.")

	// start reading the clipboard in an endless loop
	go func() {
		for {
			readClipBoard()
		}
	}()

	// start shortening service in an endless loop
	go func() {
		for {
			shortUrl()
		}
	}()

	// start writing to the clipboard in an endless loop
	go func() {
		for {
			writeClipBoard()
		}
	}()

	for {
		fmt.Printf("%s", <-writerChan)
	}
}

func readClipBoard() {
	clip, _ := clipboard.ReadAll()
	if isUrl(clip) && len(clip) > 50 {
		readerChan <- clip
	}
}

func shortUrl() {
	url := <-readerChan

	body := bytes.NewBufferString(fmt.Sprintf(`{"longUrl": "%s"}`, url))
	request, _ := http.NewRequest("POST", "https://www.googleapis.com/urlshortener/v1/url", body)
	request.Header.Add("Content-Type", "application/json")

	client := http.Client{}
	response, _ := client.Do(request)

	output, _ := ioutil.ReadAll(response.Body)
	response.Body.Close()

	var data apiData
	json.Unmarshal(output, &data)
	shortenChan <- data.Id
}

func writeClipBoard() {
	shortened := <-shortenChan
	err := clipboard.WriteAll(shortened)
	if err != nil {
		log.Fatal(err)
	}
	writerChan <- "wrote " + shortened + " to the clipboard\n"
}

func isUrl(url string) bool {
	// base case: already shortened
	if strings.Contains(url, "goo.gl") {
		return false
	}

	// check if is valid url
	if strings.HasPrefix(url, "http://") || strings.HasPrefix(url, "https://") {
		return true
	}
	return false
}
