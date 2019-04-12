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
			if !strings.HasPrefix(who, "#") {
				return
			}
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
			if !strings.HasPrefix(who, "#") {
				return
			}
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
				congratulations := []rune(" ~ H I G H  S C O R E ~")
				m := []string{}
				for i, b := range congratulations {
					m = append(m, colorString(string(b), i%14+2))
				}
				msg = msg + strings.Join(m, "")
			}
			sendMessage(who, msg)
		},
	}
	botCommands["scores"] = &botCommand{
		false,
		func(who, arg, nick string) {
			highscores := sb.HighScores()
			msg := []string{}
			for _, s := range highscores {
				if s == nil {
					break
				}
				msg = append(msg, s.String())
			}
			if len(msg) == 0 {
				sendMessage(who, "(scoreboard is empty)")
				return
			}
			sendMessage(who, strings.Join(msg, " | "))
		},
	}
	botCommands["askhn"] = &botCommand{false, func(who, arg, nick string) { sendMessage(who, hn("ask")) }}
	botCommands["showhn"] = &botCommand{false, func(who, arg, nick string) { sendMessage(who, hn("show")) }}
	botCommands["tophn"] = &botCommand{false, func(who, arg, nick string) { sendMessage(who, hn("top")) }}
	botCommands["besthn"] = &botCommand{false, func(who, arg, nick string) { sendMessage(who, hn("best")) }}
	botCommands["newhn"] = &botCommand{false, func(who, arg, nick string) { sendMessage(who, hn("new")) }}

}
