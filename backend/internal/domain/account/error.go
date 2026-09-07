package account

import (
	"errors"
)

var (
	ErrInvalidName            = errors.New("account name is invalid")
	ErrAccountNotFound        = errors.New("account not found")
	ErrAccountNameExists      = errors.New("account name already exists")
	ErrAccountAlreadyArchived = errors.New("account is already archived")
	ErrAccountNotArchived     = errors.New("account is not archived")
)
