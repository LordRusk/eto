package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/exec"
	"strconv"
	"strings"

	"github.com/diamondburned/arikawa/v2/discord"
	"github.com/diamondburned/arikawa/v2/gateway"
	"github.com/diamondburned/arikawa/v2/voice"
	"github.com/diamondburned/arikawa/v2/voice/voicegateway"
	"github.com/lordrusk/eto/music"
	"github.com/lordrusk/eto/util"
)

var dj = make(music.DJ)
var musicLog = log.New(os.Stdout, "music: ", 0)

// errors
var noMusic = fmt.Errorf("No music is currently being played in this guild. Start a music session with `%smusic`", *prefix)
var yesMusic = fmt.Errorf("Music already playing in this guild. Close current session with `%skill`", *prefix)

// Play a song. Automatically starts session if needed.
func play(m *gateway.MessageCreateEvent) {
	if !strings.HasPrefix(m.Content, fmt.Sprintf("%splay", *prefix)) {
		return
	}

	args := util.GetArgs(m.Content, *prefix)
	if len(args) < 1 {
		if _, err := s.SendMessage(m.ChannelID, "Song name or link not given!", nil); err != nil {
			musicLog.Printf("Failed to send message: %s\n")
		}

		return
	}

	if dj[m.GuildID] == nil {
		// make sure session is active
		newSesh(m)
	}

	if _, err := s.VoiceState(m.GuildID, u.User.ID); err != nil {
		musicLog.Printf("Failed to get bot's voice state: %s\n", err)
		if _, err := s.SendMessage(m.ChannelID, "Cannot play song! Not in channel", nil); err != nil {
			musicLog.Printf("Failed to send message: %s\n", err)
		}
	}

	id := strings.Join(args, " ")
	if !util.IsLink(id) {
		if _, err := s.SendMessage(m.ChannelID, fmt.Sprintf("Searching `%s`", id), nil); err != nil {
			musicLog.Printf("Failed to send message: %s\n", err)
		}

		var err error
		id, err = music.GetID(id)
		if err != nil {
			if _, err := s.SendMessage(m.ChannelID, fmt.Sprintf("%s", err), nil); err != nil {
				musicLog.Printf("Failed to send message: %s\n", err)
			}

			return
		}
	}

	media, err := music.GetVideo(id)
	if err != nil {
		musicLog.Printf("Failed to get video: %s\n", err)
		if _, err := s.SendMessage(m.ChannelID, "Failed to get video", nil); err != nil {
			musicLog.Printf("Failed to send message: %s\n", err)
		}
	}

	dj[m.GuildID].Player <- media

	if len(dj[m.GuildID].Player) != 0 {
		dj[m.GuildID].Queue = append(dj[m.GuildID].Queue, media.Title)
		if _, err := s.SendMessage(m.ChannelID, fmt.Sprintf("`%s` Added to queue", media.Title), nil); err != nil {
			musicLog.Printf("Failed to send message: %s\n", err)
		}
	}
}

// kill the current session
func kill(m *gateway.MessageCreateEvent) {
	if !strings.HasPrefix(m.Content, fmt.Sprintf("%skill", *prefix)) {
		return
	}

	if dj[m.GuildID] == nil {
		if _, err := s.SendMessage(m.ChannelID, fmt.Sprintf("%s", noMusic), nil); err != nil {
			musicLog.Printf("Failed to send message: %s\n", err)
		}

		return
	}

	if dj[m.GuildID].Cancel != nil {
		dj[m.GuildID].Cancel()
	}

	if err := dj[m.GuildID].Leave(); err != nil {
		musicLog.Printf("Failed to leave channel: %s\n", err)
	}

	dj[m.GuildID] = nil
	if _, err := s.SendMessage(m.ChannelID, "Music session killed!", nil); err != nil {
		musicLog.Printf("Failed to send message: %s\n", err)
	}
}

func skip(m *gateway.MessageCreateEvent) {
	if !strings.HasPrefix(m.Content, fmt.Sprintf("%sskip", *prefix)) {
		return
	}

	if dj[m.GuildID] == nil {
		if _, err := s.SendMessage(m.ChannelID, fmt.Sprintf("%s", noMusic), nil); err != nil {
			musicLog.Printf("Failed to send message: %s\n", err)
		}

		return
	}

	if dj[m.GuildID].Cancel != nil {
		dj[m.GuildID].Cancel()
		dj[m.GuildID].Cancel = nil
	}
}

func queue(m *gateway.MessageCreateEvent) {
	if !strings.HasPrefix(m.Content, fmt.Sprintf("%squeue", *prefix)) {
		return
	}

	if dj[m.GuildID] == nil {
		if _, err := s.SendMessage(m.ChannelID, fmt.Sprintf("%s", noMusic), nil); err != nil {
			musicLog.Printf("Failed to send message: %s\n", err)
		}

		return
	}

	if len(dj[m.GuildID].Queue) == 0 {
		if _, err := s.SendMessage(m.ChannelID, "No songs in queue", nil); err != nil {
			musicLog.Printf("Failed to send message: %s\n", err)
		}

		return
	}

	var builder strings.Builder
	for pos, title := range dj[m.GuildID].Queue {
		builder.WriteString(fmt.Sprintf("%s: `%s`\n", strconv.Itoa(pos+1), title))
	}

	if _, err := s.SendMessage(m.ChannelID, builder.String(), nil); err != nil {
		musicLog.Printf("Failed to send message: %s\n", err)
	}
}

// start a new session.
func newSesh(m *gateway.MessageCreateEvent) {
	vs, err := s.VoiceState(m.GuildID, m.Author.ID)
	if err != nil {
		musicLog.Printf("Failed to get voice state of %s: %s\n", m.Author.Username, err)
		if _, err := s.SendMessage(m.ChannelID, fmt.Sprintf("Cannot join channel: %s not in channel", m.Author.Username), nil); err != nil {
			musicLog.Printf("Failed to send message: %s\n", err)
		}

		return
	}

	v, err := voice.NewSession(s)
	if err != nil {
		musicLog.Printf("Failed to make new voice session: %s\n", err)
		return
	}

	dj[m.GuildID] = music.New(v) // create new session for guild

	if err := dj[m.GuildID].JoinChannel(m.GuildID, vs.ChannelID, false, true); err != nil {
		musicLog.Printf("Failed to join channel: %s\n", err)
		return
	}

	// setup the queue system
	go stereo(m.GuildID, m.ChannelID)
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

		oggWriter := music.NewOggWriter(dj[gid])
		defer oggWriter.Close()
		dj[gid].Cancel = func() { cancel(); oggWriter.Close() }

		cmd.Stdin = media.Stream
		cmd.Stdout = oggWriter
		cmd.Stderr = os.Stderr

		if _, err := s.SendMessage(cid, fmt.Sprintf("Playing `%s`", media.Title), nil); err != nil {
			musicLog.Printf("Failed to send message: %s\n", err)
		}

		// this shouldnt be called concurrently....
		if err := dj[gid].Speaking(voicegateway.Microphone); err != nil {
			musicLog.Printf("Failed to start speaking: %s\n", err)
		}

		if err := cmd.Run(); err != nil {
			musicLog.Printf("Failed to run cmd: %s\n", err)
		}

		if _, err := s.SendMessage(cid, fmt.Sprintf("Finished playing `%s`", media.Title), nil); err != nil {
			musicLog.Printf("Failed to send message: %s\n", err)
		}

		if dj[gid] != nil {
			if len(dj[gid].Player) == 0 {
				if _, err := s.SendMessage(cid, "Finished queue", nil); err != nil {
					musicLog.Printf("Failed to send message: %s\n", err)
				}
			}
		}

		if dj[gid] != nil {
			dj[gid].Playing = nil
		}
	}
}
