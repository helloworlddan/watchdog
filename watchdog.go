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
	"path/filepath"
	"strings"
	"sync"
	"syscall"
	"time"
)

// Version of the watchdog utility
const Version string = "0.0.1"

func main() {
	var directory = flag.String("d", ".", "Directory to watch,")
	var fileExtension = flag.String("e", ".txt", "File extension to watch.")
	var interval = flag.Int("i", 60, "Watch interval in seconds.")
	var uploadURL = flag.String("u", "http://localhost/{extension}/{filename}", "URL of the upload site to POST to.")
	var caseSensitivity = flag.Bool("c", true, "Set case sensitivity.")
	flag.Parse()

	fmt.Println("🐶")
	fmt.Println("Watchdog v" + Version)

	var caseSense string
	if *caseSensitivity {
		caseSense = "sensitive"
	} else {
		caseSense = "insensitive"
	}
	fmt.Printf("Watching '%s', lookng for case %s file extension '%s' every %d seconds to upload to '%s'.\n", *directory, caseSense, *fileExtension, *interval, *uploadURL)

	uploadTargets := make(chan string, 100) // buffer 100 files
	shutdown := make(chan os.Signal, 2)
	signal.Notify(shutdown, os.Interrupt, syscall.SIGTERM)

	fmt.Println("Hit Ctrl-C to initate shutdown.")

	go watch(uploadTargets, *directory, *fileExtension, *caseSensitivity, *interval)

	var wg sync.WaitGroup
	for {
		select {
		case <-shutdown:
			log.Println("Shutdown initiated.")
			wg.Wait()
			os.Exit(0)
		case filename := <-uploadTargets:
			wg.Add(1)
			go upload(*directory, filename, *uploadURL, &wg)
		}
	}
}

func watch(uploadTargets chan<- string, directory string, fileExtension string, caseSensitivity bool, interval int) {
	lastCheck := time.Now()
	for {
		files, _ := ioutil.ReadDir(directory)
		for _, f := range files {
			if filepath.Ext(f.Name()) != fileExtension {
				continue
			}
			info, err := os.Stat(directory + string(os.PathSeparator) + f.Name())
			if err != nil {
				log.Printf("Error: failed to get file info on %s", f.Name())
			}
			if info.ModTime().After(lastCheck) {
				log.Printf("Registering new upload target: %s", f.Name())
				uploadTargets <- f.Name()
			}
		}
		lastCheck = time.Now()
		time.Sleep(time.Second * time.Duration(interval))
	}
}

func upload(directory string, filename string, uploadURL string, wg *sync.WaitGroup) {
	defer wg.Done()
	contents, err := ioutil.ReadFile(directory + string(os.PathSeparator) + filename)
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
