package main

import (
	"encoding/json"
	"math/rand"
	"regexp"
	"strings"
	"sync"
)

var (
	quotes   = map[string]map[string][]string{}
	quotesMu sync.Mutex

	nobracketsRe = regexp.MustCompile(`[^a-z0-9-_]`)
)

func loadQuotes(filename string) {
	quotesMu.Lock()
	defer quotesMu.Unlock()
	if fileExists(filename) {
		data := fileGetContents(filename)
		checkErr(json.Unmarshal(data, &quotes))
	}
}

func saveQuotes(filename string) {
	quotesMu.Lock()
	defer quotesMu.Unlock()
	data, err := json.Marshal(quotes)
	checkErr(err)
	filePutContents(filename, data)
}

func addQuote(channel, arg, nick string) string {
	args := strings.SplitN(strings.TrimSpace(arg), " ", 1)
	src := args[0]
	src = strings.TrimSpace(nobracketsRe.ReplaceAllString(strings.ToLower(src), ""))
	if src == "" {
		return "usage: .addquote [nick] [message (optional)]"
	}

	q := ""

	if src == strings.ToLower(nick) {
		return "don't quote yourself you narcissistic twat"
	}

	if len(args) > 1 {
		q = args[1]
	} else {
		chatlogMu.Lock()
		defer chatlogMu.Unlock()

		if logs, ok := chatlogs[channel]; ok {
			for _, ln := range logs {
				if ln == nil {
					break
				}
				if ln[0] == src {
					q = ln[1]
					goto here
				}
			}
			return "(no quotes from " + src + " in chatlog)"
		}
		return "(no quotes from " + channel + " in chatlog)"
	}
here:

	quotesMu.Lock()
	defer quotesMu.Unlock()

	// I LOVE MAPS SO MUCH
	_, ok := quotes[channel]
	if !ok {
		quotes[channel] = map[string][]string{}
	}

	// I LOVE MAPS SO MUCH
	if _, ok := quotes[channel][src]; !ok {
		quotes[channel][src] = []string{}
	}

	quotes[channel][src] = append(quotes[channel][src], q)
	// FUCK

	return "oic 📝"
}

func getQuote(channel, src string) string {
	quotesMu.Lock()
	defer quotesMu.Unlock()
	if _, ok := quotes[channel]; ok {
		if q, ok := quotes[channel][src]; ok {
			return "<" + src + "> " + q[rand.Intn(len(q))]
		}
	}
	return "i got nothing, sorry"

}

func getRandQuote(channel string) string {
	quotesMu.Lock()
	defer quotesMu.Unlock()
	if _, ok := quotes[channel]; ok {
		i := 0
		r := rand.Intn(len(quotes[channel]))
		for src, q := range quotes[channel] {
			if i == r {
				return "<" + src + "> " + q[rand.Intn(len(q))]
			}
			i++
		}
		return "Something happened. :("
	}
	return "no really, i got nothing"
}
