package main

import "github.com/diamondburned/arikawa/v2/gateway"

var logPrefix = "logger: "

func sent(m *gateway.MessageCreateEvent) {
	guild, err := s.Guild(m.GuildID)
	if err != nil {
		logger.Printf("logger: Failed to get guild '%s': %s\n", m.GuildID, err)
		return
	}

	channel, err := s.Channel(m.ChannelID)
	if err != nil {
		logger.Printf("%sFailed to get channel '%s': %s\n", logPrefix, m.ChannelID, err)
		return
	}

	logger.Printf("%s%s: %s: %s: %s\n", logPrefix, guild.Name, channel.Name, m.Author.Username, m.Content)
}

func unsent(e *gateway.MessageDeleteEvent) {
	// Grab from the state
	m, err := s.Message(e.ChannelID, e.ID)
	if err != nil {
		logger.Printf("%sMessage not found: %d\n", logPrefix, e.ID)
		return
	}

	guild, err := s.Guild(e.GuildID)
	if err != nil {
		logger.Printf("%sFailed to get guild '%s': %s\n", logPrefix, m.GuildID, err)
		return
	}

	channel, err := s.Channel(e.ChannelID)
	if err != nil {
		logger.Printf("%sFailed to get channel '%s': %s\n", logPrefix, m.ChannelID, err)
		return
	}

	logger.Printf("logger: %s: %s: %s deleted \"%s\"\n", guild.Name, channel.Name, m.Author.Username, m.Content)
}
