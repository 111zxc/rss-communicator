package service

import (
	"context"
	"database/sql"
	"errors"
	"strings"

	"github.com/111zxc/rss-communicator/internal/domain"
	"github.com/111zxc/rss-communicator/internal/repository"
)

type GroupService struct {
	groups        repository.GroupsRepository
	feeds         repository.FeedsRepository
	contacts      repository.ContactsRepository
	subscriptions repository.SubscriptionsRepository
	txRunner      repository.Database
}

func NewGroupService(groups repository.GroupsRepository, feeds repository.FeedsRepository, contacts repository.ContactsRepository, subscriptions repository.SubscriptionsRepository, txRunner ...repository.Database) *GroupService {
	var db repository.Database
	if len(txRunner) > 0 {
		db = txRunner[0]
	}
	return &GroupService{groups: groups, feeds: feeds, contacts: contacts, subscriptions: subscriptions, txRunner: db}
}

type GroupInput struct {
	Name        string
	Description *string
}

func (s *GroupService) List(ctx context.Context, limit, offset int) ([]domain.Group, int, error) {
	return s.groups.List(ctx, limit, offset)
}

func (s *GroupService) GetByID(ctx context.Context, groupID string) (domain.Group, error) {
	if strings.TrimSpace(groupID) == "" {
		return domain.Group{}, ErrBadRequest
	}
	g, err := s.groups.GetByID(ctx, groupID)
	if errors.Is(err, sql.ErrNoRows) {
		return domain.Group{}, ErrNotFound
	}
	return g, err
}

func (s *GroupService) Create(ctx context.Context, in GroupInput) (domain.Group, error) {
	name := strings.TrimSpace(in.Name)
	if name == "" {
		return domain.Group{}, ErrBadRequest
	}
	desc := normalizeOptional(in.Description)
	return s.groups.Create(ctx, domain.Group{Name: name, Description: desc})
}

func (s *GroupService) Update(ctx context.Context, groupID string, in GroupInput) (domain.Group, error) {
	if strings.TrimSpace(groupID) == "" {
		return domain.Group{}, ErrBadRequest
	}
	name := strings.TrimSpace(in.Name)
	if name == "" {
		return domain.Group{}, ErrBadRequest
	}
	g, err := s.groups.Update(ctx, groupID, name, normalizeOptional(in.Description))
	if errors.Is(err, sql.ErrNoRows) {
		return domain.Group{}, ErrNotFound
	}
	return g, err
}

func (s *GroupService) Delete(ctx context.Context, groupID string) error {
	if strings.TrimSpace(groupID) == "" {
		return ErrBadRequest
	}
	err := s.groups.Delete(ctx, groupID)
	if errors.Is(err, sql.ErrNoRows) {
		return ErrNotFound
	}
	return err
}

func (s *GroupService) ListContacts(ctx context.Context, groupID string) ([]domain.Contact, error) {
	if strings.TrimSpace(groupID) == "" {
		return nil, ErrBadRequest
	}
	if _, err := s.groups.GetByID(ctx, groupID); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return s.groups.ListContacts(ctx, groupID)
}

func (s *GroupService) AddContact(ctx context.Context, groupID, contactID string) error {
	if strings.TrimSpace(groupID) == "" || strings.TrimSpace(contactID) == "" {
		return ErrBadRequest
	}
	if _, err := s.groups.GetByID(ctx, groupID); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return ErrNotFound
		}
		return err
	}
	if _, err := s.contacts.GetByID(ctx, contactID); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return ErrNotFound
		}
		return err
	}
	return s.withRepos(ctx, func(groups repository.GroupsRepository, feeds repository.FeedsRepository, contacts repository.ContactsRepository, subscriptions repository.SubscriptionsRepository) error {
		if err := groups.AddContact(ctx, groupID, contactID); err != nil {
			return err
		}
		feedIDs, err := groups.ListFeedIDs(ctx, groupID)
		if err != nil {
			return err
		}
		for _, feedID := range feedIDs {
			if err := subscriptions.AddGroup(ctx, feedID, contactID, groupID); err != nil {
				return err
			}
		}
		return nil
	})
}

func (s *GroupService) RemoveContact(ctx context.Context, groupID, contactID string) error {
	if strings.TrimSpace(groupID) == "" || strings.TrimSpace(contactID) == "" {
		return ErrBadRequest
	}
	return s.withRepos(ctx, func(groups repository.GroupsRepository, feeds repository.FeedsRepository, contacts repository.ContactsRepository, subscriptions repository.SubscriptionsRepository) error {
		if err := groups.RemoveContact(ctx, groupID, contactID); err != nil {
			return err
		}
		return subscriptions.RemoveGroupByContact(ctx, groupID, contactID)
	})
}

func (s *GroupService) ListFeeds(ctx context.Context, groupID string) ([]domain.Feed, error) {
	if strings.TrimSpace(groupID) == "" {
		return nil, ErrBadRequest
	}
	if _, err := s.groups.GetByID(ctx, groupID); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return s.groups.ListFeeds(ctx, groupID)
}

func (s *GroupService) AddFeed(ctx context.Context, groupID, feedID string) error {
	if strings.TrimSpace(groupID) == "" || strings.TrimSpace(feedID) == "" {
		return ErrBadRequest
	}
	if _, err := s.groups.GetByID(ctx, groupID); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return ErrNotFound
		}
		return err
	}
	if _, err := s.feeds.GetByID(ctx, feedID); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return ErrNotFound
		}
		return err
	}
	return s.withRepos(ctx, func(groups repository.GroupsRepository, feeds repository.FeedsRepository, contacts repository.ContactsRepository, subscriptions repository.SubscriptionsRepository) error {
		if err := groups.AddFeed(ctx, groupID, feedID); err != nil {
			return err
		}
		contactIDs, err := groups.ListContactIDs(ctx, groupID)
		if err != nil {
			return err
		}
		for _, contactID := range contactIDs {
			if err := subscriptions.AddGroup(ctx, feedID, contactID, groupID); err != nil {
				return err
			}
		}
		return nil
	})
}

func (s *GroupService) RemoveFeed(ctx context.Context, groupID, feedID string) error {
	if strings.TrimSpace(groupID) == "" || strings.TrimSpace(feedID) == "" {
		return ErrBadRequest
	}
	return s.withRepos(ctx, func(groups repository.GroupsRepository, feeds repository.FeedsRepository, contacts repository.ContactsRepository, subscriptions repository.SubscriptionsRepository) error {
		if err := groups.RemoveFeed(ctx, groupID, feedID); err != nil {
			return err
		}
		return subscriptions.RemoveGroupByFeed(ctx, groupID, feedID)
	})
}

func (s *GroupService) withRepos(ctx context.Context, fn func(repository.GroupsRepository, repository.FeedsRepository, repository.ContactsRepository, repository.SubscriptionsRepository) error) error {
	if s.txRunner == nil {
		return fn(s.groups, s.feeds, s.contacts, s.subscriptions)
	}
	return s.txRunner.WithinTx(ctx, func(store repository.Store) error {
		return fn(store.Groups(), store.Feeds(), store.Contacts(), store.Subscriptions())
	})
}
