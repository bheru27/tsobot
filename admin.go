package main

import (
	"strings"
	"sync"
)

var (
	botAdmins   map[string]struct{}
	botAdminsMu sync.RWMutex
)

func init() {
	botAdmins = map[string]struct{}{}
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
	botCommands["shitcan"] = &botCommand{
		true,
		func(who, arg, nick string) {
			shitlistMu.Lock()
			shitlist = append(shitlist, strings.TrimSpace(arg))
			shitlistMu.Unlock()
			sendMessage(who, "rekt")
		},
	}
}

func isAdmin(nick string) bool {
	botAdminsMu.RLock()
	defer botAdminsMu.RUnlock()
	_, ok := botAdmins[nick]
	return ok
}

func addAdmin(nick string) {
	botAdminsMu.Lock()
	defer botAdminsMu.Unlock()
	botAdmins[nick] = struct{}{}
}

func removeAdmin(nick string) {
	botAdminsMu.Lock()
	defer botAdminsMu.Unlock()
	delete(botAdmins, nick)
}
