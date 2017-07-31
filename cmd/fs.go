package cmd

import (
	"bytes"
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

	"github.com/spf13/cobra"
)

var (
	directory, fileExtension, uploadURL string
	caseSensitivity                     bool
	interval                            int
)

func init() {
	RootCmd.AddCommand(fsCmd)
	fsCmd.Flags().StringVarP(&directory, "directory", "d", ".", "Directory to watch.")
	fsCmd.Flags().StringVarP(&fileExtension, "extension", "e", ".txt", "Extension to watch.")
	fsCmd.Flags().IntVarP(&interval, "interval", "i", 60, "Watch interval in seconds.")
	fsCmd.Flags().StringVarP(&uploadURL, "upload", "u", "http://localhost/{extension}/{filename}", "URL of the upload site to POST to.")
	fsCmd.Flags().BoolVarP(&caseSensitivity, "case-sensitivity", "c", true, "Set case sensitivity.")
}

var fsCmd = &cobra.Command{
	Use:   "fs",
	Short: "Watch a filesystem and upload changes to HTTP endpoint",
	Long:  ``,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("🐶")

		var caseSense string
		if caseSensitivity {
			caseSense = "sensitive"
		} else {
			caseSense = "insensitive"
		}
		log.Printf("watching '%s', looking for case %s file extension '%s' every %d seconds to upload to '%s'.\n", directory, caseSense, fileExtension, interval, uploadURL)

		uploadTargets := make(chan string, 100) // buffer 100 files
		shutdown := make(chan os.Signal, 2)
		signal.Notify(shutdown, os.Interrupt, syscall.SIGTERM)

		fmt.Println("hit Ctrl-C to initate shutdown.")

		go watchFS(uploadTargets, directory, fileExtension, caseSensitivity, interval)

		var wg sync.WaitGroup
		for {
			select {
			case <-shutdown:
				log.Println("shutdown initiated.")
				wg.Wait()
				os.Exit(0)
			case filename := <-uploadTargets:
				wg.Add(1)
				go upload(directory, filename, uploadURL, &wg)
			}
		}
	},
}

func watchFS(uploadTargets chan<- string, directory string, fileExtension string, caseSensitivity bool, interval int) {
	lastCheck := time.Now()
	for {
		files, _ := ioutil.ReadDir(directory)
		for _, f := range files {
			if filepath.Ext(f.Name()) != fileExtension {
				continue
			}
			info, err := os.Stat(directory + string(os.PathSeparator) + f.Name())
			if err != nil {
				log.Printf("failed to get file info on %s", f.Name())
				continue
			}
			if info.ModTime().After(lastCheck) {
				log.Printf("registering new upload target: %s", f.Name())
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
		log.Printf("failed to open '%s'\n", filename)
	}
	parts := strings.Split(filename, ".")
	url := strings.Replace(strings.Replace(uploadURL, "{extension}", parts[1], 1), "{filename}", parts[0], 1)
	request, err := http.NewRequest("POST", url, bytes.NewBuffer(contents))
	client := &http.Client{}
	response, err := client.Do(request)
	if err != nil {
		log.Printf("failed to POST to '%s'\n", url)
		return
	}
	defer response.Body.Close()
	log.Printf("'%s' -> POST to '%s' %s\n", filename, url, response.Status)
}
