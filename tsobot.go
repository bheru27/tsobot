package main

import (
	"crypto/tls"
	"flag"
	"log"
	"os"
	"os/signal"
	"regexp"

	"github.com/fluffle/goirc/client"
)

var host string
var nick string
var ch string

func main() {
	flag.StringVar(&host, "host", "irc.rizon.net", "host")
	flag.StringVar(&nick, "nick", "tsobot", "nick")
	flag.StringVar(&ch, "chan", "#tso", "chan")

	flag.Parse()

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, os.Interrupt)

	//irc := client.SimpleClient(nick)
	cfg := client.NewConfig(nick)
	cfg.SSL = true
	cfg.SSLConfig = &tls.Config{ServerName: host}
	cfg.Server = host + ":6697"
	cfg.NewNick = func(n string) string { return n + "^" }
	irc := client.Client(cfg)

	ded := make(chan struct{})
	irc.HandleFunc(client.CONNECTED, func(c *client.Conn, l *client.Line) {
		irc.Join(ch)
	})
	irc.HandleFunc(client.DISCONNECTED, func(c *client.Conn, l *client.Line) {
		close(ded)
	})
	cmdRegexp := regexp.MustCompile(`:(\w+):`)

	irc.HandleFunc(client.PRIVMSG, func(c *client.Conn, l *client.Line) {
		//log.Printf("%#v\n", l)
		who, msg := l.Args[0], l.Args[1]
		if msg == ".bots" {
			irc.Privmsg(who, "Reporting in! [Go]")
			return
		}
		if cmdRegexp.MatchString(msg) {
			m := cmdRegexp.FindStringSubmatch(msg)
			if e, ok := emoji[m[1]]; ok {
				irc.Privmsg(who, e)
			}
		}
	})

	if err := irc.ConnectTo(host); err != nil {
		log.Fatalln("Connection error:", err)
	}

	select {
	case <-sig:
		log.Println("we get signal")
		irc.Quit("we get signal")
		os.Exit(0)
	case <-ded:
		log.Println("disconnected.")
		os.Exit(1)
	}
}
