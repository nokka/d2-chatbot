package bnetd

import (
	"io"

	"github.com/hpcloud/tail"
)

// subscriberRepository is the interface representation of the data layer.
type subscriberRepository interface {
	UpdateOnlineStatus(account string, online bool) error
}

// inmemRepository is the interface representation of the in mem data layer.
type inmemRepository interface {
	subscriberRepository
	SubscriberExists(account string) bool
}

// Watcher will listen for updates on the bnetd.log file to update subscriber online state.
type Watcher struct {
	filePath    string
	decoder     decoder
	inmem       inmemRepository
	subscribers subscriberRepository
}

// Start will start listening for updates to the file.
func (w *Watcher) Start() error {
	t, err := tail.TailFile(w.filePath, tail.Config{
		Follow: true,
		ReOpen: true,
		Location: &tail.SeekInfo{
			Offset: 0,
			Whence: io.SeekEnd,
		},
	})
	if err != nil {
		return err
	}

	// Receive lines written from the bnetd.log.
	go func(t *tail.Tail) {
		for line := range t.Lines {
			w.HandleUpdate(line.Text)
		}
	}(t)

	return nil
}

// HandleUpdate decodes login/logout entries on the bnetd.log and decides if
// the users online state has to be changed.
func (w *Watcher) HandleUpdate(data string) error {
	change, valid := w.decoder.Decode(data)

	if valid {
		exists := w.inmem.SubscriberExists(change.Account)

		// Subscriber exists so it needs to be updated.
		if exists {
			// Update persistent store with the new online state.
			err := w.subscribers.UpdateOnlineStatus(change.Account, change.Online)
			if err != nil {
				return err
			}

			// Update persisted, update the inmem store.
			w.inmem.UpdateOnlineStatus(change.Account, change.Online)
		}
	}

	return nil
}

// NewWatcher returns a new bnetd watcher with all the dependencies.
func NewWatcher(filePath string, inmem inmemRepository, subscribers subscriberRepository) *Watcher {
	return &Watcher{
		decoder:     decoder{},
		filePath:    filePath,
		inmem:       inmem,
		subscribers: subscribers,
	}
}
