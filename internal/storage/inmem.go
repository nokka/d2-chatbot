package storage

import (
	"context"
	"errors"

	"github.com/nokka/d2-chatbot/internal/subscriber"
)

// DB ...
type DB struct {
	Subscribers map[string]subscriber.Subscriber
}

// InmemRepository ...
type InmemRepository struct {
	DB DB
}

// Find ...
func (r *InmemRepository) Find(ctx context.Context, account string) (*subscriber.Subscriber, error) {
	if subscriber, ok := r.DB.Subscribers[account]; ok {
		return &subscriber, nil
	}

	return nil, errors.New("subscriber not found")
}

// Subscribe ...
func (r *InmemRepository) Subscribe(ctx context.Context, account string, chatID string) error {
	// Subscriber already exists.
	if subscriber, ok := r.DB.Subscribers[account]; ok {
		var subscribed bool
		for _, v := range subscriber.Chats {
			if v == chatID {
				subscribed = true
			}
		}

		// Wasn't subscribed, append chat id to list.
		if !subscribed {
			subscriber.Chats = append(subscriber.Chats, chatID)
		}

		return nil
	}

	// Subscriber not in map, add it.
	r.DB.Subscribers[account] = subscriber.Subscriber{
		Account: account,
		// Default to online true since a user need to be online to subscribe.
		Online: true,
		Chats:  []string{chatID},
	}

	return nil
}

// Unsubscribe ...
func (r *InmemRepository) Unsubscribe(ctx context.Context, account string, chatID string) error {
	if subscriber, ok := r.DB.Subscribers[account]; ok {
		for k, v := range subscriber.Chats {
			if v == chatID {
				subscriber.Chats[k] = subscriber.Chats[len(subscriber.Chats)-1]
				subscriber.Chats = subscriber.Chats[:len(subscriber.Chats)-1]
			}
		}
	}

	// Subscriber not found, nothing to do.
	return nil
}

// NewInmemRepository ...
func NewInmemRepository() *InmemRepository {
	return &InmemRepository{}
}
