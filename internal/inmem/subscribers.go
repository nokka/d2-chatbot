package inmem

import (
	"errors"
	"sync"
	"time"

	"github.com/nokka/d2-chatbot/internal/subscriber"
)

// SubscriberRepository ...
type SubscriberRepository struct {
	Chats map[string]map[string]subscriber.Subscriber
	rwm   sync.RWMutex
}

// Sync ...
func (r *SubscriberRepository) Sync(chatID string, subscribers []subscriber.Subscriber) error {
	r.rwm.RLock()
	defer r.rwm.RUnlock()

	// Make sure chat exists.
	if _, ok := r.Chats[chatID]; ok {
		for i, sub := range subscribers {
			r.Chats[chatID][sub.Account] = subscribers[i]
		}

		return nil
	}

	return errors.New("unable to sync, chat not found")
}

// FindSubscriber ...
func (r *SubscriberRepository) FindSubscriber(account string, chatID string) *subscriber.Subscriber {
	r.rwm.RLock()
	defer r.rwm.RUnlock()

	// Make sure chat exists.
	if chat, ok := r.Chats[chatID]; ok {
		for _, sub := range chat {
			if sub.Account == account {
				return &sub
			}
		}
	}

	return nil
}

// FindSubscribers ...
func (r *SubscriberRepository) FindSubscribers(chatID string) ([]subscriber.Subscriber, error) {
	r.rwm.RLock()
	defer r.rwm.RUnlock()

	// Make sure chat exists.
	if chat, ok := r.Chats[chatID]; ok {
		var subs []subscriber.Subscriber
		for _, sub := range chat {
			subs = append(subs, sub)
		}
		return subs, nil
	}

	return nil, errors.New("failed to find subscribers, chat not found")
}

// FindEligibleSubscribers ...
func (r *SubscriberRepository) FindEligibleSubscribers(chatID string) ([]subscriber.Subscriber, error) {
	r.rwm.RLock()
	defer r.rwm.RUnlock()

	// Make sure chat exists.
	if chat, ok := r.Chats[chatID]; ok {
		var subs []subscriber.Subscriber
		for _, sub := range chat {
			// The subscriber is eligible for messages if they're both online and not currently banned.
			if sub.Online && (sub.BannedUntil == nil || sub.BannedUntil.Before(time.Now())) {
				subs = append(subs, sub)
			}
		}
		return subs, nil
	}

	return nil, errors.New("failed to find online subscribers, chat not found")
}

// Subscribe ...
func (r *SubscriberRepository) Subscribe(account string, chatID string) error {
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
func (r *SubscriberRepository) Unsubscribe(account string, chatID string) error {
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

// UpdateOnlineStatus ...
func (r *SubscriberRepository) UpdateOnlineStatus(account string, online bool) error {
	r.rwm.RLock()
	defer r.rwm.RUnlock()

	// Search through all chats to find the subscriber in any of them.
	for _, chat := range r.Chats {
		if subscriber, ok := chat[account]; ok {
			subscriber.Online = online
		}
	}

	return nil
}

// SubscriberExists ...
func (r *SubscriberRepository) SubscriberExists(account string) bool {
	r.rwm.RLock()
	defer r.rwm.RUnlock()

	// Search through all chats to find the subscriber in any of them.
	for _, chat := range r.Chats {
		if _, ok := chat[account]; ok {
			return true
		}
	}

	return false
}

// NewSubscriberRepository ...
func NewSubscriberRepository() *SubscriberRepository {
	return &SubscriberRepository{
		Chats: map[string]map[string]subscriber.Subscriber{
			"chat":  make(map[string]subscriber.Subscriber, 0),
			"trade": make(map[string]subscriber.Subscriber, 0),
			"hc":    make(map[string]subscriber.Subscriber, 0),
		},
	}
}
