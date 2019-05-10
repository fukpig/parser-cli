package parsercli

import (
	"context"
	"fmt"
	"net/http"
	"regexp"
	"strings"
	"sync"
)

type pipelineCommon struct {
	config PipelineConfig
}

func (p pipelineCommon) getUrls(urls string) (chan string, int) {
	urlList := strings.Split(urls, ",")
	urlsChan := make(chan string, len(urlList))

	go func(urlsChan chan string, urlList []string) {
		defer close(urlsChan)

		for _, url := range urlList {
			urlsChan <- url
		}
	}(urlsChan, urlList)
	return urlsChan, len(urlList)
}

func (p pipelineCommon) getHTML(ctx context.Context, params htmlParams, urlsChan chan string) chan parserStruct {

	htmlChan := make(chan parserStruct)
	client := &http.Client{}

	go func(params htmlParams) {
		var wg sync.WaitGroup
		limiter := make(chan struct{}, params.maxGoroutines)

		for url := range urlsChan {
			wg.Add(1)
			go func(params htmlParams, limiter chan struct{}, url string) {
				defer wg.Done()
				defer func(ch <-chan struct{}) {
					<-ch
				}(limiter)
				limiter <- struct{}{}
				html, err := scrap(ctx, client, params.timeout, url)
				if err != nil {
					return
				}
				htmlChan <- parserStruct{url: url, html: html}

			}(params, limiter, url)
		}

		wg.Wait()
		close(htmlChan)

	}(params)
	return htmlChan
}

func (p pipelineCommon) parseHTML(
	htmlChan chan parserStruct,
	searchString string,
	maxGoroutines, urlsCount int) chan resultStruct {
	occurrencesChan := make(chan resultStruct, urlsCount)
	findSubstringRegExp := regexp.MustCompile(searchString)

	limiter := make(chan struct{}, maxGoroutines)
	go func() {
		var wg sync.WaitGroup
		for parserInfo := range htmlChan {
			wg.Add(1)
			go func(info parserStruct) {
				defer func(ch <-chan struct{}) {
					<-ch
				}(limiter)
				limiter <- struct{}{}
				count := countMatches(info.html, findSubstringRegExp)
				occurrencesChan <- resultStruct{url: info.url, count: count}
				wg.Done()
			}(parserInfo)
		}
		wg.Wait()
		close(occurrencesChan)
	}()

	return occurrencesChan
}

//render generate output Url - count
func (p pipelineCommon) render(occurrencesChan chan resultStruct) {
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		for resultStruct := range occurrencesChan {
			fmt.Println(resultStruct.url, "-", resultStruct.count)
		}
	}()
	wg.Wait()
}

func (p pipelineCommon) run(ctx context.Context) {
	params := htmlParams{
		timeout:       p.config.Timeout,
		maxGoroutines: p.config.ParsingProcessesCount,
	}

	urlsChan, urlsCount := p.getUrls(p.config.Urls)
	htmlChan := p.getHTML(ctx, params, urlsChan)
	occurrencesChan := p.parseHTML(htmlChan, p.config.SearchString, p.config.CountingProcessesCount, urlsCount)
	p.render(occurrencesChan)
}
