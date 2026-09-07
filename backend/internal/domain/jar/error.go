package jar

import "errors"

// Individual jar creation errors.
var (
	// ErrInvalidName indicates that the jar name is empty.
	ErrInvalidName = errors.New("jar name cannot be empty")

	// ErrInvalidAllocationType indicates that the allocation type is invalid.
	ErrInvalidAllocationType = errors.New("invalid jar allocation type")

	// ErrInvalidAllocationVal indicates that the allocation value is invalid
	// for the specified allocation type.
	ErrInvalidAllocationVal = errors.New("invalid jar allocation value")
)

// Jar state errors.
var (
	// ErrJarAlreadyArchived indicates that the jar is already archived.
	ErrJarAlreadyArchived = errors.New("jar is already archived")

	// ErrJarNotArchived indicates that the jar is not archived.
	ErrJarNotArchived = errors.New("jar is not archived")
)

// Allocation configuration errors.
var (
	// ErrEmptyAllocationConfiguration indicates that no jars are configured.
	ErrEmptyAllocationConfiguration = errors.New("at least one jar is required")

	// ErrNoRemainder indicates that the allocation configuration has no
	// remainder jar.
	ErrNoRemainder = errors.New("must have exactly 1 remainder jar")

	// ErrMultipleRemainders indicates that the allocation configuration has
	// more than one remainder jar.
	ErrMultipleRemainders = errors.New("cannot have multiple remainder jars")

	// ErrPercentageExceedsLimit indicates that the total percentage allocation
	// is greater than or equal to 100%.
	ErrPercentageExceedsLimit = errors.New(
		"percentages must sum to less than 100%",
	)

	// ErrAllocationExceedsIncome indicates that the configured fixed and
	// percentage allocations exceed the available income.
	ErrAllocationExceedsIncome = errors.New(
		"fixed and percentage allocations exceed available income",
	)
)

// Persistence errors.
var (
	// ErrJarNameExists indicates that a jar with the same name already exists.
	ErrJarNameExists = errors.New("jar name already exists")

	// ErrJarNotFound indicates that the requested jar does not exist.
	ErrJarNotFound = errors.New("jar not found")
)
