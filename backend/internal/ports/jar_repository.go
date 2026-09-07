// Package ports defines the interfaces required by the application layer.
package ports

import (
	"context"

	"github.com/google/uuid"
	"github.com/joshu-sajeev/paisa/internal/domain/jar"
)

// JarRepository defines the persistence port for jars.
type JarRepository interface {
	// Create creates a new jar.
	Create(ctx context.Context, j *jar.Jar) error

	// List gets all jars.
	List(ctx context.Context) ([]*jar.Jar, error)

	// FindByID gets a jar by its ID.
	FindByID(ctx context.Context, id uuid.UUID) (*jar.Jar, error)

	// Save persists the current state of an existing jar.
	Save(ctx context.Context, j *jar.Jar) error

	// UpdateAllocations persists the allocation values for the provided jars.
	UpdateAllocations(ctx context.Context, jars []*jar.Jar) error
}
