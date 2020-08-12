package client

import (
	"fmt"
	"regexp"
)

// Compile the regex once.
var r = regexp.MustCompile(`(?i)^<from\s+([a-z0-9_\-]+)>\s+([a-z]{3})\s+([a-z]{2})\s*([a-z .,_\-'"!?]+)?`)

// Decoder ...
type decoder struct{}

// Allowed message types.
const (
	// Commands.
	TypeSubscribe   = "sub"
	TypeUnsubscribe = "uns"
	TypePublish     = "pub"

	// Chat IDs.
	ChatHC    = "hc"
	ChatSC    = "sc"
	ChatTrade = "tr"

	// Indices.
	account = 1
	cmd     = 2
	chatID  = 3
	msg     = 4
)

// Allowed commands.
var cmds = map[string]struct{}{
	TypeSubscribe:   {},
	TypePublish:     {},
	TypeUnsubscribe: {},
}

// Allowed chat ids.
var chatIDs = map[string]struct{}{
	ChatHC:    {},
	ChatSC:    {},
	ChatTrade: {},
}

// Message ...
type Message struct {
	ChatID  string
	Account string
	Cmd     string
	Message string
}

// Decode ...
func (d decoder) Decode(data []byte) (*Message, bool) {
	matches := r.FindStringSubmatch(string(data))

	if len(matches) != 5 {
		return nil, false
	}

	// Return invalid if the cmd isn't allowed.
	if _, ok := cmds[matches[cmd]]; !ok {
		return nil, false
	}

	// Return invalid if the chat id isn't allowed.
	if _, ok := chatIDs[matches[chatID]]; !ok {
		return nil, false
	}

	message := &Message{
		ChatID:  matches[chatID],
		Account: matches[account],
		Cmd:     matches[cmd],
	}

	if message.Cmd == TypePublish {
		message.Message = fmt.Sprintf("[%s] %s", matches[account], matches[msg])
	}

	return message, true
}
