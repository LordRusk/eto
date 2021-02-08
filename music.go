package main

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"

	"github.com/diamondburned/arikawa/v2/discord"
	"github.com/diamondburned/arikawa/v2/gateway"
	"github.com/diamondburned/arikawa/v2/voice"
	"github.com/diamondburned/arikawa/v2/voice/voicegateway"
	musik "github.com/lordrusk/eto/music"
	"github.com/lordrusk/eto/util"
)

var dj = make(musik.DJ)
var musicPrefix = "music: "

// errors
var noMusic = fmt.Errorf("No music is currently being played in this guild. Start a music session with `%smusic`", *prefix)
var yesMusic = fmt.Errorf("Music already playing in this guild. Close current session with `%skill`", *prefix)

func music(m *gateway.MessageCreateEvent) {
	if !util.Prefix(m.Content, fmt.Sprintf("%smusic", *prefix)) {
		return
	}

	if dj[m.GuildID] != nil {
		if _, err := s.SendMessage(m.ChannelID, fmt.Sprintf("%s", yesMusic), nil); err != nil {
			logger.Printf("%sFailed to send message: %s\n", musicPrefix, err)
		}

		return
	}

	vs, err := s.VoiceState(m.GuildID, m.Author.ID)
	if err != nil {
		logger.Printf("%sFailed to get voice state of %s: %s\n", musicPrefix, m.Author.Username, err)
		if _, err := s.SendMessage(m.ChannelID, fmt.Sprintf("Cannot join channel: %s not in channel", m.Author.Username), nil); err != nil {
			logger.Printf("%sFailed to send message: %s\n", musicPrefix, err)
		}

		return
	}

	v, err := voice.NewSession(s)
	if err != nil {
		logger.Printf("%sFailed to make new voice session: %s\n", musicPrefix, err)
		return
	}

	dj[m.GuildID] = musik.New(v) // create new session for guild

	if err := dj[m.GuildID].JoinChannel(m.GuildID, vs.ChannelID, false, true); err != nil {
		logger.Printf("%sFailed to join channel: %s\n", musicPrefix, err)
		if _, err := s.SendMessage(m.ChannelID, "Failed to join channel!", nil); err != nil {
			logger.Printf("%sFailed to send message: %s\n", musicPrefix, err)
		}

		return
	}

	// setup the queue system
	go stereo(m.GuildID, m.ChannelID)

	if _, err := s.SendMessage(m.ChannelID, "Music session successfully started!", nil); err != nil {
		logger.Printf("%sFailed to send message: %s\n", musicPrefix, err)
	}
}

func kill(m *gateway.MessageCreateEvent) {
	if !util.Prefix(m.Content, fmt.Sprintf("%skill", *prefix)) {
		return
	}

	if dj[m.GuildID] == nil {
		if _, err := s.SendMessage(m.ChannelID, fmt.Sprintf("%s", noMusic), nil); err != nil {
			logger.Printf("%sFailed to send message: %s\n", musicPrefix, err)
		}

		return
	}

	if dj[m.GuildID].Cancel != nil {
		dj[m.GuildID].Cancel()
	}

	if err := dj[m.GuildID].Leave(); err != nil {
		logger.Printf("%sFailed to leave channel: %s\n", musicPrefix, err)
	}

	dj[m.GuildID] = nil
	if _, err := s.SendMessage(m.ChannelID, "Music session killed!", nil); err != nil {
		logger.Printf("%sFailed to send message: %s\n", musicPrefix, err)
	}
}

func play(m *gateway.MessageCreateEvent) {
	if !util.Prefix(m.Content, fmt.Sprintf("%splay", *prefix)) {
		return
	}

	if dj[m.GuildID] == nil {
		if _, err := s.SendMessage(m.ChannelID, fmt.Sprintf("%s", noMusic), nil); err != nil {
			logger.Printf("%sFailed to send message: %s\n", musicPrefix, err)
		}

		return
	}

	me, err := s.Me()
	if err != nil {
		logger.Printf("%sFailed to get bot's user: %s\n", musicPrefix, err)
		return
	}

	if _, err := s.VoiceState(m.GuildID, me.ID); err != nil {
		logger.Printf("%sFailed to get bot's voice state: %s\n", musicPrefix, err)
		if _, err := s.SendMessage(m.ChannelID, "Cannot play song! Not in channel", nil); err != nil {
			logger.Printf("%sFailed to send message: %s\n", musicPrefix, err)
		}
	}

	var id string
	combArgs := strings.Join(util.GetArgs(m.Content), " ")
	if musik.IsLink(combArgs) {
		id = combArgs
	} else {
		if _, err := s.SendMessage(m.ChannelID, fmt.Sprintf("Searching `%s`", combArgs), nil); err != nil {
			logger.Printf("%sFailed to send message: %s\n", musicPrefix, err)
		}

		id, err = musik.GetID(combArgs)
		if err != nil {
			if _, err := s.SendMessage(m.ChannelID, fmt.Sprintf("%s", err), nil); err != nil {
				logger.Printf("%sFailed to send message: %s\n", musicPrefix, err)
			}

			return
		}
	}

	media, err := musik.GetVideo(id)
	if err != nil {
		logger.Printf("Failed to get video: %s\n", musicPrefix, err)
		if _, err := s.SendMessage(m.ChannelID, "Failed to get video", nil); err != nil {
			logger.Printf("%sFailed to send message: %s\n", musicPrefix, err)
		}
	}

	dj[m.GuildID].Player <- media

	if len(dj[m.GuildID].Player) != 0 {
		dj[m.GuildID].Queue = append(dj[m.GuildID].Queue, media.Title)
		if _, err := s.SendMessage(m.ChannelID, fmt.Sprintf("`%s` Added to queue", media.Title), nil); err != nil {
			logger.Printf("%sFailed to send message: %s\n", musicPrefix, err)
		}
	}
}

func skip(m *gateway.MessageCreateEvent) {
	if !util.Prefix(m.Content, fmt.Sprintf("%sskip", *prefix)) {
		return
	}

	if dj[m.GuildID] == nil {
		if _, err := s.SendMessage(m.ChannelID, fmt.Sprintf("%s", noMusic), nil); err != nil {
			logger.Printf("%sFailed to send message: %s\n", musicPrefix, err)
		}

		return
	}

	if dj[m.GuildID].Cancel != nil {
		dj[m.GuildID].Cancel()
		dj[m.GuildID].Cancel = nil
	}
}

func queue(m *gateway.MessageCreateEvent) {
	if !util.Prefix(m.Content, fmt.Sprintf("%squeue", *prefix)) {
		return
	}

	if dj[m.GuildID] == nil {
		if _, err := s.SendMessage(m.ChannelID, fmt.Sprintf("%s", noMusic), nil); err != nil {
			logger.Printf("%sFailed to send message: %s\n", musicPrefix, err)
		}

		return
	}

	if len(dj[m.GuildID].Queue) == 0 {
		if _, err := s.SendMessage(m.ChannelID, "No songs in queue", nil); err != nil {
			logger.Printf("%sFailed to send message: %s\n", musicPrefix, err)
		}

		return
	}

	var builder strings.Builder
	for pos, title := range dj[m.GuildID].Queue {
		builder.WriteString(fmt.Sprintf("%s: `%s`\n", strconv.Itoa(pos+1), title))
	}

	if _, err := s.SendMessage(m.ChannelID, builder.String(), nil); err != nil {
		logger.Printf("%sFailed to send message: %s\n", musicPrefix, err)
	}
}

// the function that actually plays the music
func stereo(gid discord.GuildID, cid discord.ChannelID) {
	for dj[gid] != nil {
		media := <-dj[gid].Player
		dj[gid].Playing = &media

		if len(dj[gid].Queue) != 0 { // remove current song from queue
			dj[gid].Queue = dj[gid].Queue[1:]
		}

		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		cmd := exec.CommandContext(ctx,
			"ffmpeg",
			// Streaming is slow, so a single thread is all we need.
			"-hide_banner", "-threads", "1", "-loglevel", "error", "-ss",
			strconv.Itoa(media.StartAt), "-i", "pipe:", "-filter:a", "volume=0.25",
			"-c:a", "libopus", "-b:a", "64k", "-f", "opus", "-",
		)

		oggWriter := musik.NewOggWriter(dj[gid])
		defer oggWriter.Close()
		dj[gid].Cancel = func() { cancel(); oggWriter.Close() }

		cmd.Stdin = media.Stream
		cmd.Stdout = oggWriter
		cmd.Stderr = os.Stderr

		if _, err := s.SendMessage(cid, fmt.Sprintf("Playing `%s`", media.Title), nil); err != nil {
			logger.Printf("%sFailed to send message: %s\n", musicPrefix, err)
		}

		// this shouldnt be called concurrently....
		if err := dj[gid].Speaking(voicegateway.Microphone); err != nil {
			logger.Printf("%sFailed to start speaking: %s\n", musicPrefix, err)
		}

		if err := cmd.Run(); err != nil {
			logger.Printf("%sFailed to run cmd: %s\n", musicPrefix, err)
		}

		if _, err := s.SendMessage(cid, fmt.Sprintf("Finished playing `%s`", media.Title), nil); err != nil {
			logger.Printf("%sFailed to send message: %s\n", musicPrefix, err)
		}

		if dj[gid] != nil {
			if len(dj[gid].Player) == 0 {
				if _, err := s.SendMessage(cid, "Finished queue", nil); err != nil {
					logger.Printf("%sFailed to send message: %s\n", musicPrefix, err)
				}
			}
		}

		if dj[gid] != nil {
			dj[gid].Playing = nil
		}
	}
}
