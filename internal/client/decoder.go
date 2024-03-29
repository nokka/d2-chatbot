package client

import (
	"fmt"
	"regexp"
	"strings"
)

// Compile the regex once.
var r = regexp.MustCompile(`(?i)^<from\s+([a-z0-9_\-]+)>\s+([#@!\~]{1})\s*(.+)?`)

// IP address regex to remove sensitive information when replying.
var ipregx = regexp.MustCompile(`(25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)(\.(25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)){3}`)

// Decoder will decode incoming messages.
type decoder struct{}

// Allowed message types.
const (
	// Commands.
	TypeSubscribe   = "@"
	TypeUnsubscribe = "!"
	TypePublish     = "#"
	TypeBan         = "~"

	// Indices.
	account = 1
	cmd     = 2
	msg     = 3
)

// Allowed commands.
var cmds = map[string]struct{}{
	TypeSubscribe:   {},
	TypePublish:     {},
	TypeUnsubscribe: {},
	TypeBan:         {},
}

// Message is the message decoded.
type Message struct {
	Account string
	Cmd     string
	Message string
}

// Decode will decode incoming message string and validate it.
func (d decoder) Decode(data []byte) (*Message, bool) {
	matches := r.FindStringSubmatch(string(data))

	if len(matches) != 4 {
		return nil, false
	}

	// Return invalid if the cmd isn't allowed.
	if _, ok := cmds[matches[cmd]]; !ok {
		return nil, false
	}

	message := &Message{
		Account: strings.ToLower(matches[account]),
		Cmd:     matches[cmd],
	}

	// Clean up message from IP address that can accidentally
	// get appended by bnalias commands such as '%r'.
	processed := ipregx.ReplaceAllString(matches[msg], "")

	switch message.Cmd {
	case TypePublish:
		message.Message = fmt.Sprintf("[%s] %s", matches[account], processed)
	case TypeBan:
		message.Message = matches[msg]
	}

	return message, true
}
