package bnetd

import (
	"regexp"
	"strings"
)

// Compile the regex once.
var r = regexp.MustCompile(`(?i)^.*\s"([a-z0-9_\-]+)"\s([a-z]+\s[a-z]{2,3})`)

// Decoder ...
type decoder struct{}

// StatusChange ...
type StatusChange struct {
	Account string
	Online  bool
}

const (
	// Indices.
	account = 1
	action  = 2
)

// Decode ...
func (d decoder) Decode(data string) (*StatusChange, bool) {
	matches := r.FindStringSubmatch(data)

	if len(matches) != 3 {
		return nil, false
	}

	var online bool
	switch matches[action] {
	case "logged in":
		online = true
	case "logged out":
		online = false
	default:
		return nil, false
	}

	change := &StatusChange{
		Account: strings.ToLower(matches[account]),
		Online:  online,
	}

	return change, true
}
