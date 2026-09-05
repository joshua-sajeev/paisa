package jar

import "errors"

var (
	ErrInvalidName           = errors.New("jar name cannot be empty")
	ErrInvalidAllocationType = errors.New("invalid jar allocation type")
	ErrInvalidAllocationVal  = errors.New("invalid jar allocation value")
)
