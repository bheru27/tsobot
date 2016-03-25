package main

import rss "github.com/jteeuwen/go-pkg-rss"

type subscription struct {
	who string `irc channel`
	src string `rss channel`
}

type clickbait struct {
	tit string `title`
	url string `click`
	src string `shits`
}

func channelhandler(f *rss.Feed, newchannels []*rss.Channel) {}
func itemhandler(f *rss.Feed, channel *rss.Channel, items []*rss.Item) {
	for i, item := range items {
		if i > 4 {
			break
		}
		if len(item.Links) > 0 {
			noiz <- &clickbait{tit: item.Title, url: item.Links[0].Href, src: f.Url}
		}
	}
}

var feed *rss.Feed
var noiz chan *clickbait
var subs []*subscription
