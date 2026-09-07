package application

import (
	"context"
	"errors"
	"log/slog"

	"github.com/google/uuid"
	"github.com/joshu-sajeev/paisa/internal/domain/jar"
	"github.com/joshu-sajeev/paisa/internal/ports"
)

// JarService handles jar application use cases.
type JarService struct {
	repo   ports.JarRepository
	logger *slog.Logger
}

// NewJarService creates a new JarService.
func NewJarService(repo ports.JarRepository, logger *slog.Logger) *JarService {
	return &JarService{
		repo:   repo,
		logger: logger,
	}
}

// Create creates a new jar, validating that the resulting allocation configuration remains valid.
func (s *JarService) Create(ctx context.Context, name string, allocationType jar.AllocationType, allocationValue int64) (*jar.Jar, error) {
	s.logger.DebugContext(
		ctx,
		"creating new jar",
		slog.String("name", name),
		slog.String("allocation_type", string(allocationType)),
		slog.Int64("allocation_value", allocationValue),
	)

	newJar, err := jar.NewJar(name, allocationType, allocationValue)
	if err != nil {
		s.logger.WarnContext(
			ctx,
			"invalid jar",
			slog.String("name", name),
			slog.String("error", err.Error()),
		)
		return nil, err
	}

	existingJars, err := s.repo.List(ctx)
	if err != nil {
		s.logger.ErrorContext(
			ctx,
			"failed to fetch existing jars for validation",
			slog.String("error", err.Error()),
		)
		return nil, err
	}

	activeJars := make([]*jar.Jar, 0, len(existingJars))
	for _, j := range existingJars {
		if !j.IsArchived {
			activeJars = append(activeJars, j)
		}
	}

	proposedJars := append(activeJars, newJar)

	if err := jar.ValidateAllocationConfiguration(proposedJars); err != nil {
		s.logger.WarnContext(
			ctx,
			"proposed jar would create invalid allocation configuration",
			slog.String("name", name),
			slog.String("error", err.Error()),
		)
		return nil, err
	}

	if err := s.repo.Create(ctx, newJar); err != nil {
		if errors.Is(err, jar.ErrJarNameExists) {
			s.logger.WarnContext(
				ctx,
				"jar name already exists",
				slog.String("name", newJar.Name),
			)
			return nil, err
		}

		s.logger.ErrorContext(
			ctx,
			"failed to create jar in repository",
			slog.String("error", err.Error()),
			slog.String("name", newJar.Name),
		)
		return nil, err
	}

	s.logger.InfoContext(
		ctx,
		"jar created successfully",
		slog.String("name", newJar.Name),
		slog.String("id", newJar.ID.String()),
	)

	return newJar, nil
}

// List gets all jars.
func (s *JarService) List(ctx context.Context) ([]*jar.Jar, error) {
	s.logger.DebugContext(ctx, "listing all active jars")

	jars, err := s.repo.List(ctx)
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
		"active jars listed",
		slog.Int("count", len(jars)),
	)

	return jars, nil
}

// Update updates the name and archive state of a jar.
func (s *JarService) Update(
	ctx context.Context,
	id uuid.UUID,
	name *string,
	isArchived *bool,
) error {
	s.logger.DebugContext(
		ctx,
		"updating jar",
		slog.String("id", id.String()),
	)

	j, err := s.repo.FindByID(ctx, id)
	if err != nil {
		s.logger.ErrorContext(
			ctx,
			"failed to find jar for update",
			slog.String("error", err.Error()),
			slog.String("id", id.String()),
		)
		return err
	}

	changed := false

	if name != nil && *name != j.Name {
		if err := j.Rename(*name); err != nil {
			s.logger.WarnContext(
				ctx,
				"invalid jar name",
				slog.String("id", id.String()),
				slog.String("error", err.Error()),
			)
			return err
		}

		changed = true
	}

	if isArchived != nil && *isArchived != j.IsArchived {
		if *isArchived {
			err = j.Archive()
		} else {
			err = j.Unarchive()
		}

		if err != nil {
			s.logger.WarnContext(
				ctx,
				"invalid jar archive state transition",
				slog.String("id", id.String()),
				slog.String("error", err.Error()),
			)
			return err
		}

		changed = true
	}

	if !changed {
		return nil
	}

	// Archiving or unarchiving changes the active allocation
	// configuration, so validate the resulting configuration.
	if isArchived != nil {
		existingJars, err := s.repo.List(ctx)
		if err != nil {
			s.logger.ErrorContext(
				ctx,
				"failed to fetch jars for allocation validation",
				slog.String("error", err.Error()),
			)
			return err
		}

		proposedJars := make([]*jar.Jar, 0, len(existingJars)+1)

		found := false

		for _, existing := range existingJars {
			if existing.ID == j.ID {
				proposedJars = append(proposedJars, j)
				found = true
				continue
			}

			proposedJars = append(proposedJars, existing)
		}

		// List returns only active jars. Therefore an archived jar
		// being unarchived will not be present in the result.
		if !found && !j.IsArchived {
			proposedJars = append(proposedJars, j)
		}

		activeJars := make([]*jar.Jar, 0, len(proposedJars))

		for _, proposed := range proposedJars {
			if !proposed.IsArchived {
				activeJars = append(activeJars, proposed)
			}
		}

		if err := jar.ValidateAllocationConfiguration(activeJars); err != nil {
			s.logger.WarnContext(
				ctx,
				"jar update would create invalid allocation configuration",
				slog.String("id", id.String()),
				slog.String("error", err.Error()),
			)
			return err
		}
	}

	if err := s.repo.Save(ctx, j); err != nil {
		s.logger.ErrorContext(
			ctx,
			"repository jar save failed",
			slog.String("error", err.Error()),
			slog.String("id", id.String()),
		)
		return err
	}

	s.logger.InfoContext(
		ctx,
		"jar updated successfully",
		slog.String("id", id.String()),
	)

	return nil
}

// UpdateAllocations updates the allocations of multiple jars.
// Validates that the resulting allocation configuration remains valid.
func (s *JarService) UpdateAllocations(
	ctx context.Context,
	updates []jar.JarAllocationUpdate,
) error {
	s.logger.DebugContext(
		ctx,
		"updating jar allocations",
		slog.Int("count", len(updates)),
	)

	existingJars, err := s.repo.List(ctx)
	if err != nil {
		s.logger.ErrorContext(
			ctx,
			"failed to fetch existing jars for validation",
			slog.String("error", err.Error()),
		)
		return err
	}

	updatedMap := make(map[uuid.UUID]jar.JarAllocationUpdate, len(updates))

	for _, update := range updates {
		updatedMap[update.ID] = update
	}

	proposedJars := make([]*jar.Jar, 0, len(existingJars))
	updatedJars := make([]*jar.Jar, 0, len(updates))

	for _, existing := range existingJars {
		if existing.IsArchived {
			continue
		}

		update, ok := updatedMap[existing.ID]
		if !ok {
			proposedJars = append(proposedJars, existing)
			continue
		}

		proposed := *existing

		if err := proposed.UpdateAllocation(
			update.AllocationType,
			update.AllocationValue,
		); err != nil {
			s.logger.WarnContext(
				ctx,
				"invalid jar allocation",
				slog.String("id", update.ID.String()),
				slog.String("error", err.Error()),
			)
			return err
		}

		proposedJars = append(proposedJars, &proposed)
		updatedJars = append(updatedJars, &proposed)
	}

	if err := jar.ValidateAllocationConfiguration(proposedJars); err != nil {
		s.logger.WarnContext(
			ctx,
			"allocation update would create invalid configuration",
			slog.String("error", err.Error()),
		)
		return err
	}

	if err := s.repo.UpdateAllocations(ctx, updatedJars); err != nil {
		s.logger.ErrorContext(
			ctx,
			"repository jar allocation update failed",
			slog.String("error", err.Error()),
		)
		return err
	}

	s.logger.InfoContext(
		ctx,
		"jar allocations updated successfully",
		slog.Int("count", len(updatedJars)),
	)

	return nil
}

// AllocateIncome distributes income across jars according to their allocation rules.
// Returns map of jar ID → allocated amount (in paise).
func (s *JarService) AllocateIncome(
	ctx context.Context,
	totalIncome int64,
) (map[uuid.UUID]int64, error) {
	s.logger.DebugContext(
		ctx,
		"allocating income",
		slog.Int64("total_income", totalIncome),
	)

	jars, err := s.repo.List(ctx)
	if err != nil {
		s.logger.ErrorContext(
			ctx,
			"failed to fetch jars for income allocation",
			slog.String("error", err.Error()),
		)
		return nil, err
	}

	if err := jar.ValidateAllocationConfiguration(jars); err != nil {
		s.logger.WarnContext(
			ctx,
			"jar configuration is invalid, cannot allocate income",
			slog.String("error", err.Error()),
		)
		return nil, err
	}

	allocations, err := jar.AllocateIncome(jars, totalIncome)
	if err != nil {
		s.logger.WarnContext(
			ctx,
			"income allocation calculation failed",
			slog.String("error", err.Error()),
		)
		return nil, err
	}

	s.logger.InfoContext(
		ctx,
		"income allocated to jars",
		slog.Int("jar_count", len(allocations)),
		slog.Int64("total_income", totalIncome),
	)

	return allocations, nil
}
