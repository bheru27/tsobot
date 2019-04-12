package main

import (
	"encoding/json"
	"fmt"
	"sort"
	"sync"
)

type Score struct {
	Nick  string `json:"nick"`
	Rank  int    `json:"rank"`
	Plus  int    `json:"plus"`
	Minus int    `json:"minus"`
	Total int    `json:"total"`
}

func (s *Score) String() string {
	/*
		if s.Rank == 1 {
			return fmt.Sprintf(
				"HIGH SCORE %s - %d (+%d/-%d) YOU ARE NUMBER ONE",
				s.Nick, s.Total, s.Plus, s.Minus,
			)
		}
	*/

	return fmt.Sprintf(
		"%s %dpts (+%d/-%d)",
		s.Nick, s.Total, s.Plus, s.Minus,
	)
}

type sortableScores []*Score

func (s sortableScores) Len() int { return len(s) }
func (s sortableScores) Less(i, j int) bool {
	a, b := s[i].Total, s[j].Total
	if a == b {
		return s[i].Plus > s[j].Plus
	}
	return a > b
}
func (s sortableScores) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}

type Scoreboard struct {
	filename string
	lookup   map[string]*Score
	mu       *sync.Mutex

	Scores sortableScores `json:"scores"`
}

func newScoreboard(filename string) *Scoreboard {
	sb := &Scoreboard{
		filename: filename,
		lookup:   map[string]*Score{},
		mu:       &sync.Mutex{},
	}
	var scores []*Score
	if fileExists(filename) {
		data := fileGetContents(filename)
		checkErr(json.Unmarshal(data, &sb))
		scores = sb.Scores
		for _, s := range sb.Scores {
			sb.lookup[s.Nick] = s
		}
	} else {
		scores = []*Score{}
	}
	sb.Scores = sortableScores(scores)
	return sb
}
func (sb *Scoreboard) update() {
	sort.Sort(sb.Scores)
	for r, s := range sb.Scores {
		s.Rank = r + 1
	}
}

func (sb *Scoreboard) Score(nick string) *Score {
	sb.mu.Lock()
	defer sb.mu.Unlock()
	return sb.score(nick)
}

func (sb *Scoreboard) score(nick string) *Score {
	if s, ok := sb.lookup[nick]; ok {
		return s
	}
	s := &Score{Nick: nick}
	sb.lookup[nick] = s
	sb.Scores = append(sb.Scores, s)
	s.Rank = len(sb.Scores)
	return s
}
func (sb *Scoreboard) AddPoint(nick string) {
	sb.mu.Lock()
	defer sb.mu.Unlock()
	s := sb.score(nick)
	s.Plus++
	s.Total = s.Plus - s.Minus
	sb.update()
}
func (sb *Scoreboard) RmPoint(nick string) {
	sb.mu.Lock()
	defer sb.mu.Unlock()
	s := sb.score(nick)
	s.Minus++
	s.Total = s.Plus - s.Minus
	sb.update()
}

func (sb *Scoreboard) HighScores() []*Score {
	sb.mu.Lock()
	defer sb.mu.Unlock()
	if len(sb.Scores) < 10 {
		return sb.Scores[:]
	}
	return sb.Scores[:10]
}

func (sb *Scoreboard) Save() {
	sb.mu.Lock()
	defer sb.mu.Unlock()
	b, err := json.Marshal(sb)
	checkErr(err)
	filePutContents(sb.filename, b)
}
