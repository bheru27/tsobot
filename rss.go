package main

import (
	"fmt"
	"time"

	rss "github.com/jteeuwen/go-pkg-rss"
)

var feed *rss.Feed = rss.New(5, false, func(f *rss.Feed, newchannels []*rss.Channel) {
}, func(f *rss.Feed, channel *rss.Channel, items []*rss.Item) {
	for i, item := range items {
		if len(item.Links) > 0 {
			<-time.After(time.Second * 3)
			sendMessage("tso", fmt.Sprintf("[%d] %s : %s", i, item.Links[0].Href, item.Title))
		}
	}
})
