package client

import (
	"fmt"
	"log"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/nokka/d2-chatbot/internal/subscriber"
	"github.com/nokka/d2client"
)

// subscriberRepository is the interface representation of the data layer.
type subscriberRepository interface {
	FindSubscribers(chatID string) ([]subscriber.Subscriber, error)
	FindEligibleSubscribers(chatID string) ([]subscriber.Subscriber, error)
	Subscribe(account string, chatID string) error
	Unsubscribe(account string, chatID string) error
	UpdateBannedUntil(account string, chatID string, until *time.Time) error
	FindModerators() ([]string, error)
}

// inmemRepository is the interface representation of the in mem data layer.
type inmemRepository interface {
	subscriberRepository
	SyncSubscribers(chatID string, subscribers []subscriber.Subscriber) error
	FindSubscriber(account string, chatID string) *subscriber.Subscriber
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

	err = c.inmem.SyncSubscribers(c.chatID, subscribers)
	if err != nil {
		return err
	}

	return nil
}

// Subscribe ...
func (c *Client) Subscribe(message *Message) error {
	// Check in memory store if the account is subscribed.
	sub := c.inmem.FindSubscriber(message.Account, c.chatID)

	// Cancel the operation if subscriber is banned.
	if sub != nil && c.subscriberBanned(*sub) {
		return nil
	}

	// Subscriber didn't exist, persist them.
	if sub == nil {
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

		// Notify subscriber that they have been successfully subscribed.
		c.conn.Whisper(message.Account, fmt.Sprintf("[subscribed %s]", c.chatID))

		return nil
	}

	// Notify subscriber that they are already subscribed.
	c.conn.Whisper(message.Account, fmt.Sprintf("[already subscribed to %s] ", c.chatID))

	return nil
}

// Unsubscribe ...
func (c *Client) Unsubscribe(message *Message) error {
	// Check in memory store if the account is subscribed.
	sub := c.inmem.FindSubscriber(message.Account, c.chatID)
	if sub == nil {
		c.conn.Whisper(message.Account, fmt.Sprintf("[not subscribed to %s]", c.chatID))
		return nil
	}

	// Cancel the operation if subscriber is banned.
	if banned := c.subscriberBanned(*sub); banned {
		return nil
	}

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
	c.conn.Whisper(message.Account, fmt.Sprintf("[unsubscribed %s]", c.chatID))

	return nil
}

// Publish ...
func (c *Client) Publish(message *Message) error {
	// Lock to publish in order to preserve message order integrity.
	c.publishLock.Lock()

	// Unlock when we're done.
	defer c.publishLock.Unlock()

	// Check in memory store if the account is subscribed to the chat.
	sub := c.inmem.FindSubscriber(message.Account, c.chatID)
	if sub == nil {
		c.conn.Whisper(message.Account, fmt.Sprintf("[not subscribed to %s]", c.chatID))
		return nil
	}

	// Cancel the operation if subscriber is banned.
	if banned := c.subscriberBanned(*sub); banned {
		return nil
	}

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
		// If there's an error, log it and continue with the next message.
		if err != nil {
			log.Println("failed to deliver message", err)
		}
	}

	return nil
}

// Ban ...
func (c *Client) Ban(message *Message) error {
	mods, err := c.inmem.FindModerators()
	if err != nil {
		return err
	}

	var allowed bool
	for _, mod := range mods {
		if message.Account == mod {
			allowed = true
			break
		}
	}

	if !allowed {
		c.conn.Whisper(message.Account, "[insufficient privileges]")
		return nil
	}

	// Extract account to ban and days to ban from message.
	parts := strings.Split(message.Message, " ")

	if len(parts) < 2 {
		return fmt.Errorf("failed to extract data when banning, message: %s", message.Message)
	}

	account := parts[0]
	days, err := strconv.Atoi(parts[1])
	if err != nil {
		return err
	}

	until := time.Now().AddDate(0, 0, days)

	// Check in memory store if the account is subscribed to the chat.
	sub := c.inmem.FindSubscriber(account, c.chatID)
	if sub == nil {
		c.conn.Whisper(message.Account, fmt.Sprintf("[%s not subscribed to %s]", account, c.chatID))
		return nil
	}

	// Subscriber exists, ban them.
	err = c.subscribers.UpdateBannedUntil(account, c.chatID, &until)
	if err != nil {
		return err
	}

	// Ban persisted, update inmem store.
	err = c.inmem.UpdateBannedUntil(account, c.chatID, &until)
	if err != nil {
		return err
	}

	// Notify moderator that the ban was complete.
	c.conn.Whisper(message.Account, fmt.Sprintf("[%s has been banned from %s until %v]", account, c.chatID, until))

	// Notify subscriber that they have been banned.
	c.conn.Whisper(account, fmt.Sprintf("[you have been banned from %s until %v]", c.chatID, until))

	return nil
}

func (c *Client) subscriberBanned(sub subscriber.Subscriber) bool {
	if sub.BannedUntil == nil {
		return false
	}

	if sub.BannedUntil.After(time.Now()) {
		// Calculate days left on ban.
		remainder := sub.BannedUntil.Sub(time.Now())
		days := int(remainder.Hours() / 24)

		if days >= 1 {
			c.conn.Whisper(sub.Account, fmt.Sprintf("[you are banned on %s for %d more days]", c.chatID, days))
		} else {
			c.conn.Whisper(sub.Account, fmt.Sprintf("[you are banned on %s for %d more hours]", c.chatID, int(remainder.Hours())))
		}

		return true
	}

	return false
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
				case TypeBan:
					err := c.Ban(decoded)
					if err != nil {
						log.Printf("failed to ban %s", err)
					}
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
