package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/diamondburned/arikawa/v2/gateway"
	"github.com/diamondburned/arikawa/v2/state"
	"github.com/diamondburned/arikawa/v2/utils/handler"
)

var token = flag.String("t", "", "Set the token (overrides $BOT_TOKEN)")
var botname = flag.String("n", "eto", "Set the bot's name")

var s *state.State
var logger *log.Logger

func main() {
	flag.Parse()
	// logger = log.New(os.Stdout, fmt.Sprintf("%s: ", *botname), 0)
	logger = log.New(os.Stdout, "", 0)
	if *token == "" {
		toke := os.Getenv("BOT_TOKEN")
		if toke == "" {
			logger.Fatalln("No BOT_TOKEN: Set $BOT_TOKEN or use '-t'")
		}
		token = &toke
	}

	var err error
	s, err = state.New(fmt.Sprintf("Bot %s", *token))
	if err != nil {
		logger.Fatalf("Session failed: %s\n", err)
	}

	addPreHandlers()
	addHandlers()
	addIntents()

	if err := s.Open(); err != nil {
		logger.Fatalf("Failed to connect: %s\n", err)
	}
	defer s.Close()

	me, err := s.Me()
	if err != nil {
		logger.Fatalf("Failed to get bot user: %s\n", err)
	}

	logger.Printf("Started as %s\n", me.Username)
	// logger.Println("Bot started")

	// block *forever*
	select {}
}

func addIntents() {
	// s.Gateway.AddIntents(gateway.IntentGuilds)
	// s.Gateway.AddIntents(gateway.IntentGuildMembers)
	// s.Gateway.AddIntents(gateway.IntentGuildBans)
	// s.Gateway.AddIntents(gateway.IntentGuildEmojis)
	// s.Gateway.AddIntents(gateway.IntentGuildIntegrations)
	// s.Gateway.AddIntents(gateway.IntentGuildWebhooks)
	// s.Gateway.AddIntents(gateway.IntentGuildInvites)
	// s.Gateway.AddIntents(gateway.IntentGuildVoiceStates)
	// s.Gateway.AddIntents(gateway.IntentGuildPresences)
	s.Gateway.AddIntents(gateway.IntentGuildMessages)
	// s.Gateway.AddIntents(gateway.IntentGuildMessageReactions)
	// s.Gateway.AddIntents(gateway.IntentGuildMessageTyping)
	s.Gateway.AddIntents(gateway.IntentDirectMessages)
	// s.Gateway.AddIntents(gateway.IntentDirectMessageReactions)
	// s.Gateway.AddIntents(gateway.IntentDirectMessageTyping)
}

func addHandlers() {
	s.AddHandler(sent)
}

func addPreHandlers() {
	s.PreHandler = handler.New()
	s.PreHandler.Synchronous = true

	s.PreHandler.AddHandler(unsent)
}
