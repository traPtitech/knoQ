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
// TODO wrapして上層に伝えたい
func defaultErrorHandling(err error) error {
	var me *mysql.MySQLError
	if errors.As(err, &me) {
		switch me.Number {
		case 1032:
			return ErrInvalidArgs
		case 1062:
			return ErrDuplicateEntry
		case 1452:
			return ErrRecordNotFound
		}
	}
	return err
}
