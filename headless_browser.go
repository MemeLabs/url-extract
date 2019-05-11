package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/url"
	"strings"
	"time"

	"github.com/chromedp/cdproto/cdp"
	"github.com/chromedp/cdproto/network"
	"github.com/chromedp/chromedp"
)

type HeadlessBrowser struct {
	Info          *InstanceInfo
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
	return &HeadlessBrowser{
		Info:          ii,
		UseHeuristics: useHeuristics,
		Quiet:         quiet,
	}, nil
}

// ExtractURL visits the given targetURL until it finds a new url that is accepted by matcherFunc or timeout expires.
func (hb *HeadlessBrowser) ExtractURL(targetURL string, timeout time.Duration, matcherFunc func(url *url.URL) bool) (*url.URL, error) {
	log.Printf("Extracting from %q", targetURL)

	// Navigation stalls if the channel is blocked in the network event handler.
	// Buffer as much as is needed...
	matchChan := make(chan *url.URL, 100)
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
				log.Printf("request %q error: %q", ev.Request.URL, err)
			}
			if matcherFunc(url) {
				log.Printf("MATCH: %q", ev.Request.URL)
				matchChan <- url
			}
		}
	})

	if err := chromedp.Run(ctx,
		network.Enable(),             // enable network events
		chromedp.Navigate(targetURL), // navigate to url
	); err != nil {
		return nil, err
	}

	// TODO: respect timeout...
	waitToFinishLoading(ctx)
	log.Println("finished loading!")

	if hb.UseHeuristics {
		// HEURISTICS:
		// Click everything that "looks like" a play button to catch media that does not auto-play.
		// Clicking seems to block until any element is found by the given selector.
		// Since we cannot guarantee this to happen, use routines...
		findstr := "play"
		for _, sel := range []string{
			fmt.Sprintf("[class*='%s']", findstr),
			fmt.Sprintf("[id*='%s']", findstr),
		} {
			go func(sel string) {
				clickAllNodesBySelector(ctx, sel)
				log.Printf("done clicking by %q", sel)
			}(sel)
		}
	}

	log.Println("navigation done")

	for {
		select {
		case m := <-matchChan:
			return m, nil
		case <-ticker.C:
			chromedp.Run(ctx,
				chromedp.Stop(),
			)
			return nil, errors.New("timeout")
		}
	}
}

func clickAllNodesBySelector(ctx context.Context, selector string) {

	var nodes []*cdp.Node
	if err := chromedp.Run(ctx, chromedp.Nodes(selector, &nodes)); err != nil {
		log.Printf("error getting nodes: %q", err)
		return
	}
	log.Printf("heuristics: found %d nodes for %q", len(nodes), selector)
	for _, n := range nodes {
		// log.Println("clicking", n.NodeName, n.Attributes)
		if err := chromedp.Run(ctx, chromedp.MouseClickNode(n)); err != nil {
			log.Printf("error clicking node: %q", err)
		}
	}
}

// waitToFinishLoading waits for site to finish loading (since clicking buttons mights not work correctly otherwise)
// source: https://github.com/chromedp/chromedp/issues/252
func waitToFinishLoading(ctx context.Context) {
	state := "notloaded"
	script := `document.readyState`
	for {
		err := chromedp.Run(ctx, chromedp.EvaluateAsDevTools(script, &state))
		if err != nil {
			log.Println(err)
		}
		if strings.Compare(state, "complete") == 0 {
			break
		}
		time.Sleep(time.Millisecond * 100)
	}
}
