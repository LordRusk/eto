package util

import "strings"

// get the arguements from command
func GetArgs(str, prefix string) []string {
	return strings.Split(strings.SplitAfterN(str, prefix, 2)[1], " ")[1:] // leaves the prefix and command out
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
