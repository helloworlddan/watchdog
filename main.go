package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"sync"
	"syscall"
)

// Version of the watchdog utility
const Version string = "0.0.1"

func main() {
	var directory = flag.String("d", ".", "Directory to watch,")
	var fileExtension = flag.String("e", "txt", "File extension to watch.")
	var interval = flag.Int("i", 60, "Watch interval in seconds.")
	var uploadURL = flag.String("u", "http://localhost", "URL of the upload site to POST to.")
	var caseSensitivity = flag.Bool("c", true, "Set case sensitivity.")
	flag.Parse()

	fmt.Println("Watchdog v" + Version)
	var caseSense string
	if *caseSensitivity {
		caseSense = "case sensitive"
	} else {
		caseSense = "case insensitive"
	}
	fmt.Printf("Watchdog configured to watch '%s' and look for %s file extension '%s' every %d seconds to upload to '%s'.\n", *directory, caseSense, *fileExtension, *interval, *uploadURL)

	uploadTargets := make(chan *os.File)
	shutdown := make(chan os.Signal, 2)
	signal.Notify(shutdown, os.Interrupt, syscall.SIGTERM)

	sample, err := os.Create("sample_file.txt")
	if err != nil {
		log.Fatal("Failed to create sample file.")
	}
	uploadTargets <- sample

	var wg sync.WaitGroup
	for {
		select {
		case <-shutdown:
			fmt.Println("Shutdown initiated.")
			wg.Wait()
			os.Exit(0)
		case f := <-uploadTargets:
			fmt.Println("Starting new upload.")
			wg.Add(1)
			go upload(f, *uploadURL, &wg)
		}
	}
}

func upload(f *os.File, uploadURL string, wg *sync.WaitGroup) {
	fmt.Printf("Uploading %s to %s", *f, uploadURL)
	wg.Done()
}
