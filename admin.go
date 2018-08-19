package main

import (
	"sort"
	"strings"
)

func init() {
	botCommands["add_admin"] = &botCommand{
		true,
		func(who, arg, nick string) {
			for _, adm := range strings.Split(arg, " ") {
				irc.Whois(adm)
			}
		},
	}
	botCommands["remove_admin"] = &botCommand{
		true,
		func(who, arg, nick string) {
			for _, adm := range strings.Split(arg, " ") {
				if isAdmin(adm) {
					removeAdmin(adm)
					sendMessage(adm, "see you space cowboy...")
				}
			}
		},
	}
}

var botAdmins sort.StringSlice

func isAdmin(nick string) bool {
	idx := sort.SearchStrings(botAdmins, nick)
	if idx < 0 {
		return false
	}
	return botAdmins[idx] == nick
}

func addAdmin(nick string) {
	botAdmins = sort.StringSlice(append(botAdmins, nick))
}

func removeAdmin(nick string) {
	idx := sort.SearchStrings(botAdmins, nick)
	if idx < 0 || botAdmins[idx] != nick {
		return
	}
	botAdmins = sort.StringSlice(append(botAdmins[:idx], botAdmins[idx+1:]...))
}
