package main

import (
	"log"
	"os"
	"os/signal"
	"regexp"
	"time"

	"github.com/sorcix/irc"
)

func main() {
	cmdRegexp := regexp.MustCompile(`:(\w+):`)
	c, err := irc.Dial("irc.rizon.net:6667")
	defer c.Close()
	if err != nil {
		log.Fatalln(err)
	}
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, os.Interrupt)
	go func() {
		<-sig
		log.Println("we get signal")
		c.Close()
		os.Exit(0)
	}()

	messages := make(chan *irc.Message)
	go func() {
		for {
			msg, err := c.Decode()
			if err != nil {
				log.Fatalln(err)
			}
			messages <- msg
		}
	}()
	c.Encode(&irc.Message{
		Command: irc.NICK,
		Params:  []string{"tsobot"},
	})
	c.Encode(&irc.Message{
		Command:  irc.USER,
		Params:   []string{"tsobot", "0", "*"},
		Trailing: "tsobot",
	})
	<-time.After(time.Second)
	c.Encode(&irc.Message{
		Command: irc.JOIN,
		Params:  []string{"#tso"},
	})
	for {
		select {
		case msg := <-messages:
			log.Println(msg.String())
			if msg.Command == irc.PING {
				c.Encode(&irc.Message{
					Command:  irc.PONG,
					Params:   msg.Params,
					Trailing: msg.Trailing,
				})
			}
			if msg.Command == irc.PRIVMSG && msg.Trailing == ".bots" {
				log.Printf("%#v\n", msg)
				c.Encode(&irc.Message{
					Command:  irc.PRIVMSG,
					Params:   []string{"#tso"},
					Trailing: "Repor—",
				})
				//				c.Encode(&irc.Message{
				//					Command: irc.QUIT,
				//				})
				//				c.Close()
				//				return
			}
			if msg.Command == irc.PRIVMSG {
				if cmdRegexp.MatchString(msg.Trailing) {
					m := cmdRegexp.FindStringSubmatch(msg.Trailing)
					if e, ok := emoji[m[1]]; ok {
						c.Encode(&irc.Message{
							Command:  irc.PRIVMSG,
							Params:   []string{"#tso"},
							Trailing: e,
						})
					}
				}
			}
			//		case <-time.After(time.Second * 120):
			//			log.Fatalln("Timed out.")
		}
	}
}
