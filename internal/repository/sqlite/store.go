package sqlite

import (
	"context"
	"database/sql"

	"github.com/111zxc/rss-communicator/internal/repository"
)

type dbtx interface {
	ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error)
	QueryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error)
	QueryRowContext(ctx context.Context, query string, args ...any) *sql.Row
}

type txStore struct {
	contacts      *ContactsRepository
	feeds         *FeedsRepository
	groups        *GroupsRepository
	regCodes      *RegistrationCodesRepository
	items         *ItemsRepository
	subscriptions *SubscriptionsRepository
	deliveries    *DeliveriesRepository
	outbox        *OutboxRepository
}

func newTxStore(tx *sql.Tx) repository.Store {
	return &txStore{
		contacts:      NewContactsRepositoryTx(tx),
		feeds:         NewFeedsRepositoryTx(tx),
		groups:        NewGroupsRepositoryTx(tx),
		regCodes:      NewRegistrationCodesRepositoryTx(tx),
		items:         NewItemsRepositoryTx(tx),
		subscriptions: NewSubscriptionsRepositoryTx(tx),
		deliveries:    NewDeliveriesRepositoryTx(tx),
		outbox:        NewOutboxRepositoryTx(tx),
	}
}

func (s *txStore) Ping(context.Context) error                                { return nil }
func (s *txStore) Feeds() repository.FeedsRepository                         { return s.feeds }
func (s *txStore) Contacts() repository.ContactsRepository                   { return s.contacts }
func (s *txStore) Subscriptions() repository.SubscriptionsRepository         { return s.subscriptions }
func (s *txStore) Groups() repository.GroupsRepository                       { return s.groups }
func (s *txStore) RegistrationCodes() repository.RegistrationCodesRepository { return s.regCodes }
func (s *txStore) Items() repository.ItemsRepo                               { return s.items }
func (s *txStore) Deliveries() repository.DeliveriesRepository               { return s.deliveries }
func (s *txStore) Outbox() repository.OutboxRepository                       { return s.outbox }
