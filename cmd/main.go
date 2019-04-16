package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/fukpig/parser-cli"
)

func main() {
	urls := flag.String("urls", "", "a string")
	searchURL := flag.String("search", "", "a string")
	flag.Parse()

	if *urls == "" {
		fmt.Println("urlList are empty")
		os.Exit(1)
	}

	if *searchURL == "" {
		fmt.Println("search url is empty")
		os.Exit(1)
	}

	parser.Parse(*urls, *searchURL)
}
