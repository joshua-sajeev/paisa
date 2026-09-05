package transaction

import (
	"time"

	"github.com/google/uuid"
)

type Allocation struct {
	ID            uuid.UUID
	TransactionID uuid.UUID
	JarID         uuid.UUID
	Amount        int64
	CreatedAt     time.Time
}
