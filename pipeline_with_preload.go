package parsercli

import (
	"context"
	"fmt"
	"net/http"
	"regexp"
	"strings"
	"sync"
)

type pipelinePreload struct {
	config PipelineConfig
}

func (p pipelinePreload) getUrlsPreload(urls string) (chan string, int) {
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

func (p pipelinePreload) getHTMLPreload(ctx context.Context, params htmlParams, urlsChan chan string, parsingProcessesCount int) chan parserStruct {
	var wg sync.WaitGroup
	wg.Add(parsingProcessesCount)
	htmlChan := make(chan parserStruct)
	client := &http.Client{}
	go func() {
		for i := 1; i <= parsingProcessesCount; i++ {
			go func(params htmlParams) {
				defer wg.Done()
				for {
					url, ok := <-urlsChan
					if !ok {
						return
					}
					html, err := scrap(ctx, client, params.timeout, url)
					if err != nil {
						continue
					}
					htmlChan <- parserStruct{url: url, html: html}
				}
			}(params)
		}

		wg.Wait()
		close(htmlChan)
	}()
	return htmlChan
}

func (p pipelinePreload) parseHTMLPreload(htmlChan chan parserStruct, searchString string, countingProcessesCount, urlsCount int) chan resultStruct {
	var wg sync.WaitGroup
	wg.Add(countingProcessesCount)
	occurrencesChan := make(chan resultStruct, urlsCount)
	findSubstringRegExp := regexp.MustCompile(searchString)
	go func() {
		for i := 1; i <= countingProcessesCount; i++ {
			go func(htmlChan chan parserStruct) {
				defer wg.Done()
				for {
					info, ok := <-htmlChan
					if !ok {
						return
					}
					count := countMatches(info.html, findSubstringRegExp)
					occurrencesChan <- resultStruct{url: info.url, count: count}
				}
			}(htmlChan)
		}
		wg.Wait()
		close(occurrencesChan)
	}()
	return occurrencesChan
}

func (p pipelinePreload) render(occurrencesChan chan resultStruct) {
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

func (p pipelinePreload) run(ctx context.Context) {
	params := htmlParams{
		timeout: p.config.Timeout,
	}

	urlsChan, urlsCount := p.getUrlsPreload(p.config.Urls)
	htmlChan := p.getHTMLPreload(ctx, params, urlsChan, p.config.ParsingProcessesCount)
	occurrencesChan := p.parseHTMLPreload(htmlChan, p.config.SearchString, p.config.CountingProcessesCount, urlsCount)
	p.render(occurrencesChan)
}
