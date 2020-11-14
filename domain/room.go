package domain

import (
	"time"

	"github.com/gofrs/uuid"
)

type Room struct {
	ID    uuid.UUID
	Place string
	// Verifeid indicates if the room has been verified by privileged users.
	Verified  bool
	TimeStart time.Time
	TimeEnd   time.Time
	Events    []Event
	CreatedBy User
}
