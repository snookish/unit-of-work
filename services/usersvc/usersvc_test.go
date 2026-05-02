package usersvc_test

import (
	"context"
	"errors"
	"testing"

	"github.com/snookish/unit-of-work/models"
	"github.com/snookish/unit-of-work/repositories"
)

// -----------------------------------------------------------------------
// Mock implementations
// -----------------------------------------------------------------------

type mockUserRepo struct {
	users  map[int64]*models.User
	nextID int64
	errOn  string
}

func newMockUserRepo() *mockUserRepo {
	return &mockUserRepo{users: make(map[int64]*models.User), nextID: 1}
}

func (m *mockUserRepo) Create(_ context.Context, u *models.User) error {
	if m.errOn == "Create" {
		return errors.New("mock: create user failed")
	}
	u.ID = m.nextID
	m.nextID++
	m.users[u.ID] = u
	return nil
}

func (m *mockUserRepo) QueryByID(_ context.Context, id int64) (*models.User, error) {
	if m.errOn == "GetByID" {
		return nil, errors.New("mock: get user failed")
	}
	u, ok := m.users[id]
	if !ok {
		return nil, errors.New("mock: user not found")
	}
	return u, nil
}

func (m *mockUserRepo) Update(_ context.Context, u *models.User) error {
	if m.errOn == "Update" {
		return errors.New("mock: update user failed")
	}
	m.users[u.ID] = u
	return nil
}

type mockAuditLogRepo struct {
	logs  []*models.AuditLog
	errOn string
}

func (m *mockAuditLogRepo) Create(_ context.Context, l *models.AuditLog) error {
	if m.errOn == "Create" {
		return errors.New("mock: create audit log failed")
	}
	m.logs = append(m.logs, l)
	return nil
}

func (m *mockAuditLogRepo) ListByUserID(_ context.Context, userID int64) ([]*models.AuditLog, error) {
	var out []*models.AuditLog
	for _, l := range m.logs {
		if l.UserID == userID {
			out = append(out, l)
		}
	}
	return out, nil
}

// -----------------------------------------------------------------------
// Testable service variant
// -----------------------------------------------------------------------

type ReposFactory func() repositories.Repos

type TestableUserService struct {
	factory ReposFactory
}

func (s *TestableUserService) CreateUser(ctx context.Context, firstName, lastName, email string) (*models.User, error) {
	repos := s.factory()
	u := &models.User{Email: email, FirstName: firstName, LastName: lastName}

	if err := repos.Users.Create(ctx, u); err != nil {
		return nil, err
	}

	if err := repos.AuditLogs.Create(ctx, &models.AuditLog{UserID: u.ID, Action: "user.created"}); err != nil {
		return nil, err
	}
	return u, nil
}

// -----------------------------------------------------------------------
// Tests
// -----------------------------------------------------------------------

func TestCreateUserSuccess(t *testing.T) {
	users := newMockUserRepo()
	audit := &mockAuditLogRepo{}

	svc := &TestableUserService{
		factory: func() repositories.Repos {
			return repositories.Repos{Users: users, AuditLogs: audit}
		},
	}

	user, err := svc.CreateUser(context.Background(), "Hello", "World", "hello@world.com")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if user.ID == 0 {
		t.Fatal("expected non-zero user ID")
	}

	if len(audit.logs) != 1 {
		t.Fatalf("expected 1 audit log, got %d", len(audit.logs))
	}

	if audit.logs[0].Action != "user.created" {
		t.Errorf("unexpected audit action: %s", audit.logs[0].Action)
	}
}

func TestCreateUserAuditLogFailure(t *testing.T) {
	users := newMockUserRepo()
	audit := &mockAuditLogRepo{errOn: "Create"}

	svc := &TestableUserService{
		factory: func() repositories.Repos {
			return repositories.Repos{Users: users, AuditLogs: audit}
		},
	}

	_, err := svc.CreateUser(context.Background(), "Hello", "World", "hello@world.com")
	if err == nil {
		t.Fatal("expected error when audit log insert fails")
	}
}
