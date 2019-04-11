package main

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"regexp"
	"strings"
)

func makeRequest(url string, ch chan<- string) {
	resp, err := http.Get(url)
	if err != nil {
		fmt.Println(err)
		ch <- ""
	}
	defer resp.Body.Close()
	html, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Println(err)
		ch <- ""
	}
	if err == nil {
		ch <- string(html)
	}
}

func parse(urls string, searchUrl string) {
	result := make(map[string]int)
	var urlList []string
	ch := make(chan string)
	urlList = strings.Split(urls, ",")
	for _, url := range urlList {
		go makeRequest(url, ch)
	}

	findSubstringRegExp := regexp.MustCompile(searchUrl)
	for _, url := range urlList {
		html := <-ch
		matchesCount := 0
		if html != "" {
			matches := findSubstringRegExp.FindAllStringIndex(html, -1)
			matchesCount = len(matches)
		}
		result[url] = matchesCount
	}

	for key, value := range result {
		fmt.Println(key, " - ", value)
	}
}
