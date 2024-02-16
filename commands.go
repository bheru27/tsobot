package main

import (
	"bufio"
	"fmt"
	"log"
	"math/rand"
	"os"
	"regexp"
	"strconv"
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

	textfiles   = map[string][]string{}
	textfilesMu sync.Mutex
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
			sendMessage(who, "Reporting in! [ "+colorString("Go!", Cyan)+" ] https://github.com/dayvonjersen/tsobot")
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
	botCommands["addquote"] = &botCommand{
		false,
		func(who, arg, nick string) {
			if !strings.HasPrefix(who, "#") {
				return
			}
			sendMessage(who, addQuote(who, arg, nick))
		},
	}
	botCommands["quote"] = &botCommand{
		false,
		func(who, arg, nick string) {
			if !strings.HasPrefix(who, "#") {
				return
			}

			user := strings.SplitN(arg, " ", 1)[0]
			user = strings.TrimSpace(user)
			if user == "" {
				sendMessage(who, getRandQuote(who))
			} else {
				sendMessage(who, getQuote(who, user))
			}
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
			user := strings.SplitN(arg, " ", 2)[0]
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
				msg = msg + rainbowText(" ~ H I G H  S C O R E ~")
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
				msg = append(msg, "_"+s.String())
			}
			if len(msg) == 0 {
				sendMessage(who, "(scoreboard is empty)")
				return
			}
			sendMessage(who, strings.Join(msg, " "))
		},
	}
	botCommands["askhn"] = &botCommand{false, func(who, arg, nick string) { sendMessage(who, hn("ask")) }}
	botCommands["showhn"] = &botCommand{false, func(who, arg, nick string) { sendMessage(who, hn("show")) }}
	botCommands["tophn"] = &botCommand{false, func(who, arg, nick string) { sendMessage(who, hn("top")) }}
	botCommands["besthn"] = &botCommand{false, func(who, arg, nick string) { sendMessage(who, hn("best")) }}
	botCommands["newhn"] = &botCommand{false, func(who, arg, nick string) { sendMessage(who, hn("new")) }}
	botCommands["goodpost"] = &botCommand{false, func(who, arg, nick string) { sendMessage(who, lobsters()) }}
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
	botCommands["ngate"] = &botCommand{
		false,
		func(who, arg, nick string) {
			snark, isNew := ngate()
			sendMessage(who, snark)
			if isNew {
				<-time.After(time.Second * 5)
				sendMessage(who, rainbowText("new n-gate available")+": http://n-gate.com/")
			}
		},
	}
	botCommands["todo"] = &botCommand{true, func(who, arg, nick string) {
		arg = strings.TrimSpace(arg)
		if arg == "" {
			for i, item := range listTodo() {
				sendMessage(nick, fmt.Sprintf("%d: %s", i, item))
			}
		} else {
			sendMessage(who, addTodo(arg))
		}
	}}
	botCommands["done"] = &botCommand{true, func(who, arg, nick string) {
		arg = strings.TrimSpace(arg)
		idx, _ := strconv.Atoi(arg)
		sendMessage(who, removeTodo(idx))
	}}
	botCommands["shrug"] = &botCommand{false, func(who, arg, nick string) { sendMessage(who, `¯\_(ツ)_/¯`) }}
	botCommands["smug"] = &botCommand{false, func(who, arg, nick string) { sendMessage(who, `( ≖‿≖)`) }}
	botCommands["denko"] = &botCommand{false, func(who, arg, nick string) { sendMessage(who, "(´･ω･`)") }}

	botCommands["noided"] = &botCommand{false, func(who, arg, nick string) { sendMessage(who, randomChoice("grips")) }}
	botCommands["spooky"] = &botCommand{false, func(who, arg, nick string) { sendMessage(who, randomChoice("spooks")) }}
	botCommands["zerowing"] = &botCommand{false, func(who, arg, nick string) { sendMessage(who, randomChoice("zerowing")) }}
	botCommands["schizopost"] = &botCommand{false, func(who, arg, nick string) { sendMessage(who, randomChoice("pawel")) }}

	botCommands["cite"] = &botCommand{false, func(who, arg, nick string) { sendMessage(who, "\x1d\x0312[citation needed]") }}

	cnt := 0
	botCommands["wake"] = &botCommand{false, func(who, arg, nick string) {
		if arg == "" {
			sendMessage(who, "me up inside")
		} else if arg == "me up inside" {
			sendMessage(who, "(can't wake up)")
			go func() {
				<-time.After(time.Second * 30)
				sendMessage(who, "call my name and save me from the dark")
				<-time.After(time.Second * 10)
				sendMessage(who, "bid my blood to run")
				<-time.After(time.Second * 5)
				sendMessage(who, "before I come undone")
				<-time.After(time.Second)
				sendMessage(who, bold(color("SAVE ME", Red, Black)))
				<-time.After(time.Second * 3)
				sendMessage(who, italic("Save me from the nothing I've become~"))
			}()
		}
	}
}

func randomChoice(filename string) string {
	textfilesMu.Lock()
	defer textfilesMu.Unlock()

	lines, ok := textfiles[filename]

	if !ok {
		f, err := os.Open("texts/" + filename + ".txt")
		if err != nil {
			return err.Error()
		}
		lines = []string{}
		scanner := bufio.NewScanner(f)
		for scanner.Scan() {
			ln := strings.TrimSpace(scanner.Text())
			if ln != "" {
				lines = append(lines, ln)
			}
		}
		textfiles[filename] = lines
	}

	rand.Seed(time.Now().UnixNano())
	return lines[rand.Intn(len(lines))]
}
