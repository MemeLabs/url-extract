package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/url"
	"strings"
	"time"

	"github.com/chromedp/cdproto/network"
	"github.com/chromedp/chromedp"
)

type HeadlessBrowser struct {
	Info          *InstanceInfo
	stopChan      chan bool
	UseHeuristics bool
	Quiet         bool
}

// NewHeadlessBrowser connects to a headless browser at remote.
// If quiet is true, debug output is suppressed.
func NewHeadlessBrowser(remote string, useHeuristics bool, quiet bool) (*HeadlessBrowser, error) {
	ii, err := GetInstanceInfo(remote)
	if err != nil {
		return nil, fmt.Errorf("Unable to connect to instance: %q", err)
	}

	log.Printf("Found instace %q with User-Agent %q. Using debuggerURL %q.",
		ii.Browser,
		ii.UserAgent,
		ii.WebSocketDebuggerURL,
	)

	return &HeadlessBrowser{
		Info:          ii,
		stopChan:      make(chan bool, 1),
		UseHeuristics: useHeuristics,
		Quiet:         quiet,
	}, nil
}

// ExtractURL visits the given targetURL until it finds a new url that is accepted by matcherFunc or timeout expires.
func (hb *HeadlessBrowser) ExtractURL(targetURL string, timeout time.Duration, resultChan chan *network.Request, matcherFunc func(url *url.URL) bool) error {
	log.Printf("extracting from %q", targetURL)

	timeoutTicker := time.NewTicker(timeout)

	// source: https://github.com/chromedp/chromedp/blob/master/allocate_test.go
	allocCtx, allocCancel := chromedp.NewRemoteAllocator(context.Background(), hb.Info.WebSocketDebuggerURL)
	defer allocCancel()
	ctx, cancel := chromedp.NewContext(allocCtx)
	defer cancel()

	chromedp.ListenTarget(ctx, func(ev interface{}) {
		switch ev := ev.(type) {

		case *network.EventWebSocketCreated:
			if hb.Quiet {
				break
			}
			log.Printf("WEBSOCKET: %q", ev.URL)

		case *network.EventLoadingFailed:
			if hb.Quiet {
				break
			}
			log.Printf("FAILED: %q", ev.ErrorText)

		case *network.EventRequestWillBeSent:
			if !hb.Quiet {
				log.Printf("REQUEST: %q", ev.Request.URL)
			}

			url, err := url.Parse(ev.Request.URL)
			if err != nil {
				log.Printf("request %q error: %q", ev.Request.URL, err)
			}
			if ev.Request.URL != targetURL && matcherFunc(url) {
				// Navigation stalls if channel is blocked...
				go func() { resultChan <- ev.Request }()
			}
		}
	})

	if err := chromedp.Run(ctx,
		network.Enable(),             // enable network events
		chromedp.Navigate(targetURL), // navigate to url
	); err != nil {
		return err
	}

	log.Println("waiting for page to finish loading...")
	err := waitToFinishLoading(ctx, timeoutTicker)
	if err != nil {
		return err
	}

	if hb.UseHeuristics {
		clickAll(ctx)
	}

	log.Printf("waiting to find matching urls...")

	select {
	case <-timeoutTicker.C:
		chromedp.Run(ctx,
			chromedp.Stop(),
		)
		return errors.New("timeout")
	case <-hb.stopChan:
		chromedp.Run(ctx,
			chromedp.Stop(),
		)
		log.Println("stopped!")
		return nil
	}
}

// Stop trys to abort ExtractURL and shuts down the headless browser instance.
func (hb *HeadlessBrowser) Stop() {
	hb.stopChan <- true
}

// waitToFinishLoading waits for site to finish loading (since clicking buttons mights not work correctly otherwise)
// source: https://github.com/chromedp/chromedp/issues/252
// Only returns with an error on timeout.
func waitToFinishLoading(ctx context.Context, timeoutTicker *time.Ticker) error {
	state := "notloaded"
	script := `document.readyState`
	checkTicker := time.NewTicker(time.Millisecond * 100)
	for {
		select {
		case <-checkTicker.C:
			err := chromedp.Run(ctx, chromedp.EvaluateAsDevTools(script, &state))
			if err != nil {
				log.Printf("error in eval: %q", err)
			}
			if strings.Compare(state, "complete") == 0 {
				return nil
			}
		case <-timeoutTicker.C:
			return errors.New("timeout while waiting to finish loading")
		}
	}
}
