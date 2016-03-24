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

/**
 * Arbitrary way GoIRC handles logging
 */
type tsoLogger struct{}

// mfw
//var lines chan string

func (l *tsoLogger) Debug(f string, a ...interface{}) {
	//lines <- a[0].(string)

	log.Printf("\n\n DEBUG \n\n"+f+"\n", a...)
}
func (l *tsoLogger) Info(f string, a ...interface{})  { log.Printf("\n\n INFO \n\n"+f+"\n", a...) }
func (l *tsoLogger) Warn(f string, a ...interface{})  { log.Printf("\n\n WARN \n\n"+f+"\n", a...) }
func (l *tsoLogger) Error(f string, a ...interface{}) { log.Printf("\n\n ERROR \n\n"+f+"\n", a...) }

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
var botCommands map[string]func(who, arg, nick string)
var sendMessage func(who, msg string)

func parseMessage(who, msg, nick string) {
	if !botCommandRe.MatchString(msg) {
		return
	}

	m := botCommandRe.FindStringSubmatch(msg)
	cmd := m[1]
	arg := m[2]

	if fn, ok := botCommands[cmd]; ok {
		fn(who, arg, nick)
	}
}

func isAdmin(nick string) bool {
	ind := sort.SearchStrings(botAdmins, nick)
	retval := ind < len(botAdmins) && botAdmins[ind] == nick
	if !retval {
		sendMessage(nick, "Access denied.")
	}
	return retval
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

	botCommands = map[string]func(who, arg, nick string){
		"bots": func(who, arg, nick string) {
			sendMessage(who, "Reporting in! \x0310go\x0f get github.com/generaltso/tsobot")
		},
		"test": func(who, arg, nick string) {
			if !isAdmin(nick) {
				return
			}
			irc.Whois(arg)
		},
		"tone_police": func(who, arg, nick string) {
			if strings.TrimSpace(arg) == "" {
				sendMessage(who, "usage: .tone_police [INPUT]")
				return
			}
			lines := tonePolice([]byte(`{"text":"` + arg + `"}`))
			sendMessage(who, strings.Join(lines, " | "))
		},
		"rss": func(who, arg, nick string) {
			if !isAdmin(nick) {
				return
			}
			if strings.TrimSpace(arg) == "" {
				sendMessage(who, "usage: .rss [URL]")
				return
			}
			err := feed.Fetch(arg, nil)
			if err != nil {
				sendMessage(who, err.Error())
			}
		},
		"trans": func(who, arg, nick string) {
			arg = strings.Replace(arg, "/", "", -1)
			sendMessage(who, translate(arg))
		},
	}
	irc.HandleFunc(client.PRIVMSG, func(c *client.Conn, l *client.Line) {
		log.Printf("%#v\n", l)
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
