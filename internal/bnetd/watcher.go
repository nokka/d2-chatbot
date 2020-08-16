package bnetd

import (
	"bufio"
	"io"
	"log"
	"os"

	"gopkg.in/fsnotify.v1"
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
	ready       bool
	decoder     decoder
	inmem       inmemRepository
	subscribers subscriberRepository
}

// Start will start listening for updates to the file.
func (w *Watcher) Start() error {
	// Open the file we're supposed to listen on.
	file, err := os.Open(w.filePath)
	if err != nil {
		return err
	}

	// Create a new file watcher using fsnotify to get updates
	// on when the file is being written to by the OS.
	fw, err := fsnotify.NewWatcher()
	if err != nil {
		return err
	}

	// Add the fsnotify watcher to the list of watchers.
	err = fw.Add(w.filePath)
	if err != nil {
		return err
	}

	// Start on another thread not to block main.
	go func(file *os.File, fw *fsnotify.Watcher) {
		// Close the file when we're done.
		defer file.Close()

		// Close the fsnotify connection when we're done.
		defer fw.Close()

		r := bufio.NewReader(file)

		for {
			line, err := r.ReadBytes('\n')
			if err != nil && err != io.EOF {
				log.Println(err)
				continue
			}

			if w.ready {
				err := w.CheckForUpdates(line)
				if err != nil {
					log.Println(err)
					continue
				}
			}

			// If error is nil we need to continue looking for data since we
			// haven't reached the end yet.
			if err == nil {
				continue
			}

			// We got EOF so let's wait for changes.
			if err = w.waitForChange(fw); err != nil {
				log.Println(err)
				continue
			}
		}
	}(file, fw)

	return nil
}

// CheckForUpdates ...
func (w *Watcher) CheckForUpdates(data []byte) error {
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

func (w *Watcher) waitForChange(fw *fsnotify.Watcher) error {
	for {
		select {
		case event := <-fw.Events:
			if event.Op&fsnotify.Write == fsnotify.Write {
				if !w.ready {
					w.ready = true
				}
				return nil
			}
		case err := <-fw.Errors:
			return err
		}
	}
}

// NewWatcher ...
func NewWatcher(filePath string, inmem inmemRepository, subscribers subscriberRepository) *Watcher {
	return &Watcher{
		decoder:     decoder{},
		filePath:    filePath,
		inmem:       inmem,
		subscribers: subscribers,
	}
}
