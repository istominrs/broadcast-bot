package models

import (
	"github.com/google/uuid"
	"time"
)

type AccessKey struct {
	UUID      uuid.UUID
	Key       string
	ApiURL    string
	CreatedAt time.Time
	ExpiredAt time.Time
}
