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
			s.logger.WarnContext(ctx, "account name already exists", slog.String("name", acc.Name))
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
	s.logger.DebugContext(ctx, "listing all active accounts")

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
		"active accounts listed",
		slog.Int("count", len(accounts)),
	)

	return accounts, nil
}

// UpdateName updates an account name
func (s *AccountService) UpdateName(ctx context.Context, id uuid.UUID, name string) error {
	s.logger.DebugContext(
		ctx,
		"updating account name",
		slog.String("id", id.String()),
		slog.String("name", name),
	)

	if err := s.repo.UpdateName(ctx, id, name); err != nil {
		s.logger.ErrorContext(
			ctx,
			"repository update name failed",
			slog.String("error", err.Error()),
			slog.String("id", id.String()),
			slog.String("name", name),
		)
		return err
	}

	s.logger.InfoContext(
		ctx,
		"account name updated successfully",
		slog.String("id", id.String()),
		slog.String("name", name),
	)

	return nil
}

// Archive archives an account.
func (s *AccountService) Archive(ctx context.Context, id uuid.UUID) error {
	s.logger.InfoContext(
		ctx,
		"archiving account",
		slog.String("id", id.String()),
	)

	if err := s.repo.SetArchived(ctx, id, true); err != nil {
		s.logger.ErrorContext(
			ctx,
			"failed to archive account",
			slog.String("error", err.Error()),
			slog.String("id", id.String()),
		)
		return err
	}

	s.logger.InfoContext(
		ctx,
		"account archived successfully",
		slog.String("id", id.String()),
	)

	return nil
}

func (s *AccountService) Unarchive(ctx context.Context, id uuid.UUID) error {
	s.logger.InfoContext(
		ctx,
		"unarchiving account",
		slog.String("id", id.String()),
	)

	if err := s.repo.SetArchived(ctx, id, false); err != nil {
		s.logger.ErrorContext(
			ctx,
			"failed to unarchive account",
			slog.String("error", err.Error()),
			slog.String("id", id.String()),
		)
		return err
	}

	s.logger.InfoContext(
		ctx,
		"account unarchived successfully",
		slog.String("id", id.String()),
	)

	return nil
}
