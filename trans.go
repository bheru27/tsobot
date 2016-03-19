package main

import (
	"os/exec"
	"regexp"
)

var escapeshellstringre *regexp.Regexp = regexp.MustCompile(`([\(\)\[\]\{\}\$\#&;` + "`" + `\|\*\?~<>\^'"\s])`)

func EscapeShellString(str string) string {
	return escapeshellstringre.ReplaceAllString(str, "\\$1")
}

func translate(text string) string {
	cmd := exec.Command("sh", "-c", "trans -brief "+EscapeShellString(text))
	b, err := cmd.Output()
	checkErr(err)
	return string(b)
}