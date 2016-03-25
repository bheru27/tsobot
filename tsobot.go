package main

import (
	"crypto/tls"
	"flag"
	"fmt"
	"log"
	"math/rand"
	"os"
	"os/signal"
	"regexp"
	"sort"
	"strings"
	"time"

	"github.com/fluffle/goirc/client"
	"github.com/fluffle/goirc/logging"
)
import rss "github.com/jteeuwen/go-pkg-rss"

/**
 * Configuration variables, passed in with command line flags
 */
var host string
var port int
var ssl bool
var nick string
var pass string
var join string
var u string
var p string
var admin string
var cache_dir string

/**
 * Arbitrary way GoIRC handles logging
 */
type tsoLogger struct{}

func (l *tsoLogger) Debug(f string, a ...interface{}) { log.Printf(f+"\n", a...) }
func (l *tsoLogger) Info(f string, a ...interface{})  { log.Printf(f+"\n", a...) }
func (l *tsoLogger) Warn(f string, a ...interface{})  { log.Printf(f+"\n", a...) }
func (l *tsoLogger) Error(f string, a ...interface{}) { log.Printf(f+"\n", a...) }

/**
 * More boilerplate
 */
func checkErr(err error) {
	if err != nil {
		log.Fatalln(err)
	}
}

/**
 * Botnet
 */
var botAdmins sort.StringSlice
var botCommandRe *regexp.Regexp = regexp.MustCompile(`^\.(\w+)\s*(.*)$`)

type botCommand struct {
	admin bool
	fn    func(who, arg, nick string)
}

var botCommands map[string]*botCommand
var sendMessage func(who, msg string)

func parseMessage(who, msg, nick string) {
	if !botCommandRe.MatchString(msg) {
		return
	}

	m := botCommandRe.FindStringSubmatch(msg)
	cmd := m[1]
	arg := m[2]

	if b, ok := botCommands[cmd]; ok {
		if !b.admin || (b.admin && isAdmin(nick)) {
			b.fn(who, arg, nick)
		} else {
			//log.Printf("%#v\n", botAdmins)
			sendMessage(nick, "Access denied.")
		}
	}
}

func isAdmin(nick string) bool {
	ind := sort.SearchStrings(botAdmins, nick)
	return botAdmins[ind] == nick
}

func addAdmin(nick string) {
	botAdmins = append(botAdmins, nick)
	botAdmins = sort.StringSlice(botAdmins)
}

func removeAdmin(nick string) {
	ind := sort.SearchStrings(botAdmins, nick)
	if botAdmins[ind] == nick {
		botAdmins = append(botAdmins[:ind], botAdmins[ind+1:]...)
		botAdmins = sort.StringSlice(botAdmins)
	}
}

func main() {
	flag.StringVar(&host, "host", "irc.rizon.net", "host")
	flag.IntVar(&port, "port", 6697, "port")
	flag.BoolVar(&ssl, "ssl", true, "use ssl?")

	flag.StringVar(&nick, "nick", "tsobot", "nick")
	flag.StringVar(&pass, "pass", "", "NickServ IDENTIFY password (optional)")
	flag.StringVar(&join, "join", "tso", "space separated list of channels to join")

	flag.StringVar(&u, "wuname", "", "watson username")
	flag.StringVar(&p, "wpword", "", "watson password")

	flag.StringVar(&admin, "admin", "tso", "space separated list of privileged nicks")
	flag.StringVar(&cache_dir, "cache_dir", ".cache", "directory to cache datas like rss feeds")

	flag.Parse()

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, os.Interrupt)
	signal.Notify(sig, os.Kill)

	logging.SetLogger(&tsoLogger{})

	cfg := client.NewConfig(nick)
	if ssl {
		cfg.SSL = true
		cfg.SSLConfig = &tls.Config{ServerName: host, InsecureSkipVerify: true}
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
		botAdmins = sort.StringSlice(strings.Split(admin, " "))
	})

	ded := make(chan struct{})
	irc.HandleFunc(client.DISCONNECTED, func(c *client.Conn, l *client.Line) {
		close(ded)
	})

	sendMessage = func(who, msg string) {
		irc.Privmsg(who, msg)
	}

	feed = rss.New(5, false, channelhandler, itemhandler)
	noiz = make(chan *clickbait)

	botCommands = map[string]*botCommand{
		"bots": &botCommand{false, func(who, arg, nick string) {
			sendMessage(who, "Reporting in! "+colorString("go", White, Black)+" get github.com/generaltso/tsobot")
		}},
		"add_admin": &botCommand{true, func(who, arg, nick string) {
			for _, adm := range strings.Split(arg, " ") {
				irc.Whois(adm)
			}
		}},
		"remove_admin": &botCommand{true, func(who, arg, nick string) {
			for _, adm := range strings.Split(arg, " ") {
				if isAdmin(adm) {
					removeAdmin(adm)
					sendMessage(adm, "see you space cowboy...")
				}
			}
		}},
		"join": &botCommand{true, func(who, arg, nick string) {
			for _, ch := range strings.Split(arg, " ") {
				irc.Join(ch)
			}
		}},
		"part": &botCommand{true, func(who, arg, nick string) {
			irc.Part(who, arg)
		}},
		"tone_police": &botCommand{false, func(who, arg, nick string) {
			if strings.TrimSpace(arg) == "" {
				sendMessage(who, "usage: .tone_police [INPUT]")
				return
			}
			lines := tonePolice([]byte(`{"text":"` + arg + `"}`))
			sendMessage(who, strings.Join(lines, " | "))
		}},
		"add_rss": &botCommand{true, func(who, arg, nick string) {
			if strings.TrimSpace(arg) == "" {
				sendMessage(who, "usage: .add_rss [URL]")
				return
			}
			subs = append(subs, &subscription{who: who, src: arg})
			log.Printf("\n\nsubs:%#v\n\n", subs)
			err := feed.Fetch(arg, nil)
			if err != nil {
				log.Panicln(err)
				sendMessage(nick, err.Error())
			} else {
				sendMessage(who, "Subscribed "+who+" to "+arg)
			}
		}},
		"trans": &botCommand{false, func(who, arg, nick string) {
			arg = strings.Replace(arg, "/", "", -1)
			sendMessage(who, translate(arg))
		}},
	}
	irc.HandleFunc("307", func(c *client.Conn, l *client.Line) {
		if l.Args[0] == nick {
			addAdmin(l.Args[1])
			sendMessage(l.Args[1], "you know what you doing")
		}
		//log.Println("\n\n---\ngot auth !!\n")
		//log.Printf("%#v %#v\n", c, l)
	})
	//irc.HandleFunc("318", func(c *client.Conn, l *client.Line) {
	//log.Println("\n\n---\ngot end of whois\n\n")
	//log.Printf("%#v %#v\n", c, l)
	//})
	irc.HandleFunc(client.PRIVMSG, func(c *client.Conn, l *client.Line) {
		//log.Printf("%#v\n", l)
		who, msg := l.Args[0], l.Args[1]
		if who == nick {
			who = l.Nick
		}
		parseMessage(who, msg, l.Nick)
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
	})

	if err := irc.ConnectTo(host); err != nil {
		log.Fatalln("Connection error:", err)
	}

	for {
		select {
		case bait := <-noiz:
			log.Printf("\n\nbait:%#v\n\n", bait)
			for _, ch := range subs {
				if ch.src == bait.src {
					sendMessage(ch.who, fmt.Sprintf("%s — !%s", bait.tit, bait.url))
				}
			}
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
}
