package client

import (
	"fmt"
	"log"
	"sync"

	"github.com/nokka/d2-chatbot/internal/subscriber"
	"github.com/nokka/d2client"
)

// inmemRepository is the interface representation of the in mem data layer.
type inmemRepository interface {
	subscriberRepository
	Sync(chatID string, subscribers []subscriber.Subscriber) error
}

// subscriberRepository is the interface representation of the data layer.
type subscriberRepository interface {
	FindSubscribers(chatID string) ([]subscriber.Subscriber, error)
	FindEligibleSubscribers(chatID string) ([]subscriber.Subscriber, error)
	Subscribe(account string, chatID string) error
	Unsubscribe(account string, chatID string) error
}

// Client wraps the connection to the d2 server and is responsible for communication.
type Client struct {
	addr        string
	chatID      string
	password    string
	decoder     decoder
	conn        d2client.Client
	inmem       inmemRepository
	subscribers subscriberRepository
	publishLock sync.Mutex
}

// Open will open a tcp connection to the d2 server.
func (c *Client) Open() error {
	// Create a new d2 tcp client.
	client := d2client.New()

	// Open connection over tcp.
	err := client.Open(c.addr)
	if err != nil {
		return err
	}

	// Login with the username and password.
	err = client.Login(c.chatID, c.password)
	if err != nil {
		return err
	}

	// Add the tcp connection to our client.
	c.conn = client

	// Listen for data on the connection indefinitely.
	go c.listenAndClose()

	return nil
}

// Sync ...
func (c *Client) Sync() error {
	subscribers, err := c.subscribers.FindSubscribers(c.chatID)
	if err != nil {
		return err
	}

	err = c.inmem.Sync(c.chatID, subscribers)
	if err != nil {
		return err
	}

	return nil
}

// Subscribe ...
func (c *Client) Subscribe(message *Message) error {
	// Subscribe to persistent store first.
	err := c.subscribers.Subscribe(message.Account, c.chatID)
	if err != nil {
		return err
	}

	// Subscription persisted, add to in memory db.
	err = c.inmem.Subscribe(message.Account, c.chatID)
	if err != nil {
		return err
	}

	// Notify subscriber.
	c.conn.Whisper(message.Account, fmt.Sprintf("[subscribed] %s", c.chatID))

	return nil
}

// Unsubscribe ...
func (c *Client) Unsubscribe(message *Message) error {
	// Unsubscribe to persistent store first.
	err := c.subscribers.Unsubscribe(message.Account, c.chatID)
	if err != nil {
		return err
	}

	// Unubscription persisted, remove it from in memory db.
	err = c.inmem.Unsubscribe(message.Account, c.chatID)
	if err != nil {
		return err
	}

	// Notify subscriber.
	c.conn.Whisper(message.Account, fmt.Sprintf("[unsubscribed] %s", c.chatID))

	return nil
}

// Publish ...
func (c *Client) Publish(message *Message) error {
	// Lock to publish in order to preserve message order integrity.
	c.publishLock.Lock()

	// Unlock when we're done.
	defer c.publishLock.Unlock()

	subscribers, err := c.inmem.FindEligibleSubscribers(c.chatID)
	if err != nil {
		return err
	}

	for _, s := range subscribers {
		// Localize scope.
		sub := s

		if sub.Account == message.Account {
			continue
		}

		err := c.conn.Whisper(sub.Account, message.Message)
		if err != nil {
			log.Println("failed to deliver message", err)
		}
	}

	return nil
}

func (c *Client) listenAndClose() {
	// Setup channel to read on.
	ch := make(chan []byte)

	// Setup output error channel.
	errors := make(chan error)

	c.conn.Read(ch, errors)

	// Promise to close the connection when we're done.
	defer c.conn.Close()

	// Read the output from the chat onto a channel.
	for {
		select {
		// This case means we recieved data on the connection.
		case data := <-ch:
			if decoded, valid := c.decoder.Decode(data); valid {
				switch decoded.Cmd {
				case TypeSubscribe:
					err := c.Subscribe(decoded)
					if err != nil {
						log.Printf("failed to subscribe %s", err)
					}
				case TypeUnsubscribe:
					err := c.Unsubscribe(decoded)
					if err != nil {
						log.Printf("failed to unsubscribe %s", err)
					}
				case TypePublish:
					// Publish on a separate thread.
					go func() {
						err := c.Publish(decoded)
						if err != nil {
							log.Printf("failed to publish %s", err)
						}
					}()
				default:
					log.Printf("unknown cmd received: %s", decoded.Cmd)
				}
			}

		case err := <-errors:
			log.Println("got error while listening on client output", err)
			break
		}
	}
}

// New ...
func New(addr string, chatID string, password string, inmem inmemRepository, subscribers subscriberRepository) *Client {
	return &Client{
		addr:        addr,
		chatID:      chatID,
		password:    password,
		decoder:     decoder{},
		inmem:       inmem,
		subscribers: subscribers,
	}
}
