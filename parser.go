package main

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"regexp"
	"strings"
	"sync"
)

func parse(urls string, searchUrl string) {
	result := make(map[string]int)
	var urlList []string
	urlList = strings.Split(urls, ",")

	findSubstringRegExp := regexp.MustCompile(searchUrl)

	var wg sync.WaitGroup
	for _, url := range urlList {
		wg.Add(1)
		go func(urlA string, substringRegExp *regexp.Regexp) {
			defer wg.Done()

			resp, err := http.Get(urlA)
			if err != nil {
				fmt.Println(err)
				result[urlA] = 0
			}
			defer resp.Body.Close()
			html, err := ioutil.ReadAll(resp.Body)
			if err != nil {
				fmt.Println(err)
				result[urlA] = 0
			}

			if err == nil {
				html := string(html)
				matchesCount := 0
				if html != "" {
					matches := findSubstringRegExp.FindAllStringIndex(html, -1)
					matchesCount = len(matches)
				}

				result[urlA] = matchesCount
			}
		}(url, findSubstringRegExp)
	}

	wg.Wait()

	for key, value := range result {
		fmt.Println(key, " - ", value)
	}
}
