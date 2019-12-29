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
	if err := tad.AddTag(tagID, false); err != nil {
		return judgeErrorResponse(err)
	}

	if err := tad.Read(); err != nil {
		return internalServerError()
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
	return c.JSON(http.StatusOK, tad)

}

func HandlePostTag(c echo.Context) error {
	tag := new(repo.Tag)
	if err := c.Bind(tag); err != nil {
		return badRequest()
	}

	tag.Official = false

	if err := tag.Create(); err != nil {
		return judgeErrorResponse(err)
	}

	return c.JSON(http.StatusOK, &tag)
}

func HandleGetTags(c echo.Context) error {
	tags, err := repo.FindTags()
	if err != nil {
		return internalServerError()
	}

	return c.JSON(http.StatusOK, tags)
}
