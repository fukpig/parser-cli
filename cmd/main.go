package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"sync"
	"syscall"

	parsercli "github.com/fukpig/parsercli"
)

func main() {
	urls := flag.String("urls", "", "a string")
	searchURL := flag.String("search", "", "a string")
	parsingProcessesCount := flag.Int("parsingProccessesCount", 5, "an int")
	countingProcessesCount := flag.Int("countingProccessesCount", 5, "an int")
	timeout := flag.Int("timeout", 5, "an int")
	flag.Parse()

	var wg sync.WaitGroup

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

	wg.Add(1)
	//go parsercli.Parse(ctx, &wg, *urls, *searchURL, *parsingProcessesCount, *countingProcessesCount, *timeout)
	go parsercli.PipelineWithPreloadGoroutines(ctx, &wg, *urls, *searchURL, *parsingProcessesCount, *countingProcessesCount, *timeout)
	wg.Wait()
}
