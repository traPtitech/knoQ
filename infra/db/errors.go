package db

import (
	"errors"
	"fmt"

	"github.com/go-sql-driver/mysql"
	"gorm.io/gorm"
)

type ValueError struct {
	err  error
	args []string
}

func (ve *ValueError) Error() string {
	return fmt.Sprintf("wrong args: %s, message: %s", ve.args, ve.err)
}

func (ve *ValueError) Unwrap() error { return ve.err }

func NewValueError(err error, args ...string) error {
	return &ValueError{
		err:  err,
		args: args,
	}
}

var (
	ErrInvalidArgs     = errors.New("invalid args")
	ErrTimeConsistency = errors.New("inconsistent time")
	ErrExpression      = errors.New("invalid expression")
	ErrRoomUndefined   = errors.New("invalid room or args")
	ErrNoAdmins        = errors.New("no admins")
	ErrDuplicateEntry  = errors.New("duplicate entry")

	ErrRecordNotFound = gorm.ErrRecordNotFound
)

// defaultErrorHandling mysql等のエラーをハンドリングする
// テストと連携して、いい感じに変換する
func defaultErrorHandling(err error) error {
	var me *mysql.MySQLError
	if errors.As(err, &me) {
		switch me.Number {
		case 1032:
			return fmt.Errorf("%w: %w", ErrInvalidArgs, err)
		case 1062:
			return fmt.Errorf("%w: %w", ErrDuplicateEntry, err)
		case 1452:
			return fmt.Errorf("%w: %w", ErrRecordNotFound, err)
		}
	}
	return err
}
