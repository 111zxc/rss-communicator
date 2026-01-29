package service

import (
	"context"

	"github.com/111zxc/rss-communicator/internal/domain"
	"github.com/111zxc/rss-communicator/internal/repository"
)

type ContactService struct {
	contacts repository.ContactsRepository
}

func NewContactService(contacts repository.ContactsRepository) *ContactService {
	return &ContactService{contacts: contacts}
}

func (s *ContactService) List(ctx context.Context, limit, offset int) ([]domain.Contact, int, error) {
	return s.contacts.List(ctx, limit, offset)
}
