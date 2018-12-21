package wsStorage

import "errors"

var (
	ErrIsExisted          = errors.New("ws connection is existed")
	ErrConnectionNotFound = errors.New("ws connection is not found")
	ErrInvalidDuration = errors.New("invalid duration value")
)
