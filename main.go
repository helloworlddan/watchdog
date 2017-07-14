package main

import (
	"bytes"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"sync"
	"syscall"
	"time"
)

// Version of the watchdog utility
const Version string = "0.0.1"

func main() {
	var directory = flag.String("d", ".", "Directory to watch,")
	var fileExtension = flag.String("e", "txt", "File extension to watch.")
	var interval = flag.Int("i", 60, "Watch interval in seconds.")
	var uploadURL = flag.String("u", "http://localhost/{extension}/{filename}", "URL of the upload site to POST to.")
	var caseSensitivity = flag.Bool("c", true, "Set case sensitivity.")
	flag.Parse()

	fmt.Println("Watchdog v" + Version)

	var caseSense string
	if *caseSensitivity {
		caseSense = "case sensitive"
	} else {
		caseSense = "case insensitive"
	}
	fmt.Printf("Watching '%s', lookng for %s file extension '%s' every %d seconds to upload to '%s'.\n", *directory, caseSense, *fileExtension, *interval, *uploadURL)

	uploadTargets := make(chan string, 100) // buffer 100 files
	shutdown := make(chan os.Signal, 2)
	signal.Notify(shutdown, os.Interrupt, syscall.SIGTERM)

	fmt.Println("Hit Ctrl-C to initate shutdown.")

	go watch(uploadTargets, *directory, *caseSensitivity, *interval)

	var wg sync.WaitGroup
	for {
		select {
		case <-shutdown:
			log.Println("Shutdown initiated.")
			wg.Wait()
			os.Exit(0)
		case filename := <-uploadTargets:
			wg.Add(1)
			go upload(filename, *uploadURL, &wg)
		}
	}
}

func watch(uploadTargets chan<- string, directory string, caseSensitivity bool, interval int) {
	for {
		log.Println("Pushing new upload target.")
		// TODO watch dir on this routine
		uploadTargets <- "sample_file2.txt"
		time.Sleep(time.Second * time.Duration(interval))
	}
}

func upload(filename string, uploadURL string, wg *sync.WaitGroup) {
	defer wg.Done()
	contents, err := ioutil.ReadFile(filename)
	if err != nil {
		log.Printf("Error: failed to open '%s'\n", filename)
	}
	parts := strings.Split(filename, ".")
	url := strings.Replace(strings.Replace(uploadURL, "{extension}", parts[1], 1), "{filename}", parts[0], 1)
	request, err := http.NewRequest("POST", url, bytes.NewBuffer(contents))
	client := &http.Client{}
	response, err := client.Do(request)
	if err != nil {
		log.Printf("Error: failed to POST to '%s'\n", url)
		return
	}
	defer response.Body.Close()
	log.Printf("'%s' -> POST to '%s' %s\n", filename, url, response.Status)
}
