package main

import (
	"crypto/tls"
	"flag"
	"fmt"
	"log"
	"math/rand"
	"os"
	"os/signal"
	"strings"
	"time"

	"github.com/fluffle/goirc/client"
	"github.com/fluffle/goirc/logging"
)

var host string
var port int
var ssl bool
var nick string
var pass string
var join string
var u string
var p string

type tsoLogger struct{}

func (l *tsoLogger) Debug(f string, a ...interface{}) { log.Printf(f+"\n", a...) }
func (l *tsoLogger) Info(f string, a ...interface{})  { log.Printf(f+"\n", a...) }
func (l *tsoLogger) Warn(f string, a ...interface{})  { log.Printf(f+"\n", a...) }
func (l *tsoLogger) Error(f string, a ...interface{}) { log.Printf(f+"\n", a...) }

func checkErr(err error) {
	if err != nil {
		log.Fatalln(err)
	}
}

var sendMessage func(who, msg string)

func main() {
	flag.StringVar(&host, "host", "irc.rizon.net", "host")
	flag.IntVar(&port, "port", 6697, "port")
	flag.BoolVar(&ssl, "ssl", true, "use ssl?")

	flag.StringVar(&nick, "nick", "tsobot", "nick")
	flag.StringVar(&pass, "pass", "", "NickServ IDENTIFY password (optional)")
	flag.StringVar(&join, "join", "tso", "join these channels (space separated list)")

	flag.StringVar(&u, "wuname", "", "watson username")
	flag.StringVar(&p, "wpword", "", "watson password")

	flag.Parse()

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, os.Interrupt)
	signal.Notify(sig, os.Kill)

	l := &tsoLogger{}
	logging.SetLogger(l)

	//irc := client.SimpleClient(nick)
	cfg := client.NewConfig(nick)
	if ssl {
		cfg.SSL = true
		cfg.SSLConfig = &tls.Config{ServerName: host}
		cfg.NewNick = func(n string) string { return n + "^" }
	}
	cfg.Server = fmt.Sprintf("%s:%d", host, port)
	irc := client.Client(cfg)

	irc.HandleFunc(client.CONNECTED, func(c *client.Conn, l *client.Line) {
		if pass != "" {
			irc.Privmsg("NickServ", "IDENTIFY "+pass)
		}
		for _, ch := range strings.Split(join, " ") {
			irc.Join("#" + ch)
		}
	})

	ded := make(chan struct{})
	irc.HandleFunc(client.DISCONNECTED, func(c *client.Conn, l *client.Line) {
		close(ded)
	})

	sendMessage = func(who, msg string) {
		irc.Privmsg(who, msg)
	}

	irc.HandleFunc(client.PRIVMSG, func(c *client.Conn, l *client.Line) {
		//log.Printf("%#v\n", l)
		who, msg := l.Args[0], l.Args[1]
		if who == nick {
			who = l.Nick
		}
		if msg == ".bots" || msg == "who is tsobot" {
			irc.Privmsg(who, "Reporting in! \x0310go\x0f get github.com/generaltso/tsobot")
			return
		}
		if l.Nick == "tso" && msg == ".test" {
			irc.Quit("\"take off every `zig`\"")
			return
		}
		if cmdRegexp.MatchString(msg) {
			matches := cmdRegexp.FindAllStringSubmatch(msg, -1)
			if len(matches) == 0 {
				return
			}
			for _, m := range matches {
				var new string
				if e, ok := emoji[m[1]]; ok {
					new = e
				} else if o, ok := other[m[1]]; ok {
					new = o[rand.Intn(len(o))]
				} else if j, ok := jmote[m[1]]; ok {
					new = j[rand.Intn(len(j))]
				} else {
					return
				}
				msg = strings.Replace(msg, m[0], new, 1)
			}
			irc.Privmsg(who, msg)
			return
		}
		if strings.Index(msg, ".tone_police") == 0 {
			if msg == ".tone_police" {
				irc.Privmsg(who, "(feed me data)")
				return
			}
			text := strings.Replace(msg, ".tone_police", "", -1)
			lines := tonePolice([]byte(`{"text":"` + text + `"}`))
			irc.Privmsg(who, strings.Join(lines, " | "))
			return
		}
		if l.Nick == "tso" && strings.Index(msg, ".rss") == 0 {
			if msg == ".rss" {
				irc.Privmsg(who, "(enter rss url pls)")
				return
			}
			badidea := strings.Replace(msg, ".rss ", "", -1)
			err := feed.Fetch(badidea, nil)
			if err != nil {
				irc.Privmsg(who, err.Error())
			}
			return
		}
		if strings.Index(msg, ".trans") == 0 {
			text := strings.Replace(msg, ".trans ", "", -1)
			text = strings.Replace(text, "/", "", -1)
			irc.Privmsg(who, translate(text))
		}
	})

	if err := irc.ConnectTo(host); err != nil {
		log.Fatalln("Connection error:", err)
	}

	select {
	case <-sig:
		log.Println("we get signal")
		for _, ch := range strings.Split(join, " ") {
			irc.Part("#"+ch, "we get signal")
		}
		<-time.After(time.Second)
		irc.Quit()
		os.Exit(0)
	case <-ded:
		log.Println("disconnected.")
		os.Exit(1)
	}
}
