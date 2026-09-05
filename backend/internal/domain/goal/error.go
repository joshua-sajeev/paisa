package goal

import "errors"

var (
	ErrInvalidName   = errors.New("goal name cannot be empty")
	ErrInvalidTarget = errors.New("goal target must be greater than zero")
)
