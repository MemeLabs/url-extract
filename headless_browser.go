package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/url"
	"time"

	"github.com/chromedp/cdproto/network"
	"github.com/chromedp/chromedp"
)

type HeadlessBrowser struct {
	Info  *InstanceInfo
	Quiet bool
}

// NewHeadlessBrowser connects to a headless browser at remote.
// If quiet is true, debug output is suppressed.
func NewHeadlessBrowser(remote string, quiet bool) (*HeadlessBrowser, error) {
	ii, err := GetInstanceInfo(remote)
	if err != nil {
		return nil, fmt.Errorf("Unable to connect to instance: %q", err)
	}
	return &HeadlessBrowser{
		Info:  ii,
		Quiet: quiet,
	}, nil
}

// ExtractURL visits the given targetURL until it finds a new url that is accepted by matcherFunc or timeout expires.
func (hb *HeadlessBrowser) ExtractURL(targetURL string, timeout time.Duration, matcherFunc func(url *url.URL) bool) (*url.URL, error) {

	matchChan := make(chan *url.URL, 1)
	ticker := time.NewTicker(timeout)

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
				log.Printf("REQUEST: %q - ERR: %q", ev.Request.URL, err)
			}
			if matcherFunc(url) {
				log.Printf("MATCH: %q", ev.Request.URL)
				matchChan <- url
			}

		}
	})

	chromedp.Run(ctx, network.Enable())
	if err := chromedp.Run(ctx,
		chromedp.Navigate(targetURL),
	); err != nil {
		return nil, err
	}

	select {
	case m := <-matchChan:
		return m, nil
	case <-ticker.C:
		chromedp.Stop()
		return nil, errors.New("timeout")
	}
}
