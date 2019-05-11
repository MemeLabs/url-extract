package main

import (
	"flag"
	"log"
	"net/url"
	"strings"
	"time"
)

var (
	targetURL  string
	remote     string
	heuristics bool
	quiet      bool
)

func init() {
	flag.StringVar(&targetURL, "url", "", "the URL to analyze")
	flag.StringVar(&remote, "remote", "localhost:9222", "the endpoint of the headless instance")
	flag.BoolVar(&heuristics, "heuristics", false, "use heuristics to find media elements")
	flag.BoolVar(&quiet, "quiet", false, "discard debug output")
	flag.Parse()
}

func main() {

	if targetURL == "" {
		log.Fatalf("Please provide a URL with '-url'.")
	}

	hb, err := NewHeadlessBrowser(remote, heuristics, quiet)
	if err != nil {
		log.Fatal(err)
	}

	log.Printf("Found instace %q with User-Agent %q. Using debuggerURL %q.",
		hb.Info.Browser,
		hb.Info.UserAgent,
		hb.Info.WebSocketDebuggerURL,
	)

	matcherFunc := func(url *url.URL) bool {
		return strings.HasSuffix(url.Path, ".m3u8") || strings.HasSuffix(url.Path, ".mp4")
	}

	result, err := hb.ExtractURL(targetURL, time.Second*10, matcherFunc)
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("got result! %q", result)
}
