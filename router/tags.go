package router

import (
	"net/http"
	repo "room/repository"

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

func handleAddTagRelation(c echo.Context, tad tagAddDelete, ID uuid.UUID, tagID uuid.UUID) error {
	if tagID == uuid.Nil {
		return badRequest(message("tagID is nil"))
	}
	if err := tad.AddTag(tagID, false); err != nil {
		return judgeErrorResponse(err)
	}

	if err := tad.Read(); err != nil {
		return internalServerError()
	}
	switch v := tad.(type) {
	case *repo.Event:
		res, err := formatEventRes(v)
		if err != nil {
			return internalServerError()
		}
		return c.JSON(http.StatusOK, res)

	}
	return c.JSON(http.StatusOK, tad)
}

func handleDeleteTagRelation(c echo.Context, tad tagAddDelete, tagID uuid.UUID) error {
	if err := tad.DeleteTag(tagID); err != nil {
		return judgeErrorResponse(err)
	}
	if err := tad.Read(); err != nil {
		return internalServerError()
	}

	switch v := tad.(type) {
	case *repo.Event:
		res, err := formatEventRes(v)
		if err != nil {
			return internalServerError()
		}
		return c.JSON(http.StatusOK, res)

	}

	return c.JSON(http.StatusOK, tad)
}

func (h *Handlers) HandlePostTag(c echo.Context) error {
	req := new(TagReq)
	if err := c.Bind(&req); err != nil {
		return badRequest()
	}

	tag, err := h.Repo.CreateOrGetTag(req.Name)
	if err != nil {
		return judgeErrorResponse(err)
	}

	return c.JSON(http.StatusOK, formatTagRes(tag))
}

func (h *Handlers) HandleGetTags(c echo.Context) error {
	tags, err := h.Repo.GetAllTags()
	if err != nil {
		return internalServerError()
	}

	res := make([]*TagRes, len(tags))
	for i, tag := range tags {
		res[i] = formatTagRes(tag)
	}
	return c.JSON(http.StatusOK, res)
}
