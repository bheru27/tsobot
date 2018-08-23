package main

import (
	"fmt"
	"math/rand"
	"os"
	"reflect"
	"sync"
	"testing"
)

func assert(t *testing.T, a, b interface{}) {
	if !reflect.DeepEqual(a, b) {
		fmt.Printf("expected: %#v\n", b)
		fmt.Printf("actual:   %#v\n", a)
		t.FailNow()
	}
}

func TestScoreboard(t *testing.T) {
	os.Remove("test.txt")
	defer os.Remove("test.txt")
	sb := newScoreboard("test.txt")
	assert(t, sb.Score("user"), &Score{Nick: "user", Rank: 1})

	plus := rand.Intn(1000)
	minus := rand.Intn(1000)
	total := plus - minus

	wg := &sync.WaitGroup{}
	wg.Add(plus + minus)
	for i := 0; i < plus; i++ {
		go func() {
			sb.AddPoint("user")
			wg.Done()
		}()
	}
	for i := 0; i < minus; i++ {
		go func() {
			sb.RmPoint("user")
			wg.Done()
		}()
	}
	wg.Wait()
	assert(t, sb.Score("user"), &Score{Nick: "user", Rank: 1, Plus: plus, Minus: minus, Total: total})
	sb.Save()
}

func TestSed(t *testing.T) {
	{
		res, err := sed("APPLES", "apples", "bananas", true, false)
		assert(t, err, nil)
		assert(t, res, "bananas")
	}
	{
		res, err := sed("lmao dude weed", `(\w+)\s(\w+)\s(\w+)`, `\2 \3 \1`, false, false)
		assert(t, err, nil)
		assert(t, res, "dude weed lmao")
	}
	{
		res, err := sed("lMAo dUdE WeED", `(lmao) (dude) (weed)`, `\2 \3 \1`, true, false)
		assert(t, err, nil)
		assert(t, res, "dUdE WeED lMAo")
	}
}
func TestSeddy(t *testing.T) {
	res, err := seddy("dude weed lmao", `s/(\w+).*/\1\1\1\1`)
	assert(t, err, nil)
	assert(t, res, "dudedudedudedude")
}
