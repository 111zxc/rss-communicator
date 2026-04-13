package repository

import (
	"context"
	"time"

	"github.com/111zxc/rss-communicator/internal/domain"
)

type Tx interface {
	Commit() error
	Rollback() error
}

type Store interface {
	Ping(ctx context.Context) error
	BeginTx(ctx context.Context) (Tx, error)

	Feeds() FeedsRepository
	Contacts() ContactsRepository
	Subscriptions() SubscriptionsRepository
	Items() ItemsRepo
	Deliveries() DeliveriesRepository
	Outbox() OutboxRepository
}

type FeedsRepository interface {
	Create(ctx context.Context, f domain.Feed) (domain.Feed, error)
	ListDue(ctx context.Context, now time.Time, limit int) ([]domain.Feed, error)
	MarkFetched(ctx context.Context, feedID string, fetchedAt time.Time, nextAt time.Time, etag, lastModified *string) error
	MarkFetchError(ctx context.Context, feedID string, errMsg string) error
	GetByID(ctx context.Context, feedID string) (domain.Feed, error)
	UpdateBatching(ctx context.Context, feedID string, batchEnabled bool, batchWindowSecs int) (domain.Feed, error)
	List(ctx context.Context, limit, offset int) ([]domain.Feed, int, error)
	Delete(ctx context.Context, feedID string) error
}

type ContactsRepository interface {
	UpsertTelegramActive(ctx context.Context, chatID string, username *string, displayName *string, verifiedAt time.Time) (domain.Contact, error)
	CreateTelegram(ctx context.Context, chatID string, username *string, displayName *string, status domain.ContactStatus, verifiedAt *time.Time) (domain.Contact, error)
	UpdateTelegram(ctx context.Context, contactID string, chatID string, username *string, displayName *string, status domain.ContactStatus, verifiedAt *time.Time) (domain.Contact, error)
	CreateHTTP(ctx context.Context, value string, displayName *string, status domain.ContactStatus, cfg domain.HTTPContactConfig, verifiedAt *time.Time) (domain.Contact, error)
	UpdateHTTP(ctx context.Context, contactID string, value string, displayName *string, status domain.ContactStatus, cfg domain.HTTPContactConfig, verifiedAt *time.Time) (domain.Contact, error)
	GetHTTPConfig(ctx context.Context, contactID string) (domain.HTTPContactConfig, error)
	GetByTypeValue(ctx context.Context, typ domain.ContactType, value string) (domain.Contact, error)
	GetByID(ctx context.Context, id string) (domain.Contact, error)
	List(ctx context.Context, limit, offset int) ([]domain.Contact, int, error)
	Delete(ctx context.Context, id string) error
}

type SubscriptionsRepository interface {
	ListByFeed(ctx context.Context, feedID string) ([]domain.Subscription, error)
	ListByContact(ctx context.Context, contactID string) ([]domain.Subscription, error)
	Add(ctx context.Context, feedID, contactID string) error
	Remove(ctx context.Context, feedID, contactID string) error
}

type ItemsRepo interface {
	InsertMany(ctx context.Context, items []domain.Item) (inserted []domain.Item, err error)
}

type DeliveriesRepository interface {
	CreatePendingIfNotExists(ctx context.Context, contactID, itemID string, availableAt time.Time) (created bool, deliveryID string, err error)
	MarkSent(ctx context.Context, deliveryID string, sentAt time.Time) error
	MarkFailed(ctx context.Context, deliveryID string, errMsg string, nextRetryAt *time.Time) error
}

type OutboxRepository interface {
	Enqueue(ctx context.Context, topic string, payload any, availableAt time.Time) error
}

type Clock interface {
	NowUTC() time.Time
}
