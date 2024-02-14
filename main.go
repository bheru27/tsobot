package main

import (
	"bytes"
	"crypto/tls"
	"flag"
	"fmt"
	"io"
	"log"
	"math/rand"
	"os"
	"os/signal"
	"regexp"
	"runtime/debug"
	"strings"
	"sync"
	"time"

	"github.com/fluffle/goirc/client"
	"github.com/fluffle/goirc/logging"
	"github.com/bheru27/tsobot/dongers"
)

//
// boring stuff
//

var (
	host      string
	port      int
	ssl       bool
	nick      string
	pass      string
	join      string
	admin     string
	cache_dir string

	irc         *client.Conn
	sendMessage func(who, msg string)

	shitlist   []string
	shitlistMu sync.Mutex
)

func parseMessage(who, msg, nick string) bool {
	defer func() {
		if x := recover(); x != nil {
			sendMessage(who, "🔥 🔥 🔥 "+dongers.Raise("panic")+"🔥 🔥 🔥")
			log.Println(x)
			debug.PrintStack()
		}
	}()

	shitlistMu.Lock()
	for _, shitter := range shitlist {
		if strings.ToLower(nick) == shitter {
			shitlistMu.Unlock()
			return true
		}
	}
	shitlistMu.Unlock()

	if !botCommandRe.MatchString(msg) {
		return false
	}

	m := botCommandRe.FindStringSubmatch(msg)
	cmd, arg := m[1], m[2]

	if c, ok := botCommands[cmd]; ok {
		if !c.admin || (c.admin && isAdmin(nick)) {
			c.fn(who, arg, nick)
		} else {
			//log.Printf("%#v\n", botAdmins)
			sendMessage(nick, "Access denied.")
		}
		return true
	}
	return false
}

//
// exciting stuff
//

var (
	sb        *Scoreboard
	chatlogs  = map[string][][]string{}
	chatlogMu sync.Mutex
)

func main() {
	//
	// boring stuff
	//
	flag.StringVar(&host, "host", "irc.rizon.net", "host")
	flag.IntVar(&port, "port", 6697, "port")
	flag.BoolVar(&ssl, "ssl", true, "use ssl?")

	flag.StringVar(&nick, "nick", "tsobotv2", "nick")
	flag.StringVar(&pass, "pass", "", "NickServ IDENTIFY password (optional)")
	flag.StringVar(&join, "join", "/g/punk", "space separated list of channels to join")

	flag.StringVar(&admin, "admin", "GreyMan", "space separated list of privileged nicks")
	flag.StringVar(&cache_dir, "cache_dir", ".cache", "directory to cache datas like rss feeds")

	flag.Parse()

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, os.Interrupt)
	signal.Notify(sig, os.Kill)
	go func() {
		<-sig
		log.Println("we get signal")
		for _, ch := range strings.Split(join, " ") {
			irc.Part("#"+ch, "we get signal")
		}
		<-time.After(time.Second)
		irc.Quit()
		if sb != nil {
			sb.Save()
		}
		saveQuotes("quotes.json")
		saveTodo("todo.json")
		os.Exit(0)
	}()

	// periodically write this stuff to disk I guess
	go func() {
		for {
			<-time.After(time.Minute * 30)
			if sb != nil {
				sb.Save()
			}
			saveQuotes("quotes.json")
			saveTodo("todo.json")
		}
	}()

	// goirc logging...
	logging.SetLogger(&debugLogger{})

	cfg := client.NewConfig(nick)
	if ssl {
		cfg.SSL = true
		cfg.SSLConfig = &tls.Config{ServerName: host, InsecureSkipVerify: true}
		cfg.NewNick = func(n string) string { return n + "^" }
	}
	cfg.Server = fmt.Sprintf("%s:%d", host, port)
	irc = client.Client(cfg)

	irc.HandleFunc(client.CONNECTED, func(c *client.Conn, l *client.Line) {
		if pass != "" {
			irc.Privmsg("NickServ", "IDENTIFY "+pass)
		}
		for _, ch := range strings.Split(join, " ") {
			irc.Join("#" + ch)
		}
		for _, nick := range strings.Split(admin, " ") {
			addAdmin(nick)
		}
	})

	ded := make(chan struct{})
	irc.HandleFunc(client.DISCONNECTED, func(c *client.Conn, l *client.Line) {
		close(ded)
	})

	sendMessage = func(who, msg string) { irc.Privmsg(who, msg) }

	//
	// exciting stuff
	//

	sb = newScoreboard("scoreboard.json")
	loadQuotes("quotes.json")
	loadTodo("todo.json")

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

	// XXX
	// XXX                      THIS IS THE BAD ZONE
	// XXX

	logMessage := func(who, msg, nick string) {
		chatlogMu.Lock()
		defer chatlogMu.Unlock()
		chatlog, ok := chatlogs[who]
		if !ok {
			chatlog = make([][]string, 100, 100)
		}
		for i := 98; i >= 0; i-- {
			chatlog[i+1] = chatlog[i]
		}
		chatlog[0] = []string{strings.ToLower(nick), msg}
		chatlogs[who] = chatlog
	}
	botCommands["mock"] = &botCommand{
		false,
		func(who, arg, nick string) {
			user := strings.SplitN(arg, " ", 1)[0]
			user = strings.TrimSpace(user)

			if !strings.HasPrefix(who, "#") || user == "" {
				return
			}
			chatlogMu.Lock()
			defer chatlogMu.Unlock()
			chatlog, ok := chatlogs[who]
			if !ok || len(chatlog) == 0 {
				return
			}
			for _, ln := range chatlog {
				if ln == nil {
					break
				}
				if ln[0] == strings.ToLower(user) {
					rn := make([]rune, len([]rune(ln[1])))
					for i, c := range []rune(ln[1]) {
						if i&1 == 1 {
							rn[i] = rune(strings.ToUpper(string(c))[0])
						} else {
							rn[i] = rune(strings.ToLower(string(c))[0])
						}
					}
					sendMessage(who, string(rn))
					return
				}
			}
		}}

	trySeddy := func(who, msg, nick string) {
		if strings.Contains(msg, ": s/") {
			ln := strings.SplitN(msg, ": ", 2)
			if len(ln) != 2 {
				return
			}
			nick, msg = ln[0], ln[1]
		}
		if strings.HasPrefix(msg, "s/") {
			chatlogMu.Lock()
			chatlog, ok := chatlogs[who]
			if !ok {
				chatlogMu.Unlock()
				return
			}
			chat := chatlog[:]
			chatlogMu.Unlock()

			for _, ln := range chat {
				if ln == nil {
					break
				}
				if ln[0] == strings.ToLower(nick) {
					res, err := seddy(ln[1], msg)
					if err != nil {
						irc.Privmsg(who, err.Error())
						return
					}
					if res != "" {
						irc.Privmsg(who, res)
						return
					}
				}
			}
		}
	}

	// XXX
	// XXX  END OF BAD ZONE
	// XXX

	timestampRe := regexp.MustCompile(`\d+:\d+:\d+`)

	msgHandler := func(c *client.Conn, l *client.Line) {
		//log.Printf("%#v\n", l)
		who, msg := l.Args[0], l.Args[1]
		if who == nick {
			who = l.Nick
		}
		if parseMessage(who, msg, l.Nick) {
			return
		}
		if emojiRe.MatchString(msg) && !timestampRe.MatchString(msg) {
			matches := emojiRe.FindAllStringSubmatch(msg, -1)
			if len(matches) == 0 {
				return
			}
			originalMsg := msg
			for _, m := range matches {
				new := ""
				switch m[1] {
				case "anger", "disgust", "fear", "happiness", "neutral", "sadness", "surprise", "panic":
					new = dongers.Raise(m[1])
				default:
					if e, ok := emoji[m[1]]; ok {
						new = e
					} else if o, ok := other[m[1]]; ok {
						new = o[rand.Intn(len(o))]
					} else if j, ok := jmote[m[1]]; ok {
						new = j[rand.Intn(len(j))]
					}
				}
				if new != "" {
					msg = strings.Replace(msg, m[0], new, 1)
				}
			}
			if msg != originalMsg {
				irc.Privmsg(who, msg)
				return
			}
		}
		trySeddy(who, msg, l.Nick)
		logMessage(who, msg, l.Nick)
		msg = strings.ToLower(msg)
		if strings.Contains(msg, "normie") || strings.Contains(msg, "normalfag") || strings.Contains(msg, "normans") {
			rand.Seed(time.Now().UnixNano())
			sendMessage(who, "\x02\x034REE"+strings.Repeat("E", rand.Intn(10)))
		} else if msg == "ree" {
			sendMessage(who, "roo normans get out 🐸")
		} else if strings.Contains(msg, ":^)") || strings.Contains(msg, "(^:") {
			switch rand.Intn(5) {
			case 0:
				sendMessage(who, ":^(")
			case 1:
				sendMessage(who, ":^)))"+strings.Repeat(")", rand.Intn(7)))
			case 2:
				sendMessage(who, ". .")
				<-time.After(time.Second)
				sendMessage(who, "  >")
				<-time.After(time.Second)
				sendMessage(who, "\\_/")
			case 3:
				sendMessage(who, strings.Repeat(":^) (^: ", rand.Intn(3)))
			case 4:
				sendMessage(who, ":^|")
				<-time.After(time.Second)
				sendMessage(who, ">:^|")
			case 5:
				sendMessage(who, "c^:")
			}
		}
	}
	irc.HandleFunc(client.PRIVMSG, msgHandler)
	irc.HandleFunc(client.ACTION, msgHandler)

	if err := irc.ConnectTo(host); err != nil {
		log.Fatalln("Connection error:", err)
	}

	<-ded
	log.Println("disconnected.")
	if sb != nil {
		sb.Save()
	}
	saveQuotes("quotes.json")
	saveTodo("todo.json")
	os.Exit(1)

}

func checkErr(err error) {
	if err != nil {
		log.Panicln(err)
	}
}

func fileExists(filename string) bool {
	f, err := os.Open(filename)
	if os.IsNotExist(err) {
		return false
	}
	checkErr(f.Close())
	checkErr(err)
	return true
}

func fileGetContents(filename string) []byte {
	contents := new(bytes.Buffer)
	f, err := os.Open(filename)
	checkErr(err)
	_, err = io.Copy(contents, f)
	f.Close()
	if err != io.EOF {
		checkErr(err)
	}
	return contents.Bytes()
}

func filePutContents(filename string, contents []byte) {
	f, err := os.Create(filename)
	checkErr(err)
	_, err = f.Write(contents)
	checkErr(err)
	checkErr(f.Close())
}

// goirc logging...
type debugLogger struct{}

func (l *debugLogger) Debug(f string, a ...interface{}) { fmt.Printf(f+"\n", a...) }
func (l *debugLogger) Info(f string, a ...interface{})  { fmt.Printf(f+"\n", a...) }
func (l *debugLogger) Warn(f string, a ...interface{})  { fmt.Printf(f+"\n", a...) }
func (l *debugLogger) Error(f string, a ...interface{}) { fmt.Printf(f+"\n", a...) }
