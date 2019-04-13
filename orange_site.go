package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math/rand"
	"net/http"
	"strconv"
)

func hn(storyType string) string {
	items := []int{}
	{
		url := "https://hacker-news.firebaseio.com/v0/" + storyType + "stories.json"

		resp, err := http.Get(url)
		checkErr(err)
		if resp.StatusCode != 200 {
			printResponse(resp)
			return fmt.Sprint(resp.Status)
		}

		checkErr(json.NewDecoder(resp.Body).Decode(&items))
	}

	item := 0
	if len(items) == 0 {
		return "(no items returned)"
	}

	if len(items) > 1 {
		item = items[rand.Intn(len(items)-1)]
	} else {
		item = items[0]
	}

	url := "https://hacker-news.firebaseio.com/v0/item/" + strconv.Itoa(item) + ".json"

	resp, err := http.Get(url)
	checkErr(err)
	if resp.StatusCode != 200 {
		printResponse(resp)
		return fmt.Sprint(resp.Status)
	}

	var story struct {
		Id    int    `json:"id"`
		Title string `json:"title"`
		Url   string `json:"url"`
	}
	checkErr(json.NewDecoder(resp.Body).Decode(&story))

	ret := story.Title
	if story.Url != "" {
		ret += " - ." + story.Url
	} else {
		ret += " - .https://news.ycombinator.com/item?id=" + strconv.Itoa(story.Id)
	}

	return ret
}

func printResponse(resp *http.Response) {
	fmt.Print(resp.Status, "\n")
	for k, v := range resp.Header {
		fmt.Print(k, ": ", v[0], "\n")
	}
	body, _ := ioutil.ReadAll(resp.Body)
	fmt.Print("\n", string(body), "\n")
}
