package goal

import "errors"

var (
	ErrInvalidName          = errors.New("goal name cannot be empty")
	ErrInvalidTarget        = errors.New("goal target must be greater than zero")
	ErrGoalNotFound         = errors.New("goal not found")
	ErrGoalNameExists       = errors.New("goal name already exists")
	ErrInvalidAmount        = errors.New("contribution amount must be greater than zero")
	ErrContributionNotFound = errors.New("contribution not found")
	ErrGoalCannotBeDeleted  = errors.New("cannot delete in-progress goal; target not achieved")
)
