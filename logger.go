package main

import (
	"log"
	"os"

	"github.com/diamondburned/arikawa/v2/gateway"
)

var llogger = log.New(os.Stdout, "Logger: ", 0)

func sent(m *gateway.MessageCreateEvent) {
	guild, err := s.Guild(m.GuildID)
	if err != nil {
		llogger.Printf("Failed to get guild '%s': %s\n", m.GuildID, err)
		return
	}

	channel, err := s.Channel(m.ChannelID)
	if err != nil {
		llogger.Printf("Failed to get channel '%s': %s\n", m.ChannelID, err)
		return
	}

	llogger.Printf("%s: %s: %s: %s\n", guild.Name, channel.Name, m.Author.Username, m.Content)
}

func unsent(e *gateway.MessageDeleteEvent) {
	// Grab from the state
	m, err := s.Message(e.ChannelID, e.ID)
	if err != nil {
		llogger.Printf("Message not found: %d\n", e.ID)
		return
	}

	guild, err := s.Guild(e.GuildID)
	if err != nil {
		llogger.Printf("Failed to get guild '%s': %s\n", m.GuildID, err)
		return
	}

	channel, err := s.Channel(e.ChannelID)
	if err != nil {
		llogger.Printf("Failed to get channel '%s': %s\n", m.ChannelID, err)
		return
	}

	llogger.Printf("%s: %s: %s deleted \"%s\"\n", guild.Name, channel.Name, m.Author.Username, m.Content)
}
