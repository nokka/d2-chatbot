package inmem

import (
	"errors"
	"sync"
	"time"

	"github.com/nokka/d2-chatbot/internal/subscriber"
)

// SubscriberRepository is an in memory repository used to store subscribers.
type SubscriberRepository struct {
	Chats      map[string]map[string]subscriber.Subscriber
	Moderators []string
	rwm        sync.RWMutex
}

// SyncSubscribers syncs the given subscribers to memory.
func (r *SubscriberRepository) SyncSubscribers(chatID string, subscribers []subscriber.Subscriber) error {
	r.rwm.Lock()
	defer r.rwm.Unlock()

	// Make sure chat exists.
	if _, ok := r.Chats[chatID]; ok {
		for i, sub := range subscribers {
			r.Chats[chatID][sub.Account] = subscribers[i]
		}

		return nil
	}

	return errors.New("unable to sync, chat not found")
}

// SyncModerators syncs the given moderator names to memory.
func (r *SubscriberRepository) SyncModerators(moderators []string) {
	r.rwm.Lock()
	defer r.rwm.Unlock()

	r.Moderators = moderators
}

// FindSubscriber looks through the in memory map to find a subscriber on the given chat.
func (r *SubscriberRepository) FindSubscriber(account string, chatID string) *subscriber.Subscriber {
	r.rwm.RLock()
	defer r.rwm.RUnlock()

	// Make sure chat exists.
	if chat, ok := r.Chats[chatID]; ok {
		if sub, exists := chat[account]; exists {
			return &sub
		}
	}

	return nil
}

// FindSubscribers looks through memory to find all subscribers on the given chat id.
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

// FindEligibleSubscribers finds subscribers that are eligible to receive chat messages on a chat.
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

// Subscribe adds a new account to the given chat.
func (r *SubscriberRepository) Subscribe(account string, chatID string) error {
	r.rwm.Lock()
	defer r.rwm.Unlock()

	// Make sure chat exists.
	if chat, ok := r.Chats[chatID]; ok {
		// If we can't find the subscriber, add it.
		if _, ok := chat[account]; !ok {
			r.Chats[chatID][account] = subscriber.Subscriber{
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

// Unsubscribe removes the account from the given chat.
func (r *SubscriberRepository) Unsubscribe(account string, chatID string) error {
	r.rwm.Lock()
	defer r.rwm.Unlock()

	// Make sure chat exists.
	if chat, ok := r.Chats[chatID]; ok {
		delete(chat, account)
	} else {
		return errors.New("failed to unsubscribe, chat id doesn't exist")
	}

	return nil
}

// UpdateOnlineStatus updates the status of an account in memory.
func (r *SubscriberRepository) UpdateOnlineStatus(account string, online bool) error {
	r.rwm.Lock()
	defer r.rwm.Unlock()

	// Search through all chats to find the subscriber in any of them.
	for id, chat := range r.Chats {
		if subscriber, ok := chat[account]; ok {
			subscriber.Online = online
			r.Chats[id][account] = subscriber
		}
	}

	return nil
}

// SubscriberExists checks if a given account exists.
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

// UpdateBannedUntil updates the ban time.
func (r *SubscriberRepository) UpdateBannedUntil(account string, chatID string, until *time.Time) error {
	r.rwm.Lock()
	defer r.rwm.Unlock()

	// Make sure chat exists.
	if chat, ok := r.Chats[chatID]; ok {
		// Make sure subscriber exists.
		if subscriber, ok := chat[account]; ok {
			subscriber.BannedUntil = until
			r.Chats[chatID][account] = subscriber
		}
	} else {
		return errors.New("failed to ban subscriber, chat id doesn't exist")
	}

	return nil
}

// FindModerators finds all moderators.
func (r *SubscriberRepository) FindModerators() ([]string, error) {
	r.rwm.RLock()
	defer r.rwm.RUnlock()
	return r.Moderators, nil
}

// NewSubscriberRepository returns a repository with all dependencies set up.
func NewSubscriberRepository() *SubscriberRepository {
	return &SubscriberRepository{
		Chats: map[string]map[string]subscriber.Subscriber{
			"chat":  make(map[string]subscriber.Subscriber),
			"trade": make(map[string]subscriber.Subscriber),
			"hc":    make(map[string]subscriber.Subscriber),
		},
		Moderators: make([]string, 0),
	}
}
