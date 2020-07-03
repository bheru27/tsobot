package main

import (
	"encoding/xml"
	"io"
	"math/rand"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/dayvonjersen/tsobot/strip"
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
		PubDate     string   `xml:"pubDate"`
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

var (
	ngateLastFetch   time.Time
	ngateLastPubDate time.Time
	ngateLastFetchMu sync.Mutex

	ngateCache  rssFeed
	ngateQuotes = []string{}
)

func init() {
	var err error
	ngateLastPubDate, err = time.Parse("Mon, _2 Jan 2006 15:04:05 MST", "Sun, 14 Apr 2019 22:19:54 PDT")
	checkErr(err)
}

func ngate() (string, bool) {
	ngateLastFetchMu.Lock()
	defer ngateLastFetchMu.Unlock()

	fresh := false
	if time.Since(ngateLastFetch) > time.Hour {
		resp, err := http.Get("http://n-gate.com/index.rss")
		checkErr(err)
		if resp.StatusCode != 200 {
			printResponse(resp)
			return resp.Status, false
		}
		ngateLastFetch = time.Now()

		checkErr(xml.NewDecoder(resp.Body).Decode(&ngateCache))

		for _, item := range ngateCache.Items {
			pubDate, err := time.Parse("Mon, _2 Jan 2006 15:04:05 MST", item.PubDate)
			checkErr(err)

			if ngateLastPubDate.Before(pubDate) {
				ngateLastPubDate = pubDate
				fresh = true
			} else if len(ngateQuotes) > 0 {
				break
			}

			lines := strings.Split(item.Description, "</p> <p>")
			if len(lines) < 2 {
				f, err := os.Create("ngate-error-parsing-lines-from-paragraph-tags.log")
				checkErr(err)
				io.WriteString(f, item.Description)
				f.Close()
				return ".tell tso parser routine is broken", fresh
			}
			lines = lines[1:]
			for _, ln := range lines {
				stuff := strings.Split(ln, "<br>")
				if len(stuff) < 3 {
					f, err := os.Create("ngate-error-parsing-words-from-linebreak-tags.log")
					checkErr(err)
					io.WriteString(f, item.Description)
					f.Close()
					return ".tell tso parser routine is broken", fresh
				}

				text := strip.StripTags(strings.TrimSpace(stuff[2]))

				for _, sentence := range strings.Split(text, ". ") {
					sentence = strings.TrimSpace(sentence)
					if sentence == "" {
						continue
					}
					if len(sentence) > 300 {
						words := strings.Split(sentence, " ")
						sentence = ""
						for _, word := range words {
							if len(word)+len(sentence)+1 < 297 {
								sentence = sentence + " " + word
							} else {
								break
							}
						}
						sentence = sentence + "..."
					}

					ngateQuotes = append(ngateQuotes, sentence)
				}
			}
		}
	}

	return ngateQuotes[rand.Intn(len(ngateQuotes))], fresh
}
