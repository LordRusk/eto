package main

import (
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/diamondburned/arikawa/v2/gateway"
	"github.com/lordrusk/eto/util"
)

var (
	// path to help config
	helpPath = "./help.json" // TODO when 1.16 releases fully, use new feature to embed for binaries

	// cgg = util.CmdGroups{
	// 	"basic": []util.Cmd{
	// 		{Cmd: "help", Desc: "Generate the help message"},
	// 		{Cmd: "prefix", Args: []util.Arg{{Name: "New Prfix"}}, Desc: "Generate the help message"},
	// 	},
	// 	"music": []util.Cmd{
	// 		{Cmd: "music", Desc: "Start a new music session"},
	// 		{Cmd: "kill", Desc: "Kill current music session"},
	// 		{Cmd: "play", Args: []util.Arg{{Name: "[Search Term || Video Link]"}}, Desc: "Kill current music session"},
	// 		{Cmd: "skip", Desc: "Skip current song"},
	// 		{Cmd: "queue", Desc: "Print the queue"},
	// 	},
	// }
	// _ = util.StoreModel(helpPath, cgg)

	cg = make(util.CmdGroups)
	_  = util.GetStoredModel(helpPath, &cg)
)

var basicLog = log.New(os.Stdout, "basic: ", 0)

func help(m *gateway.MessageCreateEvent) {
	if !util.Prefix(m.Content, fmt.Sprintf("%shelp", *prefix)) {
		return
	}
	embed, err := util.GenHelpMsg(*prefix, *botname, cg)
	if err != nil {
		basicLog.Printf("Failed to generate help message: %s\n")
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
	if !util.Prefix(m.Content, fmt.Sprintf("%sprefix", *prefix)) {
		return
	}

	args := util.GetArgs(m.Content)
	if len(args) < 1 {
		if _, err := s.SendMessage(m.ChannelID, "No prefix given!", nil); err != nil {
			musicLog.Printf("Failed to send message: %s\n")
		}

		return
	}

	*prefix = strings.Join(args, "")
	if _, err := s.SendMessage(m.ChannelID, fmt.Sprintf("`%s` is the new prefix!", *prefix), nil); err != nil {
		basicLog.Printf("Failed to send message: %s\n", err)
	}
}
