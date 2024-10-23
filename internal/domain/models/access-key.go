package models

import (
	"github.com/google/uuid"
	"time"
)

type AccessKey struct {
	UUID      uuid.UUID
	KeyID     int
	Key       string
	ApiURL    string
	CreatedAt time.Time
	ExpiredAt time.Time
}
