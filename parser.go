// Package parsercli was built for parsing url for search string
package parsercli

import (
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

//render generate output Url - count
func render(urlsResult map[string]int) {
	for key, value := range urlsResult {
		fmt.Println(key, " - ", value)
	}
}

//scrapCount get html by url and find by regexp count matches
func scrapCount(client *http.Client, targetURL string, substringRegExp *regexp.Regexp) int {
	resp, err := client.Get(targetURL)
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
func Parse(urls, searchURL string, maxGoroutines int) {

	limiter := make(chan struct{}, maxGoroutines)

	urlList := strings.Split(urls, ",")

	findSubstringRegExp := regexp.MustCompile(searchURL)

	var wg sync.WaitGroup

	timeout := 5 * time.Second
	client := &http.Client{
		Timeout: timeout,
	}

	result := resultStruct{urlsWithCount: make(map[string]int, len(urlList))}
	for _, url := range urlList {
		wg.Add(1)
		go func(targetURL string, substringRegExp *regexp.Regexp) {
			defer wg.Done()
			defer result.locker.Unlock()
			defer func(ch <-chan struct{}) {
				<-ch
			}(limiter)
			limiter <- struct{}{}

			result.locker.Lock()
			result.urlsWithCount[targetURL] = scrapCount(client, targetURL, substringRegExp)

		}(url, findSubstringRegExp)
	}
	wg.Wait()

	render(result.urlsWithCount)
}
