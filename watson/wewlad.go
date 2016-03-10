package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
)

func main() {
	f, _ := os.Open("example_output.json")
	b, _ := ioutil.ReadAll(f)

	var d map[string]interface{}
	json.Unmarshal(b, &d)

	//fmt.Printf("%#v", d)

	cat := d["document_tone"].(map[string]interface{})["tone_categories"].([]interface{})
	//fmt.Printf("%#v", cat)

	for _, category_iface := range cat {
		category := category_iface.(map[string]interface{})
		name := category["category_name"].(string)
		fmt.Println(name + ":")
		tones := category["tones"].([]interface{})
		for _, tone_iface := range tones {
			tone := tone_iface.(map[string]interface{})
			name := tone["tone_name"].(string)
			score := tone["score"].(float64)
			fmt.Printf("\t%s: %f\n", name, score)
		}
	}
}
