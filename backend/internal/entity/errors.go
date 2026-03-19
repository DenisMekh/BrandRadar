package entity

import "errors"

var (
	ErrNotFound          = errors.New("not found")
	ErrDuplicate         = errors.New("duplicate")
	ErrValidation        = errors.New("validation error")
	ErrInvalidTransition = errors.New("invalid state transition")
	ErrCooldown          = errors.New("alert cooldown active")
)
