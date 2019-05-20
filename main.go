package main

import (
	"flag"
	"log"
	"net/url"
	"strings"
	"time"

	"github.com/chromedp/cdproto/network"
)

var (
	target      string
	headlessURL string
	heuristics  bool
	timeout     int
	quiet       bool

	mediaMatcher = func(url *url.URL) bool {
		return strings.HasSuffix(url.Path, ".m3u8") ||
			strings.HasSuffix(url.Path, ".mp4") ||
			strings.HasSuffix(url.Path, ".mp3")
	}
)

func init() {
	flag.StringVar(&target, "url", "", "the URL to analyze")
	flag.StringVar(&headlessURL, "remote", "localhost:9222", "the endpoint of the headless instance")
	flag.BoolVar(&heuristics, "heuristics", false, "use heuristics to find media elements")
	flag.IntVar(&timeout, "timeout", 5, "time in seconds to wait for the site to load and a result to be detected")
	flag.BoolVar(&quiet, "quiet", false, "discard debug output")
	flag.Parse()
}

func main() {

	if target == "" {
		log.Fatalf("Please provide a URL with '-url'.")
	}

	hb, err := NewHeadlessBrowser(headlessURL, heuristics, quiet)
	if err != nil {
		log.Fatal(err)
	}

	resultChan := make(chan *network.Request, 100)
	go func() {
		err := hb.ExtractURL(target, time.Second*time.Duration(timeout), resultChan, mediaMatcher)
		if err != nil {
			log.Fatalf("FATAL: %q", err)
		}
	}()

	i := 0
	maxResults := 1
	for i < maxResults {
		result := <-resultChan
		log.Printf("RESULT: %q", result.URL)
		i++
	}
}
