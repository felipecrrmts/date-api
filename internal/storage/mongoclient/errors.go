package mongoclient

import "errors"

var (
	ErrEmptyConfig = errors.New("conf cannot be empty")
	ErrDbNotSet    = errors.New("database must be set")
	ErrUriNotSet   = errors.New("uri must be set")
)
