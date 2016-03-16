package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
)

func tonePolice(txt []byte) []string {
	req, err := http.NewRequest("POST", "https://gateway.watsonplatform.net/tone-analyzer-beta/api/v3/tone?version=2016-02-11", bytes.NewBuffer(txt))
	checkErr(err)
	req.SetBasicAuth(u, p)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Watson-Learning-Opt-Out", "1")

	client := &http.Client{}
	res, err := client.Do(req)
	checkErr(err)
	defer res.Body.Close()

	//log.Println("Status:", res.Status)
	//log.Println("Headers:", res.Header)
	b, _ := ioutil.ReadAll(res.Body)
	//log.Println("Body:", string(b))

	lines := []string{"(" + res.Status + ")"}
	if res.Status == "200 OK" {

		r := parseJson(b)
		for _, c := range r {
			fmt.Println(c.Name + ":")
			line := strings.Split(c.Name, " ")[0] + ": "
			emot := []string{}
			for _, t := range c.Tones {
				fmt.Printf("\t%s: %f\n", t.Name, t.Score)
				if t.Score > 0.0 {
					emot = append(emot, fmt.Sprintf("%s %.0f%%", t.Name, t.Score*100.0))
				}
			}
			out := strings.Join(emot, ", ")
			if out == "" {
				out = "(empty)"
			}
			line += out
			lines = append(lines, line)
		}
	}
	return lines
	// Emotion: Anger 8%, Disgust 10%, Fear 4%, Joy 59%, Sadness 19% | Writing: (empty) | Social: Openness 32%, Conscientiousness 99%, Extraversion 80%, Agreeableness 85%, Emotional Range 7%
}

type Tone struct {
	Name  string
	Score float64
}

type Category struct {
	Name  string
	Tones []*Tone
}

func parseJson(b []byte) (results []*Category) {
	var d map[string]interface{}
	checkErr(json.Unmarshal(b, &d))

	cats := d["document_tone"].(map[string]interface{})["tone_categories"].([]interface{})

	for _, cat_iface := range cats {
		cat := cat_iface.(map[string]interface{})
		name := cat["category_name"].(string)
		c := &Category{Name: name}
		tones := cat["tones"].([]interface{})
		for _, tone_iface := range tones {
			tone := tone_iface.(map[string]interface{})
			name := tone["tone_name"].(string)
			score := tone["score"].(float64)
			c.Tones = append(c.Tones, &Tone{Name: name, Score: score})
		}
		results = append(results, c)
	}

	return results
}
