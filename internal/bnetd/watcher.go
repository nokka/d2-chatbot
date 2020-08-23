package bnetd

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"time"

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
	logFile     *os.File
	fw          *fsnotify.Watcher
	reader      *bufio.Reader
	decoder     decoder
	inmem       inmemRepository
	subscribers subscriberRepository
}

// Start will start listening for updates to the file.
func (w *Watcher) Start() (<-chan error, error) {
	// Create a new file watcher using fsnotify to get updates
	// on when the file is being written to by the OS.
	fw, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, err
	}

	w.fw = fw

	err = w.rotateLog()
	if err != nil {
		return nil, err
	}

	errorChan := make(chan error)

	// Start on another thread not to block main.
	go func() {
		// Close when we're done.
		defer w.fw.Close()
		defer w.logFile.Close()

		for {
			line, err := w.reader.ReadBytes('\n')
			if err != nil && err != io.EOF {
				errorChan <- err
				continue
			}

			if w.ready {
				err := w.CheckForUpdates(line)
				if err != nil {
					errorChan <- err
					continue
				}
			}

			// If error is nil we need to continue looking for data since we
			// haven't reached the end yet.
			if err == nil {
				continue
			}

			// We got EOF so let's wait for changes.
			if err = w.waitForChange(w.fw); err != nil {
				errorChan <- err
				continue
			}
		}
	}()

	return errorChan, nil
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

func (w *Watcher) rotateLog() error {
	// Make sure the watcher won't update state until log has been rotated.
	w.ready = false

	// Remove the file listener since the descriptor will be gone.
	w.fw.Remove(w.filePath)

	// Reset log file.
	if w.logFile != nil {
		err := w.logFile.Close()
		if err != nil {
			return err
		}
		w.logFile = nil
	}

	// Try to get a new file descriptor until we succeed or it errors.
	for w.logFile == nil {
		time.Sleep(500 * time.Millisecond)
		fmt.Println("attempting to get new file descriptor for bnetd log")

		file, err := os.Open(w.filePath)
		if err != nil && os.IsNotExist(err) {
			continue
		}

		if err != nil {
			return err
		}

		w.logFile = file
	}

	// Update reader with the new file descriptor.
	w.reader = bufio.NewReader(w.logFile)

	// Add the file path to the fsnotify watcher.
	err := w.fw.Add(w.filePath)
	if err != nil {
		return err
	}

	fmt.Println("bnetd log has been successfully rotated", w.filePath)

	return nil
}

func (w *Watcher) waitForChange(fw *fsnotify.Watcher) error {
	for {
		select {
		case event := <-fw.Events:
			if event.Op&fsnotify.Write == fsnotify.Write {
				fmt.Println("WRITE HAPPENED")
				if !w.ready {
					w.ready = true
				}
				return nil
			}
			// Rename happens when the log file rotates and the file
			// is removed. This means we have to get a new file descriptor.
			if event.Op&fsnotify.Rename == fsnotify.Rename {
				fmt.Println("RENAME HAPPENED")
				err := w.rotateLog()
				if err != nil {
					return err
				}
				return nil
			}
			if event.Op&fsnotify.Chmod == fsnotify.Chmod {
				fmt.Println("CHMOD HAPPENED")
			}
			if event.Op&fsnotify.Create == fsnotify.Create {
				fmt.Println("CREATE HAPPENED")
			}
			if event.Op&fsnotify.Remove == fsnotify.Remove {
				fmt.Println("REMOVE HAPPENED")
			}

		case err := <-fw.Errors:
			return err
		}
	}
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
