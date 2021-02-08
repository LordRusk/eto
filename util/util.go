package util

import "strings"

// useful for creating commands
func Prefix(str, prefix string) bool {
	if strings.Split(str, " ")[0] == prefix {
		return true
	}

	return false
}

// get the arguements from command
func GetArgs(str string) []string {
	return strings.Split(str, " ")[1:]
}
