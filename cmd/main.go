package main

import (
	"flag"
	"fmt"
	"os"

	parsercli "github.com/fukpig/parsercli"
)

func main() {
	urls := flag.String("urls", "", "a string")
	searchURL := flag.String("search", "", "a string")
	maxGoroutines := flag.Int("processes", 2, "an int")
	flag.Parse()

	if *urls == "" {
		fmt.Println("urlList are empty")
		os.Exit(1)
	}

	if *searchURL == "" {
		fmt.Println("search url is empty")
		os.Exit(1)
	}

	parsercli.Parse(*urls, *searchURL, *maxGoroutines)
}
