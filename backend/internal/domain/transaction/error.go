package transaction

import "errors"

var (
	ErrInvalidName            = errors.New("transaction name cannot be empty")
	ErrInvalidAmount          = errors.New("transaction amount must be greater than zero")
	ErrInvalidTransactionType = errors.New("invalid transaction type")
	ErrInvalidCategory        = errors.New("invalid transaction category")

	ErrTargetAccountRequired = errors.New("target account is required for income transactions")
	ErrSourceAccountRequired = errors.New("source account is required for expense transactions")
	ErrInvalidAccount        = errors.New(
		"both source and target accounts are required for transfers",
	)
	ErrInvalidTransfer = errors.New("cannot transfer to the same account")

	ErrTemplateInvalidName   = errors.New("template name cannot be empty")
	ErrTemplateInvalidAmount = errors.New("template amount must be greater than zero")
)
