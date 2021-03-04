package main

import (
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/diamondburned/arikawa/v2/api"
	"github.com/diamondburned/arikawa/v2/discord"
	"github.com/diamondburned/arikawa/v2/gateway"
)

var llogger = log.New(os.Stdout, "Logger: ", 0)

func loggerSetup() error {
	member, err := s.Member(discord.GuildID(768515803225653248), discord.UserID(650488284991979521))
	if err != nil {
		return fmt.Errorf("Failed to get memeber: %s\n", err)
	}
	user := member.User

	msgs, err := searchMsgUser(user)
	if err != nil {
		return fmt.Errorf("Failed to search messages from user: %s\n", err)
	}
	if err := printMessages(msgs); err != nil {
		return fmt.Errorf("Failed to print messages: %s\n", err)
	}

	// guilds, err := s.Guilds()
	// if err != nil {
	// 	return err
	// }
	//
	// for _, guild := range guilds {
	// 	fmt.Printf("%s: %d\n", guild.Name, guild.ID)
	// }

	// if err := audit(653739264215089162); err != nil {
	// 	return err
	// }

	if err := whosinwhat(); err != nil {
		return err
	}

	return nil
}

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

func listGuilds() error {
	guilds, err := s.Guilds()
	if err != nil {
		return err
	}

	for pos, guild := range guilds {
		llogger.Printf("Guild: %s, Pos: %d, People: ", guild.Name, pos)
		members, err := s.Members(guild.ID)
		if err != nil {
			return err
		}
		for _, member := range members {
			fmt.Printf("%s, ", member.User.Username)
		}
		fmt.Println()
	}

	return nil
}

func searchMsgUser(user discord.User) ([]discord.Message, error) {
	guilds, err := s.Guilds()
	if err != nil {
		return nil, err
	}
	var msgs []discord.Message
	for _, guild := range guilds {
		chans, err := s.Channels(guild.ID)
		if err != nil {
			return nil, err
		}
		for _, channel := range chans {
			messages, err := s.Messages(channel.ID)
			if err != nil {
				return nil, err
			}
			for _, message := range messages {
				if message.Author == user {
					msgs = append(msgs, message)
				}
			}
		}
	}

	return msgs, nil
}

func audit(gid discord.GuildID) error {
	auditlog, err := s.AuditLog(gid, api.AuditLogData{
		// ActionType: discord.AuditLogEvent(72),
		Before: 812927145185640478,
		Limit:  uint(100),
	})
	if err != nil {
		return err
	}

	for _, audit := range auditlog.Entries {
		member, err := s.Member(gid, audit.UserID)
		if err != nil {
			return err
		}
		fmt.Printf("ID: %d, Changes made by: %s, Type: %d, Changes: ", audit.ID, member.User.Username, audit.ActionType)
		for _, change := range audit.Changes {
			fmt.Printf("%+v\n", change)
		}
		fmt.Println()
	}

	return nil
}

func whosinwhat() error {
	guilds, err := s.Guilds()
	if err != nil {
		return fmt.Errorf("Failed to get guilds: %s\n", err)
	}

	for _, guild := range guilds {
		members, err := s.Members(guild.ID)
		if err != nil {
			return fmt.Errorf("Failed to get members: %s\n", err)
		}

		names := make([]string, len(members))
		for pos, member := range members {
			names[pos] = member.User.Username
		}

		fmt.Printf("%s: %s\n", guild.Name, strings.Join(names, ", "))
	}

	return nil
}

func printMessages(msgs []discord.Message) error {
	for _, message := range msgs {
		guild, err := s.Guild(message.GuildID)
		if err != nil {
			return err
		}
		channel, err := s.Channel(message.ChannelID)
		if err != nil {
			return err
		}
		fmt.Printf("%s: %s -> %s %s -> %s | %d embeds | ID: %d\n", message.Author.Username, guild.Name, channel.Name, message.Timestamp.Format(time.UnixDate), message.Content, len(message.Embeds), message.ID)
	}
	return nil
}
