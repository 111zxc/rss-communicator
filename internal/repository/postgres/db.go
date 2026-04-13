package postgres

import (
	"context"
	"database/sql"

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
	}, nil
}

func (d *DB) Close() error                   { return d.db.Close() }
func (d *DB) Ping(ctx context.Context) error { return d.db.PingContext(ctx) }
func (d *DB) SQL() *sql.DB                   { return d.db }

func (d *DB) Contacts() *ContactsRepository { return d.contacts }
func (d *DB) Feeds() *FeedsRepository       { return d.feeds }
func (d *DB) Groups() *GroupsRepository     { return d.groups }
func (d *DB) RegistrationCodes() *RegistrationCodesRepository {
	return d.regCodes
}
func (d *DB) Items() *ItemsRepository                 { return d.items }
func (d *DB) Subscriptions() *SubscriptionsRepository { return d.subscriptions }
func (d *DB) Deliveries() *DeliveriesRepository       { return d.deliveries }
