package router

import (
	"net/http"

	"github.com/traPtitech/knoQ/router/presentation"

	"github.com/labstack/echo/v4"
)

func (h *Handlers) HandlePostTag(c echo.Context) error {
	var req presentation.TagReq
	if err := c.Bind(&req); err != nil {
		return badRequest(err)
	}

	tag, err := h.Service.CreateOrGetTag(req.Name)
	if err != nil {
		return judgeErrorResponse(err)
	}

	return c.JSON(http.StatusOK, presentation.ConvdomainTagToTagRes(*tag))
}

func (h *Handlers) HandleGetTags(c echo.Context) error {
	tags, err := h.Service.GetAllTags()
	if err != nil {
		return judgeErrorResponse(err)
	}

	return c.JSON(http.StatusOK, presentation.ConvSPdomainTagToSPTagRes(tags))
}
