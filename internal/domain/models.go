package domain

import "time"

type Feed struct {
	ID              string
	URL             string
	Name            string
	Enabled         bool
	IntervalSeconds int
	BatchEnabled    bool
	BatchWindowSecs int
	ETag            *string
	LastModified    *string
	LastFetchAt     *time.Time
	NextFetchAt     *time.Time
	LastError       *string
	ErrorCount      int
	CreatedAt       time.Time
	UpdatedAt       time.Time
	InitializedAt   *time.Time
}

type ContactType string

const (
	ContactTelegram ContactType = "telegram"
	ContactEmail    ContactType = "email"
	ContactHTTP     ContactType = "http"
)

type ContactStatus string

const (
	ContactPending  ContactStatus = "pending"
	ContactActive   ContactStatus = "active"
	ContactDisabled ContactStatus = "disabled"
)

type Contact struct {
	ID          string
	Type        ContactType
	Status      ContactStatus
	Value       string
	DisplayName *string
	VerifiedAt  *time.Time
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

type Subscription struct {
	ID        string
	FeedID    string
	ContactID string
	Enabled   bool
	CreatedAt time.Time
}

type Item struct {
	ID          string
	FeedID      string
	ExternalID  *string
	UniqKey     string
	Title       string
	Link        string
	Summary     *string
	Author      *string
	PublishedAt *time.Time
	CreatedAt   time.Time
}

type DeliveryStatus string

const (
	DeliveryPending    DeliveryStatus = "pending"
	DeliveryInProgress DeliveryStatus = "in_progress"
	DeliverySent       DeliveryStatus = "sent"
	DeliveryFailed     DeliveryStatus = "failed"
	DeliveryDead       DeliveryStatus = "dead"
)

type Delivery struct {
	ID            string
	ItemID        string
	ContactID     string
	Status        DeliveryStatus
	AttemptCount  int
	LastError     *string
	LastAttemptAt *time.Time
	SentAt        *time.Time
	NextRetryAt   *time.Time
	CreatedAt     time.Time
	UpdatedAt     time.Time
}

type DeliveryWithItem struct {
	Delivery Delivery
	Item     Item
}
