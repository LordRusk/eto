package util

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/diamondburned/arikawa/v2/discord"
)

// left exported
var HelpColor = "#fafafa"
var Divider = "------------\n"

// singular arguement for command
type Arg struct {
	Name       string `json:"name,omitempty"`
	IsOptional bool   `json:"isoptional,omitempty"`
}

// singular command:
// A commands state can be one of three values:
// 0 - Working order
// 1 - Work in progress
// 2 - Does not work
type Cmd struct {
	Cmd   string `json:"cmd,omitempty"`
	Args  []Arg  `json:"args,omitempty"`
	Desc  string `json:"desc,omitempty"`
	State int    `json:"State,omitempty"`
}

// group of commands used to organize
// commands
type CmdGroup struct {
	Name   string `json:"name,omitempty"`
	CmdArr []Cmd  `json:"cmdarr,omitempty"`
}

// nice wrap
type CmdGroups map[string][]Cmd

// generate a help message
func GenHelpMsg(prefix string, botName string, cg CmdGroups) (*discord.Embed, error) {
	// generate the help command
	var helpMsg strings.Builder

	helpMsg.WriteString(Divider)
	helpMsg.WriteString(fmt.Sprintf("**Prefix:** `%s`\n%s**Commands**\n", prefix, Divider))
	helpMsg.WriteString(Divider)

	for name, cmds := range cg {
		helpMsg.WriteString(fmt.Sprintf("***%s Commands:***\n", name))
		for _, cmdInfo := range cmds {
			if cmdInfo.State == 1 {
				helpMsg.WriteString("__[WIP]__ ")
			} else if cmdInfo.State == 2 {
				helpMsg.WriteString("~~")
			}
			helpMsg.WriteString(fmt.Sprintf("**%s**", cmdInfo.Cmd))
			for i := 0; i < len(cmdInfo.Args); i++ {
				helpMsg.WriteString(" [ ")
				if cmdInfo.Args[i].IsOptional == true {
					helpMsg.WriteString("*Optional* ")
				}
				helpMsg.WriteString(fmt.Sprintf("%s ]", cmdInfo.Args[i].Name))
			}
			helpMsg.WriteString(fmt.Sprintf(" -- *%s*", cmdInfo.Desc))
			if cmdInfo.State == 2 {
				helpMsg.WriteString("~~")
			}
			helpMsg.WriteString("\n")
		}
		helpMsg.WriteString(Divider)
	}

	// color
	colorHex, err := strconv.ParseInt((HelpColor)[1:], 16, 64)
	if err != nil {
		return nil, err
	}

	embed := discord.Embed{
		Title:       fmt.Sprintf("%s Help Page", strings.Title(botName)),
		Description: helpMsg.String(),
		Color:       discord.Color(colorHex),
	}

	return &embed, nil
}
