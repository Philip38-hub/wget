package mirrorutils

import (
	"context"
	"fmt"
	"time"

	"github.com/chromedp/cdproto/network" //allows for interaction with a headless browser
	"github.com/chromedp/chromedp"
	"wget/models"
)

// getDynamicContent fetches JavaScript-rendered content and resources
func getDynamicContent(url string) (*models.DynamicContent, error) {
	// Create context with timeout
	ctx, cancel := chromedp.NewContext(context.Background())
	defer cancel()

	ctx, cancel = context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	var content models.DynamicContent
	var resources = make(map[string]models.Resource)

	// Listen for network events
	chromedp.ListenTarget(ctx, func(ev interface{}) {
		switch e := ev.(type) {
		case *network.EventResponseReceived:
			go func() {
				// Get response body
				rbp, err := network.GetResponseBody(e.RequestID).Do(ctx)
				if err != nil {
					return
				}
				
				resources[e.Response.URL] = models.Resource{
					URL:         e.Response.URL,
					ContentType: e.Response.MimeType,
					Data:        rbp,
				}
			}()
		}
	})

	// Enable network events
	if err := chromedp.Run(ctx, network.Enable()); err != nil {
		return nil, fmt.Errorf("failed to enable network monitoring: %v", err)
	}

	// Navigate and wait for network idle
	if err := chromedp.Run(ctx,
		chromedp.Navigate(url),
		chromedp.WaitReady("body", chromedp.ByQuery),
		chromedp.Sleep(2*time.Second), // Wait for dynamic content
	); err != nil {
		return nil, fmt.Errorf("failed to navigate: %v", err)
	}

	// Get rendered HTML
	var html string
	if err := chromedp.Run(ctx,
		chromedp.OuterHTML("html", &html, chromedp.ByQuery),
	); err != nil {
		return nil, fmt.Errorf("failed to get HTML: %v", err)
	}

	content.HTML = html
	for _, resource := range resources {
		content.Resources = append(content.Resources, resource)
	}

	return &content, nil
}
