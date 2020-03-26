package repository

import (
	"errors"
)

var (
	// ErrNilID id is nil
	ErrNilID = errors.New("nil id")
	// ErrNotFound not found
	ErrNotFound = errors.New("not found")
	// ErrForbidden forbidden
	ErrForbidden = errors.New("forbidden")
	// ErrAlreadyExists already exists
	ErrAlreadyExists = errors.New("already exists")
	// ErrInvalidArg argument error
	ErrInvalidArg = errors.New("argument error")
)

// gorm.ErrRecordNotFound = ErrNotFound
