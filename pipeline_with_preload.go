package parsercli

import (
	"fmt"
	"net/http"
	"regexp"
	"strings"
	"sync"
)

type pipelinePreload struct {
	config PipelineConfig
}

func (p pipelinePreload) getUrlsPreload(urls string) chan string {
	urlList := strings.Split(urls, ",")
	urlsChan := make(chan string, len(urlList))

	go func(urlsChan chan string, urlList []string) {
		defer close(urlsChan)
		for _, url := range urlList {
			urlsChan <- url
		}
	}(urlsChan, urlList)
	return urlsChan
}

func (p pipelinePreload) getHTMLPreload(params htmlParams, urlsChan chan string, parsingProcessesCount int) chan parserStruct {
	var wg sync.WaitGroup
	wg.Add(parsingProcessesCount)
	htmlChan := make(chan parserStruct)
	client := &http.Client{}
	params.client = client
	go func() {
		for i := 1; i <= parsingProcessesCount; i++ {
			go func(params htmlParams) {
				defer wg.Done()
				for {
					select {
					case url, ok := <-urlsChan:
						if !ok {
							return
						}
						html := scrap(params.ctx, params.client, params.timeout, url)
						htmlChan <- parserStruct{url: url, html: html}
					}
				}
			}(params)
		}

		wg.Wait()
		close(htmlChan)
	}()
	return htmlChan
}

func (p pipelinePreload) parseHTMLPreload(htmlChan chan parserStruct, searchString string, countingProcessesCount int) chan resultStruct {
	var wg sync.WaitGroup
	wg.Add(countingProcessesCount)
	occurrencesChan := make(chan resultStruct)
	findSubstringRegExp := regexp.MustCompile(searchString)
	go func() {
		for i := 1; i <= countingProcessesCount; i++ {
			go func(htmlChan chan parserStruct) {
				defer wg.Done()
				for {
					select {
					case info, ok := <-htmlChan:
						if !ok {
							return
						}
						count := countMatches(info.html, findSubstringRegExp)
						occurrencesChan <- resultStruct{url: info.url, count: count}
					}
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

func (p pipelinePreload) run(urls, searchString string) {
	htmlParams := htmlParams{
		ctx:     p.config.Ctx,
		timeout: p.config.Timeout,
	}

	urlsChan := p.getUrlsPreload(urls)
	htmlChan := p.getHTMLPreload(htmlParams, urlsChan, p.config.ParsingProcessesCount)
	occurrencesChan := p.parseHTMLPreload(htmlChan, searchString, p.config.CountingProcessesCount)
	p.render(occurrencesChan)
}
