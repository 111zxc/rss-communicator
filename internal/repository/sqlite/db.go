package sqlite

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	"github.com/111zxc/rss-communicator/internal/repository"
	_ "modernc.org/sqlite"
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
	if strings.TrimSpace(dsn) == "" {
		dsn = "file:rss.db"
	}

	db, err := sql.Open("sqlite", dsn)
	if err != nil {
		return nil, err
	}

	db.SetMaxOpenConns(1)
	db.SetMaxIdleConns(1)

	ctx := context.Background()
	pragmas := []string{
		"PRAGMA foreign_keys = ON",
		"PRAGMA journal_mode = WAL",
		"PRAGMA busy_timeout = 5000",
	}
	for _, pragma := range pragmas {
		if _, err := db.ExecContext(ctx, pragma); err != nil {
			_ = db.Close()
			return nil, fmt.Errorf("sqlite init pragma failed: %w", err)
		}
	}

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
