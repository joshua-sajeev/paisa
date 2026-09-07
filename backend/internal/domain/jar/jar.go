// Package jar contains the core domain model for a jar.
package jar

import (
	"strings"
	"time"

	"github.com/google/uuid"
)

type Jar struct {
	ID             uuid.UUID
	Name           string
	AllocationType AllocationType

	// AllocationValue is interpreted according to AllocationType:
	//   - percentage: percentage of income, valid values are 1-100.
	//   - fixed: fixed amount in paise, must be greater than 0.
	//   - remainder: must be 0 because the remainder is calculated dynamically.
	AllocationValue int64
	IsArchived      bool
	CreatedAt       time.Time
	UpdatedAt       time.Time
}

// JarAllocationUpdate represents a requested allocation change for a jar.
type JarAllocationUpdate struct {
	ID              uuid.UUID
	AllocationType  AllocationType
	AllocationValue int64
}

// AllocationType defines how a jar's allocation is calculated.
type AllocationType string

const (
	AllocationTypePercentage AllocationType = "percentage"
	AllocationTypeFixed      AllocationType = "fixed"
	AllocationTypeRemainder  AllocationType = "remainder"
)

// IsValid reports whether the allocation type is valid.
func (at AllocationType) IsValid() bool {
	return at == AllocationTypePercentage ||
		at == AllocationTypeFixed ||
		at == AllocationTypeRemainder
}

// NewJar creates and validates a new Jar entity.
// Individual jar validation only; it does not validate against existing jars.
func NewJar(
	name string,
	allocationType AllocationType,
	allocationValue int64,
) (*Jar, error) {
	name = strings.TrimSpace(name)

	if name == "" {
		return nil, ErrInvalidName
	}

	if err := validateAllocation(
		allocationType,
		allocationValue,
	); err != nil {
		return nil, err
	}

	now := time.Now().UTC().Truncate(time.Microsecond)

	return &Jar{
		ID:              uuid.New(),
		Name:            name,
		AllocationType:  allocationType,
		AllocationValue: allocationValue,
		IsArchived:      false,
		CreatedAt:       now,
		UpdatedAt:       now,
	}, nil
}

// Rename changes the jar name.
func (j *Jar) Rename(name string) error {
	name = strings.TrimSpace(name)

	if name == "" {
		return ErrInvalidName
	}

	j.Name = name
	j.touch()

	return nil
}

// UpdateAllocation changes the jar's allocation configuration.
func (j *Jar) UpdateAllocation(
	allocationType AllocationType,
	allocationValue int64,
) error {
	if err := validateAllocation(
		allocationType,
		allocationValue,
	); err != nil {
		return err
	}

	j.AllocationType = allocationType
	j.AllocationValue = allocationValue
	j.touch()

	return nil
}

// Archive archives the jar.
func (j *Jar) Archive() error {
	if j.IsArchived {
		return ErrJarAlreadyArchived
	}

	j.IsArchived = true
	j.touch()

	return nil
}

// Unarchive restores the jar to an active state.
func (j *Jar) Unarchive() error {
	if !j.IsArchived {
		return ErrJarNotArchived
	}

	j.IsArchived = false
	j.touch()

	return nil
}

// ValidateAllocationConfiguration checks whether a set of jars forms
// a valid allocation configuration.
//
// Rules:
//   - At least one jar is required.
//   - Exactly one remainder jar is required.
//   - Percentage and fixed allocations may be combined.
//   - Percentage allocations must sum to less than 100%.
//
// Individual jar constraints are validated by NewJar and UpdateAllocation.
func ValidateAllocationConfiguration(jars []*Jar) error {
	if len(jars) == 0 {
		return ErrEmptyAllocationConfiguration
	}

	remainderCount := 0
	var percentageSum int64

	for _, j := range jars {
		switch j.AllocationType {
		case AllocationTypeRemainder:
			remainderCount++

		case AllocationTypePercentage:
			percentageSum += j.AllocationValue

		case AllocationTypeFixed:
			// Fixed allocations are valid alongside
			// percentage allocations.
		}
	}

	if remainderCount == 0 {
		return ErrNoRemainder
	}

	if remainderCount > 1 {
		return ErrMultipleRemainders
	}

	if percentageSum >= 100 {
		return ErrPercentageExceedsLimit
	}

	return nil
}

// AllocateIncome distributes income according to the jar allocation
// configuration.
//
// Percentage allocations are calculated from total income.
// Fixed allocations are stored and calculated in paise.
// The remainder jar receives whatever income remains.
//
// Returns a map of jar ID to allocated amount in paise.
func AllocateIncome(
	jars []*Jar,
	totalIncome int64,
) (map[uuid.UUID]int64, error) {
	if err := ValidateAllocationConfiguration(jars); err != nil {
		return nil, err
	}

	if totalIncome < 0 {
		return nil, ErrAllocationExceedsIncome
	}

	allocations := make(map[uuid.UUID]int64, len(jars))

	var remainderJarID uuid.UUID
	var allocated int64

	// First pass: allocate percentage and fixed amounts.
	for _, j := range jars {
		switch j.AllocationType {
		case AllocationTypePercentage:
			amount := (j.AllocationValue * totalIncome) / 100
			allocations[j.ID] = amount
			allocated += amount

		case AllocationTypeFixed:
			allocations[j.ID] = j.AllocationValue
			allocated += j.AllocationValue

		case AllocationTypeRemainder:
			remainderJarID = j.ID
		}
	}

	// Second pass: allocate the remaining income.
	remainder := totalIncome - allocated

	if remainder < 0 {
		return nil, ErrAllocationExceedsIncome
	}

	allocations[remainderJarID] = remainder

	return allocations, nil
}

// validateAllocation validates an individual allocation configuration.
func validateAllocation(
	allocationType AllocationType,
	allocationValue int64,
) error {
	if !allocationType.IsValid() {
		return ErrInvalidAllocationType
	}

	switch allocationType {
	case AllocationTypePercentage:
		if allocationValue < 1 || allocationValue > 100 {
			return ErrInvalidAllocationVal
		}

	case AllocationTypeFixed:
		if allocationValue <= 0 {
			return ErrInvalidAllocationVal
		}

	case AllocationTypeRemainder:
		if allocationValue != 0 {
			return ErrInvalidAllocationVal
		}
	}

	return nil
}

// touch updates the jar's modification timestamp.
func (j *Jar) touch() {
	j.UpdatedAt = time.Now().UTC().Truncate(time.Microsecond)
}
