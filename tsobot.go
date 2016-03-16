package main

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"math/rand"
	"net/http"
	"os"
	"os/exec"
	"os/signal"
	"regexp"
	"strings"
	"time"

	"github.com/fluffle/goirc/client"
	"github.com/fluffle/goirc/logging"
	rss "github.com/jteeuwen/go-pkg-rss"
)

var host string
var port int
var ssl bool
var nick string
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

func tonePolice(txt []byte) []string {
	req, err := http.NewRequest("POST", "https://gateway.watsonplatform.net/tone-analyzer-beta/api/v3/tone?version=2016-02-11", bytes.NewBuffer(txt))
	checkErr(err)
	req.SetBasicAuth(u, p)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Watson-Learning-Opt-Out", "1")

	client := &http.Client{}
	res, err := client.Do(req)
	checkErr(err)
	defer res.Body.Close()

	fmt.Println("Status:", res.Status)
	fmt.Println("Headers:", res.Header)
	b, _ := ioutil.ReadAll(res.Body)
	fmt.Println("Body:", string(b))

	lines := []string{"(" + res.Status + ")"}
	if res.Status == "200 OK" {

		r := parseJson(b)
		for _, c := range r {
			fmt.Println(c.Name + ":")
			line := strings.Split(c.Name, " ")[0] + ": "
			emot := []string{}
			for _, t := range c.Tones {
				fmt.Printf("\t%s: %f\n", t.Name, t.Score)
				if t.Score > 0.0 {
					emot = append(emot, fmt.Sprintf("%s %.0f%%", t.Name, t.Score*100.0))
				}
			}
			out := strings.Join(emot, ", ")
			if out == "" {
				out = "(empty)"
			}
			line += out
			lines = append(lines, line)
		}
	}
	return lines
}

type Tone struct {
	Name  string
	Score float64
}

type Category struct {
	Name  string
	Tones []*Tone
}

func parseJson(b []byte) (results []*Category) {
	var d map[string]interface{}
	checkErr(json.Unmarshal(b, &d))

	cats := d["document_tone"].(map[string]interface{})["tone_categories"].([]interface{})

	for _, cat_iface := range cats {
		cat := cat_iface.(map[string]interface{})
		name := cat["category_name"].(string)
		c := &Category{Name: name}
		tones := cat["tones"].([]interface{})
		for _, tone_iface := range tones {
			tone := tone_iface.(map[string]interface{})
			name := tone["tone_name"].(string)
			score := tone["score"].(float64)
			c.Tones = append(c.Tones, &Tone{Name: name, Score: score})
		}
		results = append(results, c)
	}

	return results
}

var escapeshellstringre *regexp.Regexp = regexp.MustCompile(`([\(\)\[\]\{\}\$\#&;` + "`" + `\|\*\?~<>\^'"\s])`)

func EscapeShellString(str string) string {
	return escapeshellstringre.ReplaceAllString(str, "\\$1")
}

func translate(text string) string {
	cmd := exec.Command("sh", "-c", "trans -brief "+EscapeShellString(text))
	b, err := cmd.Output()
	checkErr(err)
	return string(b)
}

func main() {
	flag.StringVar(&host, "host", "irc.rizon.net", "host")
	flag.IntVar(&port, "port", 6697, "port")
	flag.BoolVar(&ssl, "ssl", true, "use ssl?")
	flag.StringVar(&nick, "nick", "tsobot", "nick")
	flag.StringVar(&join, "join", "tso", "join these channels (space separated list)")

	flag.StringVar(&u, "wuname", "", "watson username")
	flag.StringVar(&p, "wpword", "", "watson password")

	flag.Parse()

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, os.Interrupt)

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

	ded := make(chan struct{})
	irc.HandleFunc(client.CONNECTED, func(c *client.Conn, l *client.Line) {
		for _, ch := range strings.Split(join, " ") {
			irc.Join("#" + ch)
		}
	})
	irc.HandleFunc(client.DISCONNECTED, func(c *client.Conn, l *client.Line) {
		close(ded)
	})
	cmdRegexp := regexp.MustCompile(`:([^\s]+?):`)

	feed := rss.New(5, false, func(f *rss.Feed, newchannels []*rss.Channel) {
	}, func(f *rss.Feed, channel *rss.Channel, items []*rss.Item) {
		for i, item := range items {
			if len(item.Links) > 0 {
				<-time.After(time.Second * 3)
				irc.Privmsg("tso", fmt.Sprintf("[%d] %s : %s", i, item.Links[0].Href, item.Title))
			}
		}
	})

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
		if msg == ".test" {
			irc.Who(who)
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
			irc.Privmsg(who, translate(text))
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
