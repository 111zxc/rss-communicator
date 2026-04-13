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
}

func NewGroupService(groups repository.GroupsRepository, feeds repository.FeedsRepository, contacts repository.ContactsRepository, subscriptions repository.SubscriptionsRepository) *GroupService {
	return &GroupService{groups: groups, feeds: feeds, contacts: contacts, subscriptions: subscriptions}
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
	if err := s.groups.AddContact(ctx, groupID, contactID); err != nil {
		return err
	}
	feedIDs, err := s.groups.ListFeedIDs(ctx, groupID)
	if err != nil {
		return err
	}
	for _, feedID := range feedIDs {
		if err := s.subscriptions.AddGroup(ctx, feedID, contactID, groupID); err != nil {
			return err
		}
	}
	return nil
}

func (s *GroupService) RemoveContact(ctx context.Context, groupID, contactID string) error {
	if strings.TrimSpace(groupID) == "" || strings.TrimSpace(contactID) == "" {
		return ErrBadRequest
	}
	if err := s.groups.RemoveContact(ctx, groupID, contactID); err != nil {
		return err
	}
	return s.subscriptions.RemoveGroupByContact(ctx, groupID, contactID)
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
	if err := s.groups.AddFeed(ctx, groupID, feedID); err != nil {
		return err
	}
	contactIDs, err := s.groups.ListContactIDs(ctx, groupID)
	if err != nil {
		return err
	}
	for _, contactID := range contactIDs {
		if err := s.subscriptions.AddGroup(ctx, feedID, contactID, groupID); err != nil {
			return err
		}
	}
	return nil
}

func (s *GroupService) RemoveFeed(ctx context.Context, groupID, feedID string) error {
	if strings.TrimSpace(groupID) == "" || strings.TrimSpace(feedID) == "" {
		return ErrBadRequest
	}
	if err := s.groups.RemoveFeed(ctx, groupID, feedID); err != nil {
		return err
	}
	return s.subscriptions.RemoveGroupByFeed(ctx, groupID, feedID)
}
