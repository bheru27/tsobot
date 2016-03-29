package main

import "github.com/fluffle/goirc/client"

type line struct {
	channel, nick, text string
	time                int64
}

var chat chan *line

func logLine(l *client.Line) {
	chat <- &line{l.Args[0], l.Nick, l.Args[1], l.Time.Unix()}
}
