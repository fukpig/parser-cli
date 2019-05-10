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
	searchString := flag.String("search", "", "a string")
	pipelineType := flag.String("pipelineType", "common", "a string")
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

	if *searchString == "" {
		fmt.Println("search url is empty")
		os.Exit(1)
	}

	c := make(chan os.Signal)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-c
		cancel()
		fmt.Println("cancelled")
		wg.Wait()
		fmt.Println("end")
		os.Exit(1)
	}()

	wg.Add(1)

	config := parsercli.PipelineConfig{
		ParsingProcessesCount:  *parsingProcessesCount,
		CountingProcessesCount: *countingProcessesCount,
		Timeout:                *timeout,
		PipelineType:           *pipelineType,
		Wg:                     &wg,
		Urls:                   *urls,
		SearchString:           *searchString,
	}

	go parsercli.Parse(ctx, config)
	wg.Wait()
}
