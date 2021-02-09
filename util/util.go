package util

import (
	"encoding/json"
	"io/ioutil"
	"strings"
)

// useful for creating commands
func Prefix(str, prefix string) bool {
	if strings.Split(str, " ")[0] == prefix {
		return true
	}

	return false
}

// get the arguements from command
func GetArgs(str string) []string {
	return strings.Split(str, " ")[1:] // leaves the prefix out
}

// checks whether str is a link
func IsLink(str string) bool {
	pstr := strings.Split(str, "/")

	if pstr[0] == "https:" || pstr[0] == "http:" {
		if pstr[1] == "" {
			return true
		}
	}

	return false
}
