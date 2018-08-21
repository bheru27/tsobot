package main

import (
	"regexp"
	"strings"
)

var (
	botCommandRe = regexp.MustCompile(`^\.(\w+)\s*(.*)$`)
	botCommands  = map[string]*botCommand{}
)

type botCommand struct {
	admin bool
	fn    func(who, arg, nick string)
}

func init() {
	botCommands["bots"] = &botCommand{
		false,
		func(who, arg, nick string) {
			sendMessage(who, "Reporting in! [ "+colorString("Go!", Cyan)+" ] github.com/generaltso/tsobot")
		},
	}
	botCommands["join"] = &botCommand{
		true,
		func(who, arg, nick string) {
			for _, ch := range strings.Split(arg, " ") {
				irc.Join(ch)
			}
		},
	}
	botCommands["part"] = &botCommand{
		true,
		func(who, arg, nick string) {
			irc.Part(who, arg)
		},
	}
	botCommands["trans"] = &botCommand{
		false,
		func(who, arg, nick string) {
			arg = strings.Replace(arg, "/", "", -1)
			sendMessage(who, translate(arg))
		},
	}
	botCommands["addpoint"] = &botCommand{
		false,
		func(who, arg, nick string) {
			user := strings.SplitN(arg, " ", 1)[0]
			if user == "" {
				return
			}
			sb.AddPoint(user)
			sendMessage(who, sb.Score(user).String())
		},
	}
	botCommands["rmpoint"] = &botCommand{
		false,
		func(who, arg, nick string) {
			user := strings.SplitN(arg, " ", 1)[0]
			if user == "" {
				return
			}
			sb.RmPoint(user)
			sendMessage(who, sb.Score(user).String())
		},
	}
	botCommands["score"] = &botCommand{
		false,
		func(who, arg, nick string) {
			s := sb.Score(nick)
			msg := s.String()
			if s.Rank == 1 {
				msg = "H I G H S C O R E " + msg
				m := make([]string, len(msg))
				for i, b := range msg {
					m[i] = colorString(string(b), i%14+2)
				}
				msg = strings.Join(m, "")
			}
			sendMessage(who, msg)
		},
	}
	botCommands["scores"] = &botCommand{
		false,
		func(who, arg, nick string) {
			highscores := sb.HighScores()
			msg := make([]string, 0, len(highscores))
			for i, s := range highscores {
				if s == nil {
					break
				}
				msg[i] = s.String()
			}
			if len(msg) == 0 {
				msg[0] = "(empty)"
			}
			sendMessage(who, strings.Join(msg, " | "))
		},
	}
}
