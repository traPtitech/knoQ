package router

import (
	"net/http"
	"room/router/service"

	"github.com/gofrs/uuid"
	"github.com/labstack/echo/v4"
)

type tagAddDelete interface {
	Read() error
	// add unlocked tag
	AddTag(tagID uuid.UUID, locked bool) error
	// delete unlocked tag
	DeleteTag(tagID uuid.UUID) error
}

func (h *Handlers) HandlePostTag(c echo.Context) error {
	req := new(service.TagReq)
	if err := c.Bind(&req); err != nil {
		return badRequest(err)
	}

	tag, err := h.Repo.CreateOrGetTag(req.Name)
	if err != nil {
		return judgeErrorResponse(err)
	}

	return c.JSON(http.StatusOK, service.FormatTagRes(tag))
}

func (h *Handlers) HandleGetTags(c echo.Context) error {
	tags, err := h.Repo.GetAllTags()
	if err != nil {
		return judgeErrorResponse(err)
	}

	res := make([]*service.TagRes, len(tags))
	for i, tag := range tags {
		res[i] = service.FormatTagRes(tag)
	}
	return c.JSON(http.StatusOK, res)
}
