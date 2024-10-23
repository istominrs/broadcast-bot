package bot

import "errors"

var (
	ErrNoTokenProvided = errors.New("no token provided")

	ErrNoChatIDProvided = errors.New("no chat ID provided")
)
