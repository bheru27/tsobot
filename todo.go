package main

import (
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"sync"
)

var (
	todoList   = []string{}
	todoListMu sync.Mutex
)

func loadTodo(filename string) {
	todoListMu.Lock()
	defer todoListMu.Unlock()
	for _, f := range []string{filename, filename + ".bak"} {
		if !fileExists(f) {
			continue
		}
		data := fileGetContents(f)
		if len(data) == 0 {
			continue
		}
		if err := json.Unmarshal(data, &todoList); err != nil {
			log.Printf("warning: could not parse %s: %v", f, err)
			continue
		}
		if f != filename {
			log.Printf("warning: loaded todo from backup %s", f)
		}
		return
	}
}

func saveTodo(filename string) {
	todoListMu.Lock()
	defer todoListMu.Unlock()
	data, err := json.Marshal(todoList)
	checkErr(err)
	filePutContents(filename, data)
}

func addTodo(item string) string {
	todoListMu.Lock()
	defer todoListMu.Unlock()

	todoList = append(todoList, item)

	return fmt.Sprintf(
		"%s [%d thing%s to do]",
		jmote["writing"][rand.Intn(len(jmote["writing"]))],
		len(todoList),
		func() string {
			if len(todoList) == 1 {
				return ""
			}
			return "s"
		}(),
	)
}

func removeTodo(idx int) string {
	todoListMu.Lock()
	defer todoListMu.Unlock()

	if idx < 0 || idx >= len(todoList) {
		return "invalid"
	}

	item := todoList[idx]

	todoList = append(todoList[0:idx], todoList[idx+1:]...)

	return fmt.Sprintf(
		"✓ %s [%sthing%s left to do]",
		item,
		func() string {
			if len(todoList) == 0 {
				return "no"
			}
			return fmt.Sprintf("%d ", len(todoList))
		}(),
		func() string {
			if len(todoList) <= 1 {
				return ""
			}
			return "s"
		}(),
	)
}

func listTodo() []string {
	todoListMu.Lock()
	defer todoListMu.Unlock()

	return todoList[:]
}
