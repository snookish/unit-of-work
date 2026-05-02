package usersvc

import (
	"context"
	"fmt"

	"github.com/snookish/unit-of-work/models"
	"github.com/snookish/unit-of-work/repositories"
	"github.com/snookish/unit-of-work/uow"
)

type UserService struct {
	uow *uow.UnitOfWork
}

func NewUserService(u *uow.UnitOfWork) *UserService {
	return &UserService{uow: u}
}

func (s *UserService) CreateUser(ctx context.Context, firstName, lastName, email string) (*models.User, error) {
	u := &models.User{Email: email, FirstName: firstName, LastName: lastName}

	if err := s.uow.WithTx(ctx, nil, func(ctx context.Context, unit *uow.UnitOfWork) error {
		tx, err := unit.Tx()
		if err != nil {
			return err
		}

		repos := repositories.NewRepos(tx)

		if err := repos.Users.Create(ctx, u); err != nil {
			return fmt.Errorf("create user: %w", err)
		}

		if err := repos.AuditLogs.Create(ctx, &models.AuditLog{UserID: u.ID, Action: "user.created"}); err != nil {
			return fmt.Errorf("create audit log: %w", err)
		}

		return nil
	}); err != nil {
		return nil, err
	}

	return u, nil
}

func (s *UserService) UpdateUserEmail(ctx context.Context, userID int64, newEmail string) error {
	return s.uow.WithTx(ctx, nil, func(ctx context.Context, unit *uow.UnitOfWork) error {
		tx, err := unit.Tx()
		if err != nil {
			return err
		}

		repos := repositories.NewRepos(tx)

		user, err := repos.Users.QueryByID(ctx, userID)
		if err != nil {
			return fmt.Errorf("fetch user: %w", err)
		}

		user.Email = newEmail
		if err := repos.Users.Update(ctx, user); err != nil {
			return fmt.Errorf("update user: %w", err)
		}

		log := &models.AuditLog{
			UserID: user.ID,
			Action: fmt.Sprintf("user.email_changed to %s", newEmail),
		}
		if err := repos.AuditLogs.Create(ctx, log); err != nil {
			return fmt.Errorf("create audit log: %w", err)
		}

		return nil
	})
}
