// Package parsercli was built for parsing url for search string
package parsercli

import (
	"context"
	"fmt"
	"io/ioutil"
	"net/http"
	"regexp"
	"strings"
	"sync"
	"time"
)

type resultStruct struct {
	urlsWithCount map[string]int
	locker        sync.RWMutex
}

var (
	ctx    context.Context
	cancel context.CancelFunc
)

//render generate output Url - count
func render(urlsResult map[string]int) {
	for key, value := range urlsResult {
		fmt.Println(key, " - ", value)
	}
}

//scrapCount get html by url and find by regexp count matches
func scrapCount(client *http.Client, targetURL string, substringRegExp *regexp.Regexp, timeout int) int {
	ctx, cancel = context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	req, err := http.NewRequest("GET", targetURL, nil)
	if err != nil {
		fmt.Println(err)
		return 0
	}

	resp, err := http.DefaultClient.Do(req.WithContext(ctx))
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
}

//Parse get array of urls and parse them to find occurrences of search string
func Parse(urls, searchURL string, maxGoroutines, timeout int) {
	limiter := make(chan struct{}, maxGoroutines)

	urlList := strings.Split(urls, ",")

	findSubstringRegExp := regexp.MustCompile(searchURL)

	var wg sync.WaitGroup

	client := &http.Client{}

	result := resultStruct{urlsWithCount: make(map[string]int, len(urlList))}
	for _, url := range urlList {
		wg.Add(1)
		go func(targetURL string, substringRegExp *regexp.Regexp, timeout int) {
			defer wg.Done()
			defer result.locker.Unlock()
			defer func(ch <-chan struct{}) {
				<-ch
			}(limiter)
			limiter <- struct{}{}

			result.locker.Lock()
			result.urlsWithCount[targetURL] = scrapCount(client, targetURL, substringRegExp, timeout)

		}(url, findSubstringRegExp, timeout)
	}
	wg.Wait()

	render(result.urlsWithCount)
}
