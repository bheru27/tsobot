package main

import (
	"encoding/xml"
	"math/rand"
	"net/http"
	"sync"
	"time"
)

type (
	rssFeed struct {
		Description string    `xml:"channel>description"`
		Link        string    `xml:"channel>link"`
		Items       []rssItem `xml:"channel>item"`
	}

	rssItem struct {
		Title       string   `xml:"title"`
		Description string   `xml:"description"`
		Link        string   `xml:"link"`
		Category    []string `xml:"category"`
	}
)

var (
	rssLastFetch   time.Time
	rssLastFetchMu sync.Mutex

	lobstersCache rssFeed
)

func lobsters() string {
	rssLastFetchMu.Lock()
	defer rssLastFetchMu.Unlock()
	if time.Since(rssLastFetch) > time.Hour {
		resp, err := http.Get("https://lobste.rs/rss")
		checkErr(err)
		if resp.StatusCode != 200 {
			printResponse(resp)
			return resp.Status
		}
		rssLastFetch = time.Now()

		checkErr(xml.NewDecoder(resp.Body).Decode(&lobstersCache))
	}

	return lobstersCache.Items[rand.Intn(len(lobstersCache.Items))].Link
}
