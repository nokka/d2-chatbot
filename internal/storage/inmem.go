package storage

import (
	"context"
	"errors"

	"github.com/nokka/d2-chatbot/internal/subscriber"
)

// InmemRepository ...
type InmemRepository struct {
	Chats map[string]map[string]subscriber.Subscriber
}

// FindSubscribers ...
func (r *InmemRepository) FindSubscribers(ctx context.Context, chatID string) ([]subscriber.Subscriber, error) {
	// Make sure chat exists.
	if chat, ok := r.Chats[chatID]; ok {
		var online []subscriber.Subscriber
		for _, sub := range chat {
			if sub.Online {
				online = append(online, sub)
			}
		}
		return online, nil
	}

	return nil, errors.New("failed to find subscribers, chat not found")
}

// Subscribe ...
func (r *InmemRepository) Subscribe(ctx context.Context, account string, chatID string) error {
	// Make sure chat exists.
	if chat, ok := r.Chats[chatID]; ok {
		// If we can't find the subscriber, add it.
		if _, ok := chat[account]; !ok {
			chat[account] = subscriber.Subscriber{
				Account: account,
				// Default to online true since a user need to be online to subscribe.
				Online: true,
			}
		}
	} else {
		return errors.New("failed to subscribe, chat id doesn't exist")
	}

	return nil
}

// Unsubscribe ...
func (r *InmemRepository) Unsubscribe(ctx context.Context, account string, chatID string) error {
	// Make sure chat exists.
	if chat, ok := r.Chats[chatID]; ok {
		delete(chat, account)
	} else {
		return errors.New("failed to unsubscribe, chat id doesn't exist")
	}

	return nil
}

// NewInmemRepository ...
func NewInmemRepository() *InmemRepository {
	return &InmemRepository{
		Chats: map[string]map[string]subscriber.Subscriber{
			"hc": make(map[string]subscriber.Subscriber, 0),
		},
	}
}
