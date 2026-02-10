package service

import (
	"errors"
	"fmt"

	"github.com/traPtitech/knoQ/domain"
	"github.com/traPtitech/knoQ/infra/db"
	"github.com/traPtitech/knoQ/infra/traq"
)

var (
	ErrInvalidArgs     = errors.New("invalid args")
	ErrTimeConsistency = errors.New("inconsistent time")
	ErrExpression      = errors.New("invalid expression")
	ErrRoomUndefined   = errors.New("invalid room or args")
	ErrNoAdmins        = errors.New("no admins")
	ErrDuplicateEntry  = errors.New("duplicate entry")
)

func handleTraQError(err error) error {
	if errors.Is(err, traq.ErrUnAuthorized) {
		return domain.ErrInvalidToken
	}
	// Forbiddenになるようなものは実装していない
	if errors.Is(err, traq.ErrForbidden) {
		return domain.ErrInvalidToken
	}
	if errors.Is(err, traq.ErrNotFound) {
		return domain.ErrNotFound
	}

	return err
}

func handleDBError(err error) error {
	if errors.Is(err, db.ErrTimeConsistency) {
		return fmt.Errorf("%w: %s", domain.ErrBadRequest, err)
	}
	if errors.Is(err, db.ErrExpression) {
		return fmt.Errorf("%w: %s", domain.ErrBadRequest, err)
	}
	if errors.Is(err, db.ErrRoomUndefined) {
		return fmt.Errorf("%w: %s", domain.ErrBadRequest, err)
	}
	if errors.Is(err, db.ErrNoAdmins) {
		return fmt.Errorf("%w: %s", domain.ErrBadRequest, err)
	}
	if errors.Is(err, db.ErrDuplicateEntry) {
		return fmt.Errorf("%w: %s", domain.ErrBadRequest, err)
	}
	if errors.Is(err, db.ErrRecordNotFound) {
		return fmt.Errorf("%w: %s", domain.ErrNotFound, err)
	}

	if errors.Is(err, db.ErrInvalidArgs) {
		return domain.ErrBadRequest
	}

	return err
}

func defaultErrorHandling(err error) error {
	err = handleTraQError(err)
	err = handleDBError(err)
	return err
}
