// Package parsercli was built for parsing url for search string
package parsercli

import (
	"context"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"sync"
	"time"
)

type resultStruct struct {
	url   string
	count int
}

type parserStruct struct {
	url  string
	html string
}

func scrap(ctx context.Context, client *http.Client, timeout int, targetURL string) string {

	ctxTimeout, timeoutCancel := context.WithTimeout(ctx, time.Duration(timeout)*time.Second)
	defer timeoutCancel()

	req, err := http.NewRequest("GET", targetURL, nil)
	if err != nil {
		fmt.Println(err)
		return ""
	}

	resp, err := client.Do(req.WithContext(ctxTimeout))

	if err != nil {
		fmt.Println(err)
		return ""
	}
	defer resp.Body.Close()
	htmlBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Println(err)
		return ""
	}

	html := string(htmlBytes)
	return html
}

func getUrls(ctx context.Context, urls string) chan string {
	var wg sync.WaitGroup

	urlList := strings.Split(urls, ",")
	urlsChan := make(chan string, len(urlList))

	defer close(urlsChan)
	for _, url := range urlList {
		wg.Add(1)
		go func(ctx context.Context, urlsChan chan string, url string) {
			defer wg.Done()
			urlsChan <- url
		}(ctx, urlsChan, url)
	}
	wg.Wait()
	return urlsChan
}

func getHTML(ctx context.Context, urlsChan chan string, timeout int) chan parserStruct {
	var wg sync.WaitGroup
	htmlChan := make(chan parserStruct)
	client := &http.Client{}

	wg.Add(1)
	go func(ctx context.Context, client *http.Client, timeout int) {
		for url := range urlsChan {
			fmt.Println("get html", url)
			html := scrap(ctx, client, timeout, url)
			htmlChan <- parserStruct{url: url, html: html}
		}
		defer wg.Done()
	}(ctx, client, timeout)
	go func() {
		wg.Wait()
		close(htmlChan)
		fmt.Println("finished")
	}()

	return htmlChan
}

func parseHTML(ctx context.Context, htmlChan chan parserStruct, searchString string) chan string {
	fmt.Println("parse")
	var wg sync.WaitGroup

	occurrencesChan := make(chan string)

	wg.Add(1)
	go func() {
		for parserInfo := range htmlChan {
			//go func(htmlChan chan string, url string) {
			fmt.Println("parse html", parserInfo)
			occurrencesChan <- parserInfo.url
			//}(occurrencesChan, url)
		}
		wg.Done()
	}()

	go func() {
		wg.Wait()
		close(occurrencesChan)
	}()

	return occurrencesChan
}

//render generate output Url - count
func render(ctx context.Context, occurrencesChan chan string) {
	fmt.Println("render")
	var mx sync.RWMutex
	go func(occurrencesChan chan string) {
		for url := range occurrencesChan {
			mx.Lock()
			fmt.Println("render html", url)
			mx.Unlock()
		}
	}(occurrencesChan)
}

/*func render(urlsResult map[string]int) {
	for key, value := range urlsResult {
		fmt.Println(key, " - ", value)
	}
}*/

//scrapCount get html by url and find by regexp count matches
/*func scrapCount(
	ctx context.Context,
	client *http.Client,
	targetURL string,
	substringRegExp *regexp.Regexp,
	timeout int,
) int {
	ctxTimeout, timeoutCancel := context.WithTimeout(ctx, time.Duration(timeout)*time.Second)
	defer timeoutCancel()

	req, err := http.NewRequest("GET", targetURL, nil)
	if err != nil {
		fmt.Println(err)
		return 0
	}

	resp, err := client.Do(req.WithContext(ctxTimeout))
	if err != nil {
		fmt.Println(err)
		return 0
	}
	defer resp.Body.Close()
	htmlBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Println(err)
		return 0
	}

	html := string(htmlBytes)
	matchesCount := 0
	if html != "" {
		matches := substringRegExp.FindAllStringIndex(html, -1)
		matchesCount = len(matches)
	}
	return matchesCount
}*/

//Parse get array of urls and parse them to find occurrences of search string
func Parse(ctx context.Context, urls, searchString string, maxGoroutines, timeout int) {
	urlsChan := getUrls(ctx, urls)
	htmlChan := getHTML(ctx, urlsChan, timeout)
	occurrencesChan := parseHTML(ctx, htmlChan, searchString)
	render(ctx, occurrencesChan)
	/*limiter := make(chan struct{}, maxGoroutines)

	urlList := strings.Split(urls, ",")

	findSubstringRegExp := regexp.MustCompile(searchURL)

	var wg sync.WaitGroup

	client := &http.Client{}

	result := resultStruct{urlsWithCount: make(map[string]int, len(urlList))}
	for _, url := range urlList {
		wg.Add(1)
		go func(ctx context.Context, targetURL string, substringRegExp *regexp.Regexp, timeout int) {
			defer wg.Done()
			defer func(ch <-chan struct{}) {
				<-ch
			}(limiter)
			limiter <- struct{}{}

			count := scrapCount(ctx, client, targetURL, substringRegExp, timeout)

			result.addCountToURL(targetURL, count)
		}(ctx, url, findSubstringRegExp, timeout)
	}
	wg.Wait()

	render(result.urlsWithCount)*/
}
