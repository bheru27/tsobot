package main

import "fmt"

const (
	White      int = 0
	Black      int = 1
	DarkBlue   int = 2
	DarkGreen  int = 3
	Red        int = 4
	DarkRed    int = 5
	DarkViolet int = 6
	Orange     int = 7
	Yellow     int = 8
	LightGreen int = 9
	Cyan       int = 10
	LightCyan  int = 11
	Blue       int = 12
	Violet     int = 13
	DarkGray   int = 14
	LightGray  int = 15

	Bold          int = 0x02
	Color         int = 0x03
	Italic        int = 0x09
	StrikeThrough int = 0x13
	Reset         int = 0x0f
	Underline     int = 0x15
	Underline2    int = 0x1f
	Reverse       int = 0x16
)

func colorString(str string, col ...int) string {
	if len(col) > 1 {
		return fmt.Sprintf("\x03%d,%d%s\x0f", col[0], col[1], str)
	} else {
		return fmt.Sprintf("\x03%d%s\x0f", col[0], str)
	}
}
