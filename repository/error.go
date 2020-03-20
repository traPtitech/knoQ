package repository

import "errors"

var (
	// ErrNilID id is nil
	ErrNilID = errors.New("nil id")
	// ErrForbidden forbidden
	ErrForbidden = errors.New("forbidden")
)
