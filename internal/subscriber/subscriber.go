package subscriber

import "time"

// Subscriber is the heart of the domain, a subscriber
// represents an account and it's current state.
type Subscriber struct {
	Account     string
	Online      bool
	BannedUntil *time.Time
}
