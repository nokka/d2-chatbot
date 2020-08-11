package client

import (
	"context"
	"fmt"
	"log"
	"sync"

	"github.com/nokka/d2-chatbot/internal/subscriber"
	"github.com/nokka/d2client"
)

// subscriberRepository is the interface representation of the data layer
// the service depend on.
type subscriberRepository interface {
	FindSubscribers(ctx context.Context, chatID string) ([]subscriber.Subscriber, error)
	Subscribe(ctx context.Context, account string, chatID string) error
	Unsubscribe(ctx context.Context, account string, chatID string) error
}

// Client wraps the connection to the d2 server and is responsible for communication.
type Client struct {
	addr        string
	username    string
	password    string
	decoder     decoder
	conn        d2client.Client
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
	err = client.Login(c.username, c.password)
	if err != nil {
		return err
	}

	// Add the tcp connection to our client.
	c.conn = client

	// Listen for data on the connection indefinitely.
	go c.listenAndClose()

	return nil
}

// Subscribe ...
func (c *Client) Subscribe(message *Message) error {
	err := c.subscribers.Subscribe(context.Background(), message.Account, message.ChatID)
	if err != nil {
		return err
	}

	// Notify subscriber.
	c.conn.Whisper(message.Account, fmt.Sprintf("subscribed to %s", message.ChatID))

	return nil
}

// Publish ...
func (c *Client) Publish(message *Message) error {
	// Lock to publish in order to preserve message order integrity.
	c.publishLock.Lock()

	// Unlock when we're done.
	defer c.publishLock.Unlock()

	subscribers, err := c.subscribers.FindSubscribers(context.Background(), message.ChatID)
	if err != nil {
		return err
	}

	for _, sub := range subscribers {
		if sub.Account == message.Account {
			continue
		}

		err := c.conn.Whisper(sub.Account, message.Message)
		if err != nil {
			log.Println("failed to delivery message", err)
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
					c.Subscribe(decoded)
				case TypeUnsubscribe:
					log.Println("unsubscribe", decoded)
				case TypePublish:
					go c.Publish(decoded)
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
func New(addr string, username string, password string, subscribers subscriberRepository) *Client {
	return &Client{
		addr:        addr,
		username:    username,
		password:    password,
		decoder:     decoder{},
		subscribers: subscribers,
	}
}
