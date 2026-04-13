package service

import (
	"context"
	"database/sql"
	"errors"
	"strings"
	"time"

	"github.com/111zxc/rss-communicator/internal/domain"
	"github.com/111zxc/rss-communicator/internal/repository"
)

type RegistrationCodeService struct {
	codes  repository.RegistrationCodesRepository
	groups repository.GroupsRepository
}

func NewRegistrationCodeService(codes repository.RegistrationCodesRepository, groups repository.GroupsRepository) *RegistrationCodeService {
	return &RegistrationCodeService{codes: codes, groups: groups}
}

type RegistrationCodeInput struct {
	Code        string
	Name        string
	Description *string
	Enabled     bool
	MaxUses     *int
	ExpiresAt   *time.Time
}

func (s *RegistrationCodeService) List(ctx context.Context, limit, offset int) ([]domain.RegistrationCode, int, error) {
	return s.codes.List(ctx, limit, offset)
}

func (s *RegistrationCodeService) GetByID(ctx context.Context, codeID string) (domain.RegistrationCode, error) {
	if strings.TrimSpace(codeID) == "" {
		return domain.RegistrationCode{}, ErrBadRequest
	}
	rc, err := s.codes.GetByID(ctx, codeID)
	if errors.Is(err, sql.ErrNoRows) {
		return domain.RegistrationCode{}, ErrNotFound
	}
	return rc, err
}

func (s *RegistrationCodeService) Create(ctx context.Context, in RegistrationCodeInput) (domain.RegistrationCode, error) {
	code, name, desc, maxUses, expiresAt, err := normalizeRegistrationCodeInput(in)
	if err != nil {
		return domain.RegistrationCode{}, err
	}
	return s.codes.Create(ctx, domain.RegistrationCode{
		Code:        code,
		Name:        name,
		Description: desc,
		Enabled:     in.Enabled,
		MaxUses:     maxUses,
		ExpiresAt:   expiresAt,
	})
}

func (s *RegistrationCodeService) Update(ctx context.Context, codeID string, in RegistrationCodeInput) (domain.RegistrationCode, error) {
	if strings.TrimSpace(codeID) == "" {
		return domain.RegistrationCode{}, ErrBadRequest
	}
	code, name, desc, maxUses, expiresAt, err := normalizeRegistrationCodeInput(in)
	if err != nil {
		return domain.RegistrationCode{}, err
	}
	rc, err := s.codes.Update(ctx, codeID, code, name, desc, in.Enabled, maxUses, expiresAt)
	if errors.Is(err, sql.ErrNoRows) {
		return domain.RegistrationCode{}, ErrNotFound
	}
	return rc, err
}

func (s *RegistrationCodeService) Delete(ctx context.Context, codeID string) error {
	if strings.TrimSpace(codeID) == "" {
		return ErrBadRequest
	}
	err := s.codes.Delete(ctx, codeID)
	if errors.Is(err, sql.ErrNoRows) {
		return ErrNotFound
	}
	return err
}

func (s *RegistrationCodeService) ListGroups(ctx context.Context, codeID string) ([]domain.Group, error) {
	if strings.TrimSpace(codeID) == "" {
		return nil, ErrBadRequest
	}
	if _, err := s.codes.GetByID(ctx, codeID); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return s.codes.ListGroups(ctx, codeID)
}

func (s *RegistrationCodeService) AddGroup(ctx context.Context, codeID, groupID string) error {
	if strings.TrimSpace(codeID) == "" || strings.TrimSpace(groupID) == "" {
		return ErrBadRequest
	}
	if _, err := s.codes.GetByID(ctx, codeID); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return ErrNotFound
		}
		return err
	}
	if _, err := s.groups.GetByID(ctx, groupID); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return ErrNotFound
		}
		return err
	}
	return s.codes.AddGroup(ctx, codeID, groupID)
}

func (s *RegistrationCodeService) RemoveGroup(ctx context.Context, codeID, groupID string) error {
	if strings.TrimSpace(codeID) == "" || strings.TrimSpace(groupID) == "" {
		return ErrBadRequest
	}
	return s.codes.RemoveGroup(ctx, codeID, groupID)
}

func normalizeRegistrationCodeInput(in RegistrationCodeInput) (string, string, *string, *int, *time.Time, error) {
	code := strings.ToUpper(strings.TrimSpace(in.Code))
	name := strings.TrimSpace(in.Name)
	if code == "" || name == "" {
		return "", "", nil, nil, nil, ErrBadRequest
	}
	desc := normalizeOptional(in.Description)
	var maxUses *int
	if in.MaxUses != nil {
		if *in.MaxUses <= 0 {
			return "", "", nil, nil, nil, ErrBadRequest
		}
		v := *in.MaxUses
		maxUses = &v
	}
	var expiresAt *time.Time
	if in.ExpiresAt != nil {
		t := in.ExpiresAt.UTC()
		expiresAt = &t
	}
	return code, name, desc, maxUses, expiresAt, nil
}
