// Package parsercli was built for parsing url for search string
package parsercli

import (
	"context"
	"fmt"
	"strings"
	"sync"
)

type resultStruct struct {
	urlsWithCount map[string]int
	mx            sync.RWMutex
}

//addCountToUrl add count of occurrences to targetURL
func (r *resultStruct) addCountToURL(targetURL string, count int) {
	r.mx.Lock()
	r.urlsWithCount[targetURL] = count
	r.mx.Unlock()
}

type contextKey string

var (
	urlsSize = contextKey("urlsSize")
)

func getUrls(ctx context.Context, urls string) (chan string, context.Context) {
	var wg sync.WaitGroup

	urlList := strings.Split(urls, ",")
	ctx = context.WithValue(ctx, urlsSize, len(urlList))
	urlsChan := make(chan string, len(urlList))

	defer close(urlsChan)
	for _, url := range urlList {
		wg.Add(1)
		go func(ctx context.Context, urlsChan chan string, url string) {
			defer wg.Done()
			fmt.Println(url)
			urlsChan <- url
		}(ctx, urlsChan, url)
	}
	wg.Wait()
	return urlsChan, ctx
}

func getHTML(ctx context.Context, urlsChan chan string) chan string {
	var wg sync.WaitGroup

	urlsSize, _ := ctx.Value(urlsSize).(int)

	htmlChan := make(chan string, urlsSize)
	defer close(htmlChan)

	for url, ok := <-urlsChan; ok; url, ok = <-urlsChan {
		wg.Add(1)
		go func(htmlChan chan string, url string) {
			defer wg.Done()
			fmt.Println("get html", url)
			htmlChan <- url
		}(htmlChan, url)
	}
	wg.Wait()
	return htmlChan
}

func parseHTML(ctx context.Context, htmlChan chan string, searchString string) chan string {
	var wg sync.WaitGroup
	urlsSize, _ := ctx.Value(urlsSize).(int)
	occurrencesChan := make(chan string, urlsSize)

	for url, ok := <-htmlChan; ok; url, ok = <-htmlChan {
		wg.Add(1)
		go func(htmlChan chan string, url string) {
			defer wg.Done()
			fmt.Println("get html", url)
			occurrencesChan <- url
		}(occurrencesChan, url)
	}

	wg.Wait()
	close(occurrencesChan)
	return occurrencesChan
}

//render generate output Url - count
func render(ctx context.Context, occurrencesChan chan string) {
	fmt.Println("render")
	go func(occurrencesChan chan string) {
		for url := range occurrencesChan {
			fmt.Println("render html", url)
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
	urlsChan, ctx := getUrls(ctx, urls)
	htmlChan := getHTML(ctx, urlsChan)
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
