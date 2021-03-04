package main

import (
	_ "embed"

	"encoding/json"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/diamondburned/arikawa/v2/gateway"
	"github.com/lordrusk/eto/util"
)

var cg util.CmdGroups
var basicLog = log.New(os.Stdout, "basic: ", 0)

//go:embed help.json
var hjson []byte

// prerequisites for the basic commands
func basicSetup() error {
	if err := json.Unmarshal(hjson, &cg); err != nil {
		return fmt.Errorf("Failed to unmarshal 'help.json': %s\n", err)
	}

	return nil
}

func help(m *gateway.MessageCreateEvent) {
	if !strings.HasPrefix(m.Content, fmt.Sprintf("%shelp", *prefix)) {
		return
	}

	embed, err := util.GenHelpMsg(*prefix, *botname, cg)
	if err != nil {
		basicLog.Printf("Failed to generate help message: %s\n", err)
		if _, err := s.SendMessage(m.ChannelID, "Failed to generate help message", nil); err != nil {
			basicLog.Printf("Failed to send message: %s\n", err)
		}

		return
	}

	if _, err := s.SendMessage(m.ChannelID, "", embed); err != nil {
		basicLog.Printf("Failed to send message: %s\n", err)
	}
}

func setPrefix(m *gateway.MessageCreateEvent) {
	if !strings.HasPrefix(m.Content, fmt.Sprintf("%sprefix", *prefix)) {
		return
	}

	args := util.GetArgs(m.Content, *prefix)
	if len(args) < 1 {
		if _, err := s.SendMessage(m.ChannelID, "No prefix given!", nil); err != nil {
			musicLog.Printf("Failed to send message: %s\n", err)
		}

		return
	}

	*prefix = strings.Join(args, " ")
	if _, err := s.SendMessage(m.ChannelID, fmt.Sprintf("`%s` is the new prefix!", *prefix), nil); err != nil {
		basicLog.Printf("Failed to send message: %s\n", err)
	}
}
