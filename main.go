package main

import (
	"flag"
	"log"
	"net/url"
	"strings"
	"time"
)

var (
	targetURL string
	remote    string
	quiet     bool
)

func init() {
	flag.StringVar(&targetURL, "url", "", "the URL to analyze")
	flag.StringVar(&remote, "remote", "localhost:9222", "the endpoint of the headless instance")
	flag.BoolVar(&quiet, "quiet", false, "discard debug output")
	flag.Parse()
}

func main() {

	if targetURL == "" {
		log.Fatalf("Please provide a URL with '-url'.")
	}

	hb, err := NewHeadlessBrowser(remote, quiet)
	if err != nil {
		log.Fatal(err)
	}

	log.Printf("Success! Found instace %q with User-Agent %q. Using debuggerURL %q.",
		hb.Info.Browser,
		hb.Info.UserAgent,
		hb.Info.WebSocketDebuggerURL,
	)

	matcherFunc := func(url *url.URL) bool {
		return strings.HasSuffix(url.Path, ".m3u8")
	}

	result, err := hb.ExtractURL(targetURL, time.Second*5, matcherFunc)
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("got result! %q", result)
}
