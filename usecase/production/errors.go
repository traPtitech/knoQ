package production

import (
	"errors"

	"github.com/traPtitech/knoQ/domain"
	"github.com/traPtitech/knoQ/infra/traq"
)

func handleTraQError(err error) error {
	if errors.Is(err, traq.ErrUnAuthorized) {
		return domain.ErrInvalidToken
	}
	return err
}
