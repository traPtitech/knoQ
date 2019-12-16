package router

import (
	"net/http"
	"runtime"

	"github.com/labstack/echo/v4"
)

type tagAddDelete interface {
	Read() error
	// add unlocked tag
	AddTag(ID uint64, tagName string) error
	// delete unlocked tag
	DeleteTag(tagID uint64) error
}

func handleAddTagRelation(c echo.Context, tad tagAddDelete, ID uint64, tagName string) error {
	if err := tad.AddTag(ID, tagName); err != nil {
		return judgeErrorResponse(err, errorRuntime(newErrorRuntime(runtime.Caller(0))))
	}

	if err := tad.Read(); err != nil {
		return internalServerError()
	}
	return c.JSON(http.StatusOK, tad)
}

func handleDeleteTagRelation(c echo.Context, tad tagAddDelete, tagID uint64) error {
	if err := tad.DeleteTag(tagID); err != nil {
		return judgeErrorResponse(err)
	}
	if err := tad.Read(); err != nil {
		return internalServerError()
	}
	return c.JSON(http.StatusOK, tad)

}
