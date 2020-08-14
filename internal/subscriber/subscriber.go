package subscriber

import "time"

// Subscriber ...
type Subscriber struct {
	Account     string
	Online      bool
	BannedUntil *time.Time
}
