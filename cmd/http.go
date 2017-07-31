package cmd

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/spf13/cobra"
)

var (
	downloadDirectory, downloadURL, downloadExtension string
	watchInterval                                     int
)

type objectStat struct {
	Hash      string    `json:"hash"`
	Fext      string    `json:"type"`
	Name      string    `json:"name"`
	WriteTime time.Time `json:"time"`
}

func init() {
	RootCmd.AddCommand(httpCmd)
	httpCmd.Flags().StringVarP(&downloadDirectory, "directory", "d", ".", "Directory to download to.")
	httpCmd.Flags().StringVarP(&downloadExtension, "extension", "e", ".txt", "Extension to watch.")
	httpCmd.Flags().StringVarP(&downloadURL, "url", "u", "http://localhost/{extension}/{filename}", "URL to download from.")
	httpCmd.Flags().IntVarP(&watchInterval, "interval", "i", 60, "Watch interval in seconds.")
}

var httpCmd = &cobra.Command{
	Use:   "http",
	Short: "Watch an URL and download to a filesystem directory",
	Long: `Watch an HTTP endpoint listing for a specific timestamp and download it to local filesystem if time is younger than previous iteration.
	NOTE: 
	This command assumes that calling the url's base ('url' without {filename} & {extension}) will return some 
	json listing of file info, including modification time in RFC.3339 (this is what we are watching).`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("🐶")
		log.Printf("watching '%s' looking for extension '%s' every %d seconds to download to '%s'.\n", downloadURL, downloadExtension, watchInterval, downloadDirectory)

		downloadTargets := make(chan string, 100) // buffer 100 files
		shutdown := make(chan os.Signal, 2)
		signal.Notify(shutdown, os.Interrupt, syscall.SIGTERM)

		fmt.Println("hit Ctrl-C to initate shutdown.")

		go watchHTTP(downloadTargets, downloadURL, downloadExtension, watchInterval)

		var wg sync.WaitGroup
		for {
			select {
			case <-shutdown:
				log.Println("shutdown initiated.")
				wg.Wait()
				os.Exit(0)
			case filename := <-downloadTargets:
				wg.Add(1)
				go download(downloadURL, downloadDirectory, filename, &wg)
			}
		}
	},
}

func watchHTTP(downloadTargets chan<- string, url, extension string, interval int) {
	lastCheck := time.Now()
	ext := strings.Replace(extension, ".", "", 1)
	for {
		baseURL := strings.Replace(strings.Replace(url, "/{filename}", "", -1), "{extension}", ext, -1)
		req, err := http.NewRequest("GET", baseURL, nil)
		if err != nil {
			log.Printf("failed to create request for %s : %v \n", baseURL, err)
			continue
		}
		client := &http.Client{}
		resp, err := client.Do(req)
		if err != nil {
			log.Printf("failed to read response for %s : %v \n", baseURL, err)
			time.Sleep(time.Second * time.Duration(interval))
			continue
		}
		defer resp.Body.Close()
		objects := make([]objectStat, 10, 10)

		rawBytes, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			log.Printf("failed get raw bytes from %v : %v \n", resp.Body, err)
			continue
		}
		err = json.Unmarshal(rawBytes, &objects)
		if err != nil {
			log.Printf("failed decode json for %s '%s' : %v \n", baseURL, resp.Body, err)
			continue
		}

		for _, obj := range objects {
			if obj.WriteTime.After(lastCheck) {
				log.Printf("registering new download target %s.%s \n", obj.Name, obj.Fext)
				downloadTargets <- fmt.Sprintf("%s.%s", obj.Name, obj.Fext)
			}
		}

		lastCheck = time.Now()
		time.Sleep(time.Second * time.Duration(interval))
	}
}

func download(baseURL, directory, filename string, wg *sync.WaitGroup) {
	defer wg.Done()
	split := strings.Split(filename, ".")
	file := split[0]
	extension := split[1]

	url := strings.Replace(strings.Replace(baseURL, "{filename}", file, -1), "{extension}", extension, -1)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		log.Printf("failed to create request for %s : %v \n", url, err)
		return
	}
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Printf("failed to read response for %s : %v \n", url, err)
		return
	}
	defer resp.Body.Close()

	targetPath := directory + string(os.PathSeparator) + filename
	target, err := os.Create(targetPath)
	if err != nil {
		log.Printf("failed to open target file %s : %v \n", url, err)
		return
	}
	defer target.Close()

	length, err := io.Copy(target, resp.Body)
	if err != nil {
		log.Printf("failed to copy data for %s : %v \n", url, err)
		return
	}
	log.Printf("GET %s -> '%s' %s, %d bytes written\n", url, filename, resp.Status, length)
}
