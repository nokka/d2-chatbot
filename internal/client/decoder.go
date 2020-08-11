package client

import (
	"regexp"
)

// Compile the regex once.
var r = regexp.MustCompile(`^<from\s+([a-z0-9]+)>\s+([a-z]{3})\s+([a-z]{2})\s*([a-z]+)?`)

// Decoder ...
type decoder struct{}

// Allowed message types.
const (
	TypeSubscribe = "sub"
	TypePublish   = "pub"
	account       = 1
	cmd           = 2
	id            = 3
	msg           = 4
)

// Message ...
type Message struct {
	ID      string
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
		ID:      matches[id],
		Account: matches[account],
		Cmd:     matches[cmd],
		Message: matches[msg],
	}, true
}
