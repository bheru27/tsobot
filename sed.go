package main

import (
	"fmt"
	"regexp"
	"strings"
)

func sed(s, pattern, replace string, insensitive, global bool) (result string, err error) {
	if insensitive {
		pattern = "(?i)" + pattern
	}
	re, err := regexp.Compile(pattern)
	if err != nil {
		return
	}
	m := re.FindStringSubmatch(s)
	if m == nil {
		return
	}
	n := 1
	if global {
		n = -1
	}
	result = strings.Replace(s, m[0], replace, n)
	for i := 0; i < len(m); i++ {
		result = strings.Replace(result, "\\"+string(byte(48+i)), m[i], -1)
	}
	return
}

func seddy(input, s string) (result string, err error) {
	if !strings.HasPrefix(s, "s/") {
		err = fmt.Errorf("not sed")
		return
	}
	str := []rune(s[2:])
	cmd := [][]rune{[]rune{}}
	for i := 0; i < len(str); i++ {
		if str[i] == '/' {
			if i == 0 || str[i-1] != '\\' {
				cmd = append(cmd, []rune{})
				continue
			}
		}
		cmd[len(cmd)-1] = append(cmd[len(cmd)-1], str[i])
	}
	if len(cmd) < 2 || len(cmd) > 3 {
		err = fmt.Errorf("not sed")
		return
	}
	pat := string(cmd[0])
	rep := string(cmd[1])
	ins := false
	glo := false
	fmt.Printf("%#v\n", cmd)
	if len(cmd) == 3 {
		for _, flag := range cmd[2] {
			switch flag {
			case 'i':
				if ins {
					err = fmt.Errorf("invalid flag argument")
					return
				}
				ins = true
			case 'g':
				if glo {
					err = fmt.Errorf("invalid flag argument")
					return
				}
				glo = true
			default:
				err = fmt.Errorf("invalid flag argument")
				return
			}
		}
	}
	result, err = sed(input, pat, rep, ins, glo)
	return
}
