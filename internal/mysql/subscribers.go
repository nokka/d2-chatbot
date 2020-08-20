package mysql

import (
	"database/sql"
	"time"

	"github.com/nokka/d2-chatbot/internal/subscriber"
)

// SubscriberRepository ...
type SubscriberRepository struct {
	db *sql.DB
}

// FindSubscribers ...
func (r *SubscriberRepository) FindSubscribers(chatID string) ([]subscriber.Subscriber, error) {
	results, err := r.db.Query(`SELECT account, online, banned_until FROM subscribers WHERE chat = ?`, chatID)
	if err != nil {
		return nil, err
	}

	subs := make([]subscriber.Subscriber, 0)

	for results.Next() {
		var sub subscriber.Subscriber

		err = results.Scan(&sub.Account, &sub.Online, &sub.BannedUntil)
		if err != nil {
			return nil, err
		}

		subs = append(subs, sub)
	}

	return subs, nil
}

// FindEligibleSubscribers ...
func (r *SubscriberRepository) FindEligibleSubscribers(chatID string) ([]subscriber.Subscriber, error) {
	results, err := r.db.Query(`
	SELECT account, online FROM subscribers
		WHERE chat = ?
		AND online = true
		AND (banned_until IS NULL OR banned_until <= NOW())
		`, chatID)
	if err != nil {
		return nil, err
	}

	subs := make([]subscriber.Subscriber, 0)

	for results.Next() {
		var sub subscriber.Subscriber

		err = results.Scan(&sub.Account, &sub.Online, &sub.BannedUntil)
		if err != nil {
			return nil, err
		}

		subs = append(subs, sub)
	}

	return subs, nil
}

// Subscribe ...
func (r *SubscriberRepository) Subscribe(account string, chatID string) error {
	result, err := r.db.Query(`INSERT INTO subscribers (account, chat) VALUES (?,?) ON DUPLICATE KEY UPDATE account=account;`, account, chatID)
	if err != nil {
		return err
	}

	defer result.Close()

	return nil
}

// Unsubscribe ...
func (r *SubscriberRepository) Unsubscribe(account string, chatID string) error {
	result, err := r.db.Query(`DELETE FROM subscribers WHERE account = ? AND chat = ?;`, account, chatID)
	if err != nil {
		return err
	}

	defer result.Close()

	return nil
}

// UpdateOnlineStatus ...
func (r *SubscriberRepository) UpdateOnlineStatus(account string, online bool) error {
	result, err := r.db.Query(`UPDATE subscribers set online = ? WHERE account = ?;`, online, account)
	if err != nil {
		return err
	}

	defer result.Close()

	return nil
}

// UpdateBannedUntil ...
func (r *SubscriberRepository) UpdateBannedUntil(account string, chatID string, until *time.Time) error {
	result, err := r.db.Query(`UPDATE subscribers set banned_until = ? WHERE account = ? AND chat = ?;`, until, account, chatID)
	if err != nil {
		return err
	}

	defer result.Close()

	return nil
}

// FindModerators ...
func (r *SubscriberRepository) FindModerators() ([]string, error) {
	results, err := r.db.Query(`SELECT account FROM moderators`)
	if err != nil {
		return nil, err
	}

	mods := make([]string, 0)

	for results.Next() {
		var mod string

		err = results.Scan(&mod)
		if err != nil {
			return nil, err
		}

		mods = append(mods, mod)
	}

	return mods, nil
}

// NewSubscriberRepository ...
func NewSubscriberRepository(db *sql.DB) *SubscriberRepository {
	return &SubscriberRepository{
		db: db,
	}
}
