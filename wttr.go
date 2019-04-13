package main

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
)

func wttr(loc string) string {
	loc = strings.TrimSpace(loc)
	loc = url.PathEscape(loc)
	if loc == "" {
		return "usage: .w [location]"
	}
	resp, err := http.Get("http://wttr.in/" + loc + "?format=2")
	checkErr(err)
	if resp.StatusCode != 200 {
		printResponse(resp)
		return fmt.Sprint(resp.StatusCode)
	}
	body, _ := ioutil.ReadAll(resp.Body)
	return string(body)
}
