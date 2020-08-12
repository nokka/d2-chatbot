package inmem

import (
	"context"
	"errors"
	"sync"

	"github.com/nokka/d2-chatbot/internal/subscriber"
)

// Repository ...
type Repository struct {
	Chats map[string]map[string]subscriber.Subscriber
	rwm   sync.RWMutex
}

// FindSubscribers ...
func (r *Repository) FindSubscribers(ctx context.Context, chatID string) ([]subscriber.Subscriber, error) {
	r.rwm.RLock()
	defer r.rwm.RUnlock()

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
func (r *Repository) Subscribe(ctx context.Context, account string, chatID string) error {
	r.rwm.RLock()
	defer r.rwm.RUnlock()

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
func (r *Repository) Unsubscribe(ctx context.Context, account string, chatID string) error {
	r.rwm.RLock()
	defer r.rwm.RUnlock()

	// Make sure chat exists.
	if chat, ok := r.Chats[chatID]; ok {
		delete(chat, account)
	} else {
		return errors.New("failed to unsubscribe, chat id doesn't exist")
	}

	return nil
}

// NewRepository ...
func NewRepository() *Repository {
	return &Repository{
		Chats: map[string]map[string]subscriber.Subscriber{
			"chat":  make(map[string]subscriber.Subscriber, 0),
			"trade": make(map[string]subscriber.Subscriber, 0),
			"hc":    make(map[string]subscriber.Subscriber, 0),
		},
	}
}
