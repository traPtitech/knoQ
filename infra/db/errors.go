package db

import (
	"errors"
	"fmt"
)

type ValueError struct {
	err  error
	args []string
}

func (ve *ValueError) Error() string {
	return fmt.Sprintf("wrong args: %s, message: %s", ve.args, ve.err)
}

func (ve *ValueError) Unwrap() error { return ve.err }

var (
	ErrTimeConsistency = errors.New("inconsistent time")
)
