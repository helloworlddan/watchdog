package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
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
		uploadTargets <- "sample_file.txt"
		time.Sleep(time.Second * time.Duration(interval))
	}
}

func upload(filename string, uploadURL string, wg *sync.WaitGroup) {
	log.Printf("Uploading %s to %s\n", filename, uploadURL)
	// TODO Open and upload file, close
	wg.Done()
}
