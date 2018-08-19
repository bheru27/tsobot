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
}
