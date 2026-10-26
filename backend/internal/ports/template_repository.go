package ports

import (
	"context"
	"github.com/google/uuid"
	"github.com/joshu-sajeev/paisa/internal/domain/transaction"
)

// TemplateRepository defines persistence operations for Template entities.
type TemplateRepository interface {
	Create(ctx context.Context, tmpl *transaction.Template) error
	GetByID(ctx context.Context, id uuid.UUID) (*transaction.Template, error)
	Update(ctx context.Context, tmpl *transaction.Template) error
	Delete(ctx context.Context, id uuid.UUID) error
	List(ctx context.Context, opts ListOptions) ([]*transaction.Template, error)
}

// ListOptions can be used for pagination or filtering when listing templates.
// For now it's a placeholder; you can extend it as needed.
type ListOptions struct {
	Limit      int
	Offset     int
	IsArchived bool
}
