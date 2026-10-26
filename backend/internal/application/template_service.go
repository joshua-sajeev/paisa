package application

import (
	"context"
	"log/slog"

	"github.com/google/uuid"
	"github.com/joshu-sajeev/paisa/internal/domain/transaction"
	"github.com/joshu-sajeev/paisa/internal/ports"
)

type TemplateService struct {
	repo   ports.TemplateRepository
	logger *slog.Logger
}

func NewTemplateService(repo ports.TemplateRepository, logger *slog.Logger) *TemplateService {
	return &TemplateService{repo: repo, logger: logger}
}
func (s *TemplateService) Create(ctx context.Context, name string, typ transaction.TransactionType, category transaction.TransactionCategory, from, to, jar *uuid.UUID, amount *int64) (*transaction.Template, error) {
	t, err := transaction.NewTemplate(name, typ, uuidValue(from), uuidValue(to), uuidValue(jar), amount, category)
	if err != nil {
		return nil, err
	}
	t.FromAccountID, t.ToAccountID, t.JarID = from, to, jar
	if err := s.repo.Create(ctx, t); err != nil {
		return nil, err
	}
	return t, nil
}
func (s *TemplateService) List(ctx context.Context, opts ports.ListOptions) ([]*transaction.Template, error) {
	return s.repo.List(ctx, opts)
}
func (s *TemplateService) GetByID(ctx context.Context, id uuid.UUID) (*transaction.Template, error) {
	return s.repo.GetByID(ctx, id)
}
func (s *TemplateService) Update(ctx context.Context, t *transaction.Template) error {
	if t.Name == "" {
		return transaction.ErrTemplateInvalidName
	}
	if !t.Type.IsValid() {
		return transaction.ErrInvalidTransactionType
	}
	if !t.Category.IsValid() {
		return transaction.ErrInvalidCategory
	}
	if t.Amount != nil && *t.Amount <= 0 {
		return transaction.ErrTemplateInvalidAmount
	}
	return s.repo.Update(ctx, t)
}
func (s *TemplateService) Delete(ctx context.Context, id uuid.UUID) error {
	return s.repo.Delete(ctx, id)
}
func uuidValue(id *uuid.UUID) uuid.UUID {
	if id == nil {
		return uuid.Nil
	}
	return *id
}
