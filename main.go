package main

import "flag"
import "fmt"
//import "errors"
import "os"
import "strings"
import "net/http"
import "io/ioutil"
import "regexp"

func main() {
	result := make(map[string]int)
	urls := flag.String("urls", "", "a string")
	searchUrl := flag.String("search", "", "a string")

	flag.Parse()

	if *urls == "" {
		fmt.Println("urlList are empty")
		os.Exit(1)
	}

	if *searchUrl == "" {
		fmt.Println("search url is empty")
		os.Exit(1)
	}

	var urlList []string
    urlList = strings.Split(*urls, ",")
    for _, url := range urlList {
        result[url] = 0
        resp, err := http.Get(url)
        if err != nil {
			panic(err)
		}

		defer resp.Body.Close()
		html, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			panic(err)
		}

		findSubstringRegExp := regexp.MustCompile(*searchUrl)
		matches := findSubstringRegExp.FindAllStringIndex(string(html), -1)

		result[url] = len(matches)
    }

    for key, value := range result {
    	fmt.Println(key, " - ", value)
    }	
    
}