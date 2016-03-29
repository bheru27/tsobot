package main

import (
	"crypto/md5"
	"database/sql"
	"fmt"
	"log"
	"time"

	rss "github.com/jteeuwen/go-pkg-rss"
	_ "github.com/mattn/go-sqlite3"
)

type subscription struct {
	who string `irc channel`
	src string `rss channel`
}

type clickbait struct {
	tit string `title`
	url string `click`
	src string `shits`
}

var cache chan *clickbait
var noiz chan *clickbait
var subs []*subscription

func hashFn(input string) string {
	return fmt.Sprintf("%x", md5.Sum([]byte(input)))
}

var getLines func(string, string) []byte

func cacheHandler() {
	/*
		CREATE TABLE `clickbait` (
		`id` INTEGER PRIMARY KEY AUTOINCREMENT,
		`hash` TEXT NOT NULL,
		`url` TEXT NOT NULL,
		`title` TEXT NOT NULL,
		`createdat` INTEGER NOT NULL
		)
	*/
	db, err := sql.Open("sqlite3", "./.cache/tsobot.db")
	checkErr(err)
	defer db.Close()

	getLines = func(field, value string) []byte {
		stmt, err := db.Prepare("SELECT `line` FROM `log` WHERE " + field + " = ?")
		checkErr(err)
		ret := ""
		rows, err := stmt.Query(value)
		checkErr(err)
		for rows.Next() {
			var ln string
			rows.Scan(&ln)
			ret += ln + "\n"
		}
		return []byte(ret)
	}

	ins, err := db.Prepare("INSERT INTO `clickbait` (`hash`, `url`, `title`, `createdat`) VALUES (?, ?, ?, ?)")
	checkErr(err)
	defer ins.Close()
	ins2, err := db.Prepare("INSERT INTO `log` (`chan`, `nick`, `line`, `time`) VALUES (?, ?, ?, ?)")
	checkErr(err)
	defer ins2.Close()

	cnt, err := db.Prepare("SELECT COUNT(*) FROM `clickbait` WHERE `hash` = ?")
	checkErr(err)
	defer cnt.Close()

	for {
		select {
		case bait := <-cache:
			hash := hashFn(bait.url)
			row := cnt.QueryRow(hash)
			var count int
			row.Scan(&count)
			if count == 0 {
				now := time.Now().Unix()
				log.Printf("%#v\n%#v\n%#v\n%#v\n", hash, bait.url, bait.tit, now)
				_, err := ins.Exec(hash, bait.url, bait.tit, now)
				checkErr(err)
				noiz <- bait
			}
		case line := <-chat:
			_, err := ins2.Exec(line.channel, line.nick, line.text, line.time)
			checkErr(err)
		}
	}
}

func channelhandler(f *rss.Feed, newchannels []*rss.Channel) {}
func itemhandler(f *rss.Feed, channel *rss.Channel, items []*rss.Item) {
	for _, item := range items {
		if len(item.Links) > 0 {
			cache <- &clickbait{tit: item.Title, url: item.Links[0].Href, src: f.Url}
		}
	}
}

func pollFeed(url string) {
	feed := rss.New(5, false, channelhandler, itemhandler)
	for {
		checkErr(feed.Fetch(url, nil))
		<-time.After(time.Duration(feed.SecondsTillUpdate() * 1e9))
	}
}
