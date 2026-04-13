package service

import (
	"context"
	"testing"
	"time"

	"github.com/111zxc/rss-communicator/internal/domain"
)

func TestRegistrationCodeServiceCreateNormalizesCode(t *testing.T) {
	repo := &registrationCodesRepoStub{}
	svc := NewRegistrationCodeService(repo, &groupRepoStub{})

	_, err := svc.Create(context.Background(), RegistrationCodeInput{
		Code:    " abc123 ",
		Name:    "Promo",
		Enabled: true,
	})
	if err != nil {
		t.Fatalf("Create returned error: %v", err)
	}
	if repo.created.Code != "ABC123" {
		t.Fatalf("expected normalized code, got %q", repo.created.Code)
	}
}

type registrationCodesRepoStub struct {
	created domain.RegistrationCode
}

func (s *registrationCodesRepoStub) Create(_ context.Context, code domain.RegistrationCode) (domain.RegistrationCode, error) {
	s.created = code
	return code, nil
}
func (s *registrationCodesRepoStub) Update(context.Context, string, string, string, *string, bool, *int, *time.Time) (domain.RegistrationCode, error) {
	return domain.RegistrationCode{}, nil
}
func (s *registrationCodesRepoStub) GetByID(context.Context, string) (domain.RegistrationCode, error) {
	return domain.RegistrationCode{ID: "code-1"}, nil
}
func (s *registrationCodesRepoStub) GetByCode(context.Context, string) (domain.RegistrationCode, error) {
	return domain.RegistrationCode{}, nil
}
func (s *registrationCodesRepoStub) List(context.Context, int, int) ([]domain.RegistrationCode, int, error) {
	return nil, 0, nil
}
func (s *registrationCodesRepoStub) Delete(context.Context, string) error { return nil }
func (s *registrationCodesRepoStub) ListGroups(context.Context, string) ([]domain.Group, error) {
	return nil, nil
}
func (s *registrationCodesRepoStub) AddGroup(context.Context, string, string) error { return nil }
func (s *registrationCodesRepoStub) RemoveGroup(context.Context, string, string) error {
	return nil
}
func (s *registrationCodesRepoStub) IncrementUse(context.Context, string) error { return nil }
