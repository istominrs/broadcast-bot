package errors

import "errors"

var (
	ErrNoLoggerProvided = errors.New("no logger provided")

	ErrNoServiceProvided = errors.New("no service provided")

	ErrNoClientProvided = errors.New("no client provided")

	ErrNoRepositoryProvided = errors.New("no repository provided")
)
