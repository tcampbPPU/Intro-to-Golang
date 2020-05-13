package main

import (
	"crypto/tls"
	"fmt"
	"net/http"
	"net/url"
	"os"

	"github.com/steelx/extractlinks"
)

// GLOBALS
var (
	config = &tls.Config{
		InsecureSkipVerify: true,
	}

	transport = &http.Transport{
		TLSClientConfig: config,
	}

	netClient = &http.Client{
		Transport: transport,
	}

	queue = make(chan string)

	hasVisted = make(map[string]bool)
)

func main() {
	// get url from command line arg
	args := os.Args[1:]

	fmt.Println(args)

	if len(args) == 0 {
		fmt.Println("No URL passed.")
		os.Exit(1)
	}

	// build concurrent queue for the URL found on the given paage
	baseURL := args[0]
	go func() {
		queue <- baseURL
	}()

	for href := range queue {
		// start crawling the URL in the queue
		if !hasVisted[href] && isSameDomain(href, baseURL) {
			crawlURL(href)
		}
	}

}

/*
 @desc - crawls the current URL in the queue
*/
func crawlURL(href string) {
	hasVisted[href] = true
	fmt.Printf("Crawling URL -> %v \n", href)
	response, err := netClient.Get(href)
	checkError(err)
	defer response.Body.Close()

	// array of links
	links, _ := extractlinks.All(response.Body)

	for _, link := range links {

		// need to call async, otherwise links will be added before the queue even gets them... I think
		absoluteURL := toFixedURL(link.Href, href)
		go func() {
			queue <- absoluteURL
		}()

	}
}

/*
 @desc - Checks if the current URL is the same host address as the starting URL
*/
func isSameDomain(href, baseURL string) bool {
	uri, err := url.Parse(href)
	if err != nil {
		return false
	}
	parentURI, err := url.Parse(baseURL)
	if err != nil {
		return false
	}

	if uri.Host != parentURI.Host {
		return false
	}
	return true
}

/*
 @desc - maps the path of the URL to the original base URL
*/
func toFixedURL(href, baseURL string) string {
	uri, err := url.Parse(href)
	if err != nil {
		return ""
	}
	base, err := url.Parse(baseURL)
	if err != nil {
		return ""
	}

	toFixedURI := base.ResolveReference(uri)
	// host from base and path from uri

	return toFixedURI.String()

}

/*
 @desc - Handle errors
*/
func checkError(err error) {
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
