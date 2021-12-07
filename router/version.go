package router

import (
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/traPtitech/knoQ/domain"
	"github.com/traPtitech/knoQ/presentation"
)

func (h *Handlers) HandleGetVersion(c echo.Context) error {
	return c.JSON(http.StatusOK, presentation.Version{
		Version:  domain.VERSION,
		Revision: domain.REVISION,
	})
}
