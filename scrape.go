package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"

	"github.com/generaltso/tsobot/strip"
)

func scrape(uri string) (string, error) {

	v := url.Values{}
	v.Set("sanitize", "y")
	v.Add("url", uri)

	api := "http://localhost:3000/api/get?" + v.Encode()

	log.Println("GET", api)

	res, err := http.Get(api)
	if err != nil {
		return "", err
	}
	defer res.Body.Close()

	log.Println("Status:", res.Status)
	log.Println("Headers:", res.Header)
	b, _ := ioutil.ReadAll(res.Body)
	log.Println("Body:", string(b))

	var readability map[string]interface{}
	err = json.Unmarshal(b, &readability)
	if err != nil {
		return "", err
	}

	log.Printf("%#v\n", readability)

	switch res.StatusCode {
	case 200:
		content := readability["content"].(string)
		return strip.StripTags(content), nil
	case 500:
		var err string
		if e, ok := readability["error"]; ok {
			msg := e.(map[string]interface{})
			err = msg["message"].(string)
		} else {
			err = "No error returned"
		}
		return "", errors.New(fmt.Sprintf("%s: %s", res.Status, err))
	default:
		var err string
		if e, ok := readability["error"]; ok {
			err = e.(string)
		} else {
			err = "No error returned"
		}
		return "", errors.New(fmt.Sprintf("%s: %s", res.Status, err))
	}

}
