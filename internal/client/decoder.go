package client

import (
	"fmt"
	"regexp"
)

// Compile the regex once.
var r = regexp.MustCompile(`^<from\s+([a-z0-9_\-]+)>\s+([a-z]{3})\s+([a-z]{2})\s*([a-z]+)?`)

// Decoder ...
type decoder struct{}

// Allowed message types.
const (
	TypeSubscribe   = "sub"
	TypeUnsubscribe = "uns"
	TypePublish     = "pub"
	account         = 1
	cmd             = 2
	chatID          = 3
	msg             = 4
)

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

	return &Message{
		ChatID:  matches[chatID],
		Account: matches[account],
		Cmd:     matches[cmd],
		Message: fmt.Sprintf("[%s] %s", matches[account], matches[msg]),
	}, true
}
