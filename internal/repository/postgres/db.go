package postgres

import (
	"context"
	"database/sql"

	"github.com/111zxc/rss-communicator/internal/repository"
	_ "github.com/jackc/pgx/v5/stdlib"
)

type DB struct {
	db *sql.DB

	contacts      *ContactsRepository
	feeds         *FeedsRepository
	groups        *GroupsRepository
	regCodes      *RegistrationCodesRepository
	items         *ItemsRepository
	subscriptions *SubscriptionsRepository
	deliveries    *DeliveriesRepository
	outbox        *OutboxRepository
}

func New(dsn string) (*DB, error) {
	db, err := sql.Open("pgx", dsn)
	if err != nil {
		return nil, err
	}
	db.SetMaxOpenConns(10)
	db.SetMaxIdleConns(10)

	return &DB{
		db:            db,
		contacts:      NewContactsRepository(db),
		feeds:         NewFeedsRepository(db),
		groups:        NewGroupsRepository(db),
		regCodes:      NewRegistrationCodesRepository(db),
		items:         NewItemsRepository(db),
		subscriptions: NewSubscriptionsRepository(db),
		deliveries:    NewDeliveriesRepository(db),
		outbox:        NewOutboxRepository(db),
	}, nil
}

func (d *DB) Close() error                   { return d.db.Close() }
func (d *DB) Ping(ctx context.Context) error { return d.db.PingContext(ctx) }
func (d *DB) SQL() *sql.DB                   { return d.db }
func (d *DB) WithinTx(ctx context.Context, fn func(repository.Store) error) error {
	tx, err := d.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()
	if err := fn(newTxStore(tx)); err != nil {
		return err
	}
	return tx.Commit()
}

func (d *DB) Contacts() repository.ContactsRepository { return d.contacts }
func (d *DB) Feeds() repository.FeedsRepository       { return d.feeds }
func (d *DB) Groups() repository.GroupsRepository     { return d.groups }
func (d *DB) RegistrationCodes() repository.RegistrationCodesRepository {
	return d.regCodes
}
func (d *DB) Items() repository.ItemsRepo                       { return d.items }
func (d *DB) Subscriptions() repository.SubscriptionsRepository { return d.subscriptions }
func (d *DB) Deliveries() repository.DeliveriesRepository       { return d.deliveries }
func (d *DB) Outbox() repository.OutboxRepository               { return d.outbox }
