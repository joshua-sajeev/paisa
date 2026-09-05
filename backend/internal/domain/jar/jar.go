// Package jar contains the core domain model for a jar
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

	// AllocationValue is the percentage of a master income transaction
	// allocated to this jar when AllocationTypePercentage is used.
	// Valid values are 1-100.
	// It is 0 when AllocationTypeRemainder is used because remainder
	// jars do not have a fixed allocation percentage.
	AllocationValue int64
	IsArchived      bool
	CreatedAt       time.Time
	UpdatedAt       time.Time
}

// AllocationType defines how a jar's allocation is calculated.
type AllocationType string

const (
	AllocationTypePercentage AllocationType = "percentage"
	AllocationTypeRemainder  AllocationType = "remainder"
)

// IsValid reports whether the allocation type is valid.
func (at AllocationType) IsValid() bool {
	return at == AllocationTypePercentage || at == AllocationTypeRemainder
}

// NewJar creates and validates a new Jar entity.
func NewJar(name string, allocationType AllocationType, allocationValue int64) (*Jar, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		return nil, ErrInvalidName
	}

	if !allocationType.IsValid() {
		return nil, ErrInvalidAllocationType
	}

	if allocationType == AllocationTypePercentage &&
		(allocationValue < 1 || allocationValue > 100) {
		return nil, ErrInvalidAllocationVal
	}

	if allocationType == AllocationTypeRemainder &&
		allocationValue != 0 {
		return nil, ErrInvalidAllocationVal
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
