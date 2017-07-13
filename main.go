package main

import (
	"flag"
	"fmt"
	"strconv"
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
	fmt.Println("Watchdog configured to watch '" + *directory + "' and look for " + caseSense + " file extension '" + *fileExtension + "' every " + strconv.Itoa(*interval) + " seconds to upload to '" + *uploadURL + "'.")
}
