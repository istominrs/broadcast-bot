package models

import "github.com/google/uuid"

type Server struct {
	UUID      uuid.UUID
	IpAddress string
	Port      int
	Key       string
	IsActive  bool
}
