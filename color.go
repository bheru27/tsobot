package main

import (
	"fmt"
	"strings"
)

const (
	White      = 0
	Black      = 1
	DarkBlue   = 2
	DarkGreen  = 3
	Red        = 4
	DarkRed    = 5
	DarkViolet = 6
	Orange     = 7
	Yellow     = 8
	LightGreen = 9
	Cyan       = 10
	LightCyan  = 11
	Blue       = 12
	Violet     = 13
	DarkGray   = 14
	LightGray  = 15

	Bold          = 0x02
	Color         = 0x03
	Italic        = 0x09
	StrikeThrough = 0x13
	Reset         = 0x0f
	Underline     = 0x15
	Underline2    = 0x1f
	Reverse       = 0x16
)

func colorString(str string, col ...int) string {
	if len(col) > 1 {
		return fmt.Sprintf("\x03%d,%d%s\x0f", col[0], col[1], str)
	} else {
		return fmt.Sprintf("\x03%d%s\x0f", col[0], str)
	}
}

func rainbowText(text string) string {
	congratulations := []rune(text)
	m := []string{}
	for i, b := range congratulations {
		m = append(m, colorString(string(b), i%14+2))
	}
	return strings.Join(m, "")
}
