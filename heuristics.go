package main

import (
	"context"
	"log"

	"github.com/chromedp/cdproto/cdp"
	"github.com/chromedp/chromedp"
)

var (
	clickSelectors = []string{
		"[class*='play']",
		"[id*='play']",
		"[class*='btn-danger']",
	}
)

// Click everything that "looks like" a play button to catch media that does not auto-play.
// Clicking seems to block until any element is found by the given selector.
// Since we cannot guarantee this to happen, use routines...
func clickAll(ctx context.Context) {
	for _, sel := range clickSelectors {
		go func(sel string) {
			clickAllNodes(ctx, sel)
			log.Printf("done clicking by %q", sel)
		}(sel)
	}
}

// clickAllNodes clicks all nodes that match the given selector.
// errors are ignored and logged.
func clickAllNodes(ctx context.Context, selector string) {

	var nodes []*cdp.Node
	if err := chromedp.Run(ctx, chromedp.Nodes(selector, &nodes)); err != nil {
		log.Printf("error getting nodes: %q", err)
		return
	}
	log.Printf("heuristics: found %d nodes for selector %q", len(nodes), selector)
	for _, n := range nodes {
		if err := chromedp.Run(ctx, chromedp.MouseClickNode(n)); err != nil {
			log.Printf("error clicking node: %q", err)
		}
	}
}
