package repository

import "errors"

var (
	ErrServerNotFound = errors.New("server not found")

	AccessKeysNotFound = errors.New("access keys not found")
)
