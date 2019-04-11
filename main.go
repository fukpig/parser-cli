package main

import (
	"flag"
	"fmt"
	"os"
)

func main() {
	urls := flag.String("urls", "", "a string")
	searchUrl := flag.String("search", "", "a string")
	flag.Parse()

	if *urls == "" {
		fmt.Println("urlList are empty")
		os.Exit(1)
	}

	if *searchUrl == "" {
		fmt.Println("search url is empty")
		os.Exit(1)
	}

	parse(*urls, *searchUrl)
}
