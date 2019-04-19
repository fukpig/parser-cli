package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	parsercli "github.com/fukpig/parsercli"
)

func main() {
	urls := flag.String("urls", "", "a string")
	searchURL := flag.String("search", "", "a string")
	maxGoroutines := flag.Int("processes", 2, "an int")
	timeout := flag.Int("timeput", 5, "an int")
	flag.Parse()

	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)

	if *urls == "" {
		fmt.Println("urlList are empty")
		os.Exit(1)
	}

	if *searchURL == "" {
		fmt.Println("search url is empty")
		os.Exit(1)
	}

	c := make(chan os.Signal)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-c
		cancel()
		os.Exit(1)
	}()

	parsercli.Parse(ctx, *urls, *searchURL, *maxGoroutines, *timeout)
}
