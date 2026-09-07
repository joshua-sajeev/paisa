// Package application contains application services.
package application

import (
	"context"
	"errors"
	"log/slog"

	"github.com/google/uuid"
	"github.com/joshu-sajeev/paisa/internal/domain/account"
	"github.com/joshu-sajeev/paisa/internal/ports"
)

// AccountService handles account application use cases.
type AccountService struct {
	repo   ports.AccountRepository
	logger *slog.Logger
}

// NewAccountService creates a new AccountService.
func NewAccountService(repo ports.AccountRepository, logger *slog.Logger) *AccountService {
	return &AccountService{
		repo:   repo,
		logger: logger,
	}
}

// Create creates a new account.
func (s *AccountService) Create(ctx context.Context, name string) (*account.Account, error) {
	s.logger.DebugContext(
		ctx,
		"creating new account",
		slog.String("name", name),
	)

	acc, err := account.NewAccount(name)
	if err != nil {
		s.logger.WarnContext(
			ctx,
			"invalid account",
			slog.String("name", name),
			slog.String("error", err.Error()),
		)
		return nil, err
	}

	if err := s.repo.Create(ctx, acc); err != nil {
		if errors.Is(err, account.ErrAccountNameExists) {
			s.logger.WarnContext(
				ctx,
				"account name already exists",
				slog.String("name", acc.Name),
			)
			return nil, err
		}

		s.logger.ErrorContext(
			ctx,
			"failed to create account in repository",
			slog.String("error", err.Error()),
			slog.String("name", acc.Name),
		)

		return nil, err
	}

	s.logger.InfoContext(
		ctx,
		"account created successfully",
		slog.String("name", acc.Name),
		slog.String("id", acc.ID.String()),
	)

	return acc, nil
}

// List gets all accounts.
func (s *AccountService) List(ctx context.Context) ([]*account.Account, error) {
	s.logger.DebugContext(ctx, "listing all accounts")

	accounts, err := s.repo.List(ctx)
	if err != nil {
		s.logger.ErrorContext(
			ctx,
			"repository list failed",
			slog.String("error", err.Error()),
		)
		return nil, err
	}

	s.logger.InfoContext(
		ctx,
		"accounts listed",
		slog.Int("count", len(accounts)),
	)

	return accounts, nil
}

// Update updates the provided account fields.
func (s *AccountService) Update(
	ctx context.Context,
	id uuid.UUID,
	name *string,
	isArchived *bool,
) error {
	s.logger.DebugContext(
		ctx,
		"updating account",
		slog.String("id", id.String()),
	)

	acc, err := s.repo.FindByID(ctx, id)
	if err != nil {
		s.logger.ErrorContext(
			ctx,
			"failed to find account",
			slog.String("error", err.Error()),
			slog.String("id", id.String()),
		)
		return err
	}

	changed := false

	if name != nil {
		if err := acc.Rename(*name); err != nil {
			s.logger.WarnContext(
				ctx,
				"invalid account name",
				slog.String("id", id.String()),
				slog.String("error", err.Error()),
			)
			return err
		}

		changed = true
	}

	if isArchived != nil {
		if *isArchived {
			err := acc.Archive()
			if err != nil {
				if !errors.Is(err, account.ErrAccountAlreadyArchived) {
					return err
				}

				s.logger.DebugContext(
					ctx,
					"account already archived",
					slog.String("id", id.String()),
				)
			} else {
				changed = true
			}
		} else {
			err := acc.Unarchive()
			if err != nil {
				if !errors.Is(err, account.ErrAccountNotArchived) {
					return err
				}

				s.logger.DebugContext(
					ctx,
					"account already active",
					slog.String("id", id.String()),
				)
			} else {
				changed = true
			}
		}
	}

	if !changed {
		return nil
	}

	if err := s.repo.Save(ctx, acc); err != nil {
		s.logger.ErrorContext(
			ctx,
			"failed to save account",
			slog.String("error", err.Error()),
			slog.String("id", id.String()),
		)
		return err
	}

	s.logger.InfoContext(
		ctx,
		"account updated successfully",
		slog.String("id", id.String()),
	)

	return nil
}
