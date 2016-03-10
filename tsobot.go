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
	"os/signal"
	"regexp"
	"strings"

	"github.com/fluffle/goirc/client"
)

var host string
var nick string
var ch string
var u string
var p string

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
			line += strings.Join(emot, ", ")
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

func main() {
	flag.StringVar(&host, "host", "irc.rizon.net", "host")
	flag.StringVar(&nick, "nick", "tsobot", "nick")
	flag.StringVar(&ch, "chan", "#tso", "chan")
	flag.StringVar(&u, "wuname", "", "watson username")
	flag.StringVar(&p, "wpword", "", "watson password")

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
		//		irc.Join(ch)
		irc.Join("#tso")
		irc.Join("#/g/punk")
		irc.Join("#code")
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
			} else if o, ok := other[m[1]]; ok {
				irc.Privmsg(who, o[rand.Intn(len(o))])
			} else {
				irc.Privmsg(who, "(404 emoji not found)")
			}
			return
		}
		if strings.Index(msg, ".tone_police") == 0 {
			if msg == ".tone_police" {
				irc.Privmsg(who, "(feed me data)")
			}
			text := strings.Replace(msg, ".tone_police", "", -1)
			lines := tonePolice([]byte(`{"text":"` + text + `"}`))
			irc.Privmsg(who, strings.Join(lines, " | "))
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
