package music

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"unicode/utf8"

	"github.com/diamondburned/arikawa/v2/discord"
	"github.com/diamondburned/arikawa/v2/voice"
	"github.com/jonas747/ogg"
	"github.com/kkdai/youtube/v2"
)

// Maximum queue length
var MaxQueueLength = 512

// base youtube link
var (
	scheme      = "https"
	base        = "www.youtube.com"
	search      = "results?search_query="
	YtSearchURL = scheme + "://" + base + "/" + search
)

// youtube client
var uc = youtube.Client{}

var nonYoutubeLink = errors.New("Not a youtube link")

// handle media
// potentially expandable to other platforms
type Media struct {
	*youtube.Video
	Stream  io.Reader
	StartAt int // used for rebasing
}

// a music session
type Sesh struct {
	*voice.Session
	Player chan Media
	Cancel func()

	Playing *Media
	Queue   []string
}

// returns a new session
func New(v *voice.Session) *Sesh {
	return &Sesh{Session: v, Player: make(chan Media, MaxQueueLength)}
}

// DJ makes it easier to run multiple
// voice.Sessions at once
type DJ map[discord.GuildID]*Sesh

// gets top search result's video ID from youtube
func GetID(str string) (string, error) {
	str = strings.ReplaceAll(str, " ", "+")

	resp, err := http.Get(YtSearchURL + str)
	if err != nil {
		return "", err
	} else if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("http status not ok: %s", resp.StatusCode)
	}
	defer resp.Body.Close()

	scanner := bufio.NewScanner(resp.Body)
	buf := make([]byte, 1024)
	scanner.Buffer(buf, 512*1024)
	scanner.Split(scanQuotes)

	var passes int
	for scanner.Scan() {
		if scanner.Text() == "videoId" {
			passes++
			if passes == 2 {
				scanner.Scan()
				scanner.Scan()
				return scanner.Text(), nil
			}
		}
	}

	return "", errors.New("Error! Could not find video id")
}

// returns media of given id
func GetVideo(id string) (Media, error) {
	video, err := uc.GetVideo(id)
	if err != nil {
		return Media{}, err
	}

	resp, err := uc.GetStream(video, &video.Formats[0])
	if err != nil {
		return Media{}, err
	}

	return Media{Stream: resp.Body, Video: video}, nil
}

// bufio.Split function
// token between two '"'
func scanQuotes(data []byte, atEOF bool) (advance int, token []byte, err error) {
	// Skip leading qoutes.
	start := 0
	for width := 0; start < len(data); start += width {
		var r rune
		r, width = utf8.DecodeRune(data[start:])
		if r != '"' {
			break
		}
	}

	// Scan until qoutes, marking end of word.
	for width, i := 0, start; i < len(data); i += width {
		var r rune
		r, width = utf8.DecodeRune(data[i:])
		if r == '"' {
			return i + width, data[start:i], nil
		}
	}

	// If we're at EOF, we have a final, non-empty, non-terminated word. Return it.
	if atEOF && len(data) > start {
		return len(data), data[start:], nil
	}

	// Request more data.
	return start, nil, nil
}

// OggWriter is used to play sound through voice.
type OggWriter struct {
	pr    *io.PipeReader
	pw    *io.PipeWriter
	errCh chan error
}

// returns a new OggWriter
func NewOggWriter(w io.Writer) *OggWriter {
	pr, pw := io.Pipe()
	errCh := make(chan error, 1)

	go func() {
		oggDec := ogg.NewPacketDecoder(ogg.NewDecoder(pr))
		for {
			packet, _, err := oggDec.Decode()
			if err != nil {
				errCh <- err
				break
			}
			if _, err := w.Write(packet); err != nil {
				errCh <- err
				break
			}
		}
	}()

	return &OggWriter{
		pw:    pw,
		pr:    pr,
		errCh: errCh,
	}
}

// Write to an OggWriter
func (w *OggWriter) Write(b []byte) (int, error) {
	select {
	case err := <-w.errCh:
		return 0, err
	default:
		return w.pw.Write(b)
	}
}

// Close an OggWriter
func (w *OggWriter) Close() error {
	w.pw.Close()
	return w.pr.Close()
}
