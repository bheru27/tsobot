package main

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"reflect"
	"strconv"
)

type Tone struct {
	Anger, Disgust, Fear, Happiness, Neutral, Sadness, Surprise float64
}

func (t Tone) forEach(callbackFn func(string, float64)) {
	st := reflect.TypeOf(t)
	v := reflect.ValueOf(t)
	for i := 0; i < st.NumField(); i++ {
		callbackFn(st.Field(i).Name, v.Field(i).Float())
	}
}

func (t *Tone) Max() (string, float64) {
	emote := "Neutral"
	score := 0.0
	t.forEach(func(e string, s float64) {
		if s > score {
			score = s
			emote = e
		}
	})
	return emote, score
}

func tonePolice(txt []byte) Tone {
	data := url.Values{}
	data.Set("message", string(txt))

	req, err := http.NewRequest("POST", "http://localhost:4567/", bytes.NewBufferString(data.Encode()))
	checkErr(err)
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Add("Content-Length", strconv.Itoa(len(data.Encode())))

	client := &http.Client{}
	res, err := client.Do(req)
	checkErr(err)
	defer res.Body.Close()

	log.Println("Status:", res.Status)
	log.Println("Headers:", res.Header)
	b, _ := ioutil.ReadAll(res.Body)
	log.Println("Body:", string(b))

	var tone Tone
	checkErr(json.Unmarshal(b, &tone))

	log.Printf("%+v\n", tone)

	return tone
}
