package client

import (
	"fmt"
	"regexp"
)

// Compile the regex once.
var r = regexp.MustCompile(`(?i)^<from\s+([a-z0-9_\-]+)>\s+([#@!]{1})\s*([a-z .:,_\-'"!?]+)?`)

// Decoder ...
type decoder struct{}

// Allowed message types.
const (
	// Commands.
	TypeSubscribe   = "@"
	TypeUnsubscribe = "!"
	TypePublish     = "#"

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
}

// Message ...
type Message struct {
	Account string
	Cmd     string
	Message string
}

// Decode ...
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
		Account: matches[account],
		Cmd:     matches[cmd],
	}

	if message.Cmd == TypePublish {
		message.Message = fmt.Sprintf("[%s] %s", matches[account], matches[msg])
	}

	return message, true
}
