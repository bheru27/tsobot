package main

import (
	"log"
	"regexp"
	"strings"
	"sync"
	"time"
)

var (
	botCommandRe = regexp.MustCompile(`^\.(\w+)\s*(.*)$`)
	botCommands  = map[string]*botCommand{}

	botGlobalCooldown   = map[string]*botCommandGlobalCooldown{}
	botGlobalCooldownMu sync.Mutex

	botCooldown   = map[string]*botCommandCooldown{}
	botCooldownMu sync.Mutex
)

type botCommand struct {
	admin bool
	fn    func(who, arg, nick string)
}

type botCommandGlobalCooldown struct {
	amount time.Duration
	last   time.Time
}

type botCommandCooldown struct {
	amount time.Duration
	users  map[string]time.Time
}

func setGlobalCooldown(command string, amt time.Duration) {
	botGlobalCooldownMu.Lock()
	defer botGlobalCooldownMu.Unlock()

	botGlobalCooldown[command] = &botCommandGlobalCooldown{
		amount: amt,
	}
}

func setCooldown(command string, amt time.Duration) {
	botCooldownMu.Lock()
	defer botCooldownMu.Unlock()

	botCooldown[command] = &botCommandCooldown{
		amount: amt,
		users:  map[string]time.Time{},
	}
}

func globalCooldown(command string) bool {
	botGlobalCooldownMu.Lock()
	defer botGlobalCooldownMu.Unlock()

	cd, ok := botGlobalCooldown[command]
	if !ok {
		log.Println("warning:", command, "has no cooldown time defined but globalCooldown() was called")
		return false
	}

	// fmt.Println("///////////////////////////////\n\n\n")
	// fmt.Printf("amt: %d, last: %s\nnow: %s\nsince: %s", cd.amount, cd.last, time.Now(), time.Since(cd.last))
	// fmt.Println("///////////////////////////////\n\n\n")
	if time.Since(cd.last) >= cd.amount {
		botGlobalCooldown[command].last = time.Now()
		return false
	}

	return true
}

func cooldown(command, nick string) bool {
	botCooldownMu.Lock()
	defer botCooldownMu.Unlock()

	cd, ok := botCooldown[command]
	if !ok {
		log.Println("warning:", command, "has no cooldown time defined but cooldown() was called")
		return false
	}

	last, ok := cd.users[nick]
	if !ok || time.Since(last) >= cd.amount {
		botCooldown[command].users[nick] = time.Now()
		return false
	}

	return true
}

func init() {
	botCommands["bots"] = &botCommand{
		false,
		func(who, arg, nick string) {
			sendMessage(who, "Reporting in! [ "+colorString("Go!", Cyan)+" ] https://github.com/generaltso/tsobot")
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
	setCooldown("addpoint", time.Second*30)
	botCommands["addpoint"] = &botCommand{
		false,
		func(who, arg, nick string) {
			if cooldown("addpoint", nick) {
				return
			}
			if !strings.HasPrefix(who, "#") {
				return
			}
			user := strings.SplitN(arg, " ", 1)[0]
			user = strings.TrimSpace(user)
			if user == "" || strings.ToLower(user) == strings.ToLower(nick) {
				return
			}
			sb.AddPoint(user)
			sendMessage(who, sb.Score(user).String())
		},
	}
	setCooldown("rmpoint", time.Second*30)
	botCommands["rmpoint"] = &botCommand{
		false,
		func(who, arg, nick string) {
			if cooldown("rmpoint", nick) {
				return
			}
			if !strings.HasPrefix(who, "#") {
				return
			}
			user := strings.SplitN(arg, " ", 1)[0]
			user = strings.TrimSpace(user)
			if user == "" || strings.ToLower(user) == strings.ToLower(nick) {
				return
			}
			sb.RmPoint(user)
			sendMessage(who, sb.Score(user).String())
		},
	}
	setCooldown("score", time.Second*15)
	botCommands["score"] = &botCommand{
		false,
		func(who, arg, nick string) {
			if cooldown("score", nick) {
				return
			}
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
	setGlobalCooldown("scores", time.Second*30)
	botCommands["scores"] = &botCommand{
		false,
		func(who, arg, nick string) {
			if globalCooldown("scores") {
				return
			}
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
	botCommands["w"] = &botCommand{
		false,
		func(who, arg, nick string) {
			args := strings.Split(arg, " ")
			freedom := false
			if len(args) > 1 {
				switch args[0] {
				case "-f", "-F", "--fahrenheit", "--freedom-units":
					freedom = true
					args = args[1:]
				}
			}
			sendMessage(who, wttr(strings.Join(args, " "), freedom))
		},
	}
}
