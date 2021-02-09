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
	return strings.Split(str, " ")[1:]
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

// store a model in a json file
func StoreModel(path string, model interface{}) error {
	jsonBytes, err := json.MarshalIndent(model, "", "	")
	if err != nil {
		return err
	}

	if err := ioutil.WriteFile(path, jsonBytes, 0666); err != nil {
		return err

	}

	return nil
}

// get stored model from json file.
func GetStoredModel(path string, model interface{}) error {
	bytes, err := ioutil.ReadFile(path)
	if err != nil {
		return err
	}

	if err := json.Unmarshal(bytes, model); err != nil {
		return err
	}

	return nil
}
